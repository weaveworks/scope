package ptrace

import (
	"fmt"
	"log"
	"net"
	"syscall"
	"unsafe"
)

// Syscall numbers
const (
	READ          = 0
	WRITE         = 1
	CLOSE         = 3
	STAT          = 4
	MMAP          = 9
	MPROTECT      = 10
	SELECT        = 23
	MADVISE       = 28
	SOCKET        = 41
	CONNECT       = 42
	ACCEPT        = 43
	SENDTO        = 44
	RECVFROM      = 45
	CLONE         = 56
	GETID         = 186
	SETROBUSTLIST = 273
	ACCEPT4       = 288
)

// States for a given thread
const (
	NORMAL = iota
	INSYSCALL
)

type thread struct {
	tid      int
	attached bool
	process  *process // might be nil!
	tracer   *PTracer

	state      int
	callRegs   syscall.PtraceRegs
	resultRegs syscall.PtraceRegs

	currentIncoming map[int]*Fd
	currentOutgoing map[int]*Fd
}

func newThread(pid int, process *process, tracer *PTracer) *thread {
	t := &thread{
		tid:             pid,
		process:         process,
		tracer:          tracer,
		currentIncoming: map[int]*Fd{},
		currentOutgoing: map[int]*Fd{},
	}
	return t
}

// trace thread calls this
func (t *thread) syscallStopped() {
	var err error

	if t.state == NORMAL {
		if err = syscall.PtraceGetRegs(t.tid, &t.callRegs); err != nil {
			t.logf("GetRegs failed, pid=%d, err=%v", t.tid, err)
		}
		t.state = INSYSCALL
		return
	}

	t.state = NORMAL

	if err = syscall.PtraceGetRegs(t.tid, &t.resultRegs); err != nil {
		t.logf("GetRegs failed, pid=%d, err=%v", t.tid, err)
		return
	}

	if t.process == nil {
		t.logf("Got syscall, but don't know parent process!")
		return
	}

	switch t.callRegs.Orig_rax {
	case ACCEPT, ACCEPT4:
		t.handleAccept(&t.callRegs, &t.resultRegs)

	case CLOSE:
		t.handleClose(&t.callRegs, &t.resultRegs)

	case CONNECT:
		t.handleConnect(&t.callRegs, &t.resultRegs)

	case READ, WRITE, RECVFROM, SENDTO:
		t.handleIO(&t.callRegs, &t.resultRegs)

	// we can ignore these syscalls
	case SETROBUSTLIST, GETID, MMAP, MPROTECT, MADVISE, SOCKET, CLONE, STAT, SELECT:
		return

	default:
		t.logf("syscall(%d)", t.callRegs.Orig_rax)
	}
}

func (t *thread) getSocketAddress(ptr uintptr) (addr net.IP, port uint16, err error) {
	var (
		buf  = make([]byte, syscall.SizeofSockaddrAny)
		read int
	)

	if ptr == 0 {
		err = fmt.Errorf("Null ptr")
		return
	}

	read, err = syscall.PtracePeekData(t.tid, ptr, buf)
	if read != syscall.SizeofSockaddrAny || err != nil {
		return
	}

	var sockaddr4 = (*syscall.RawSockaddrInet4)(unsafe.Pointer(&buf[0]))
	if sockaddr4.Family != syscall.AF_INET {
		return
	}

	addr = net.IP(sockaddr4.Addr[0:])
	port = sockaddr4.Port
	return
}

func (t *thread) handleAccept(call, result *syscall.PtraceRegs) {
	var (
		err             error
		ok              bool
		listeningFdNum  int
		connectionFdNum int
		addrPtr         uintptr
		addr            net.IP
		port            uint16
		listeningFd     *Fd
		connection      *Fd
	)

	listeningFdNum = int(result.Rdi)
	connectionFdNum = int(result.Rax)
	addrPtr = uintptr(result.Rsi)
	addr, port, err = t.getSocketAddress(addrPtr)
	if err != nil {
		t.logf("failed to read sockaddr: %v", err)
		return
	}

	listeningFd, ok = t.process.fds[listeningFdNum]
	if !ok {
		listeningFd, err = newListeningFd(t.process.pid, listeningFdNum)
		if err != nil {
			t.logf("Failed to read listening port: %v", err)
			return
		}
		t.process.fds[listeningFdNum] = listeningFd
	}

	connection, err = listeningFd.newConnection(addr, port, connectionFdNum)
	if err != nil {
		t.logf("Failed to create connection fd: %v", err)
		return
	}

	t.process.newFd(connection)

	t.logf("Accepted connection from %s:%d -> %s:%d on fd %d, new fd %d",
		addr, port, connection.ToAddr, connection.ToPort, listeningFdNum, connectionFdNum)
}

func (t *thread) handleConnect(call, result *syscall.PtraceRegs) {
	fd := int(result.Rdi)
	ptr := result.Rsi
	addr, port, err := t.getSocketAddress(uintptr(ptr))
	if err != nil {
		t.logf("failed to read sockaddr: %v", err)
		return
	}

	connection, err := newConnectionFd(t.process.pid, fd, addr, port)
	if err != nil {
		t.logf("Failed to create connection fd: %v", err)
		return
	}

	t.process.newFd(connection)

	t.logf("Made connection from %s:%d -> %s:%d on fd %d",
		connection.ToAddr, connection.ToPort, connection.FromAddr,
		connection.FromPort, fd)
}

func (t *thread) handleClose(call, result *syscall.PtraceRegs) {
	fdNum := int(call.Rdi)
	fd, ok := t.process.fds[fdNum]
	if !ok {
		t.logf("Got close unknown fd %d", fdNum)
		return
	}

	t.logf("Closing fd %d", fdNum)
	fd.close()

	// if this connection was incoming, add it to 'the registry'
	if fd.direction == incoming {
		// collect all the outgoing connections this thread has made
		// and treat them as caused by this incoming connections
		for _, outgoing := range t.currentOutgoing {
			t.logf("Fd %d caused %d", fdNum, outgoing.fd)
			fd.Children = append(fd.Children, outgoing)
		}
		t.currentOutgoing = map[int]*Fd{}

		t.tracer.store.RecordConnection(t.process.pid, fd)
	}

	// now make sure we've remove it from everywhere
	delete(t.process.fds, fdNum)
	for _, thread := range t.process.threads {
		delete(thread.currentIncoming, fdNum)
	}
}

func (t *thread) handleIO(call, result *syscall.PtraceRegs) {
	fdNum := int(call.Rdi)
	fd, ok := t.process.fds[fdNum]
	if !ok {
		t.logf("IO on unknown fd %d", fdNum)
		return
	}

	if fd.direction == incoming {
		t.logf("IO on incoming connection %d; setting affinity", fdNum)
		t.currentIncoming[fdNum] = fd
	} else {
		t.logf("IO on outgoing connection %d; setting affinity", fdNum)
		t.currentOutgoing[fdNum] = fd
	}
}

func (t *thread) handleClone(pid int) error {
	// We can't use the pid in the process, as it may be in a namespace
	newPid, err := syscall.PtraceGetEventMsg(pid)
	if err != nil {
		log.Printf("PtraceGetEventMsg failed: %v", err)
		return err
	}

	t.logf("New thread clone'd, pid=%d", newPid)
	return nil
}

func (t *thread) handleExit() {
	t.logf("Exiting")
}

func (t *thread) logf(fmt string, args ...interface{}) {
	log.Printf("[thread %d] "+fmt, append([]interface{}{t.tid}, args...)...)
}
