package procconnector

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/common/fs"
)

const (
	// <linux/connector.h>
	cnIdxProc = 0x1
	cnValProc = 0x1

	// <linux/cn_proc.h>
	procCnMcastListen = 1

	procEventFork = 0x00000001 // fork() events
	procEventExec = 0x00000002 // exec() events
	procEventExit = 0x80000000 // exit() events
)

var (
	byteOrder binary.ByteOrder
)

func init() {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	if b == 0x04 {
		byteOrder = binary.LittleEndian
	} else {
		byteOrder = binary.BigEndian
	}
}

// ProcConnector receives events from the proc connector and maintain the set
// of processes.
type ProcConnector struct {
	sockfd       int
	seq          uint32
	lock         sync.RWMutex
	activePids   map[int]Process
	bufferedPids []Process
	running      bool
}

// Process represents a single process. Only include the constant details here.
type Process struct {
	Pid     int
	Name    string
	Cmdline string
}

// linux/connector.h: struct cb_id
type cbID struct {
	Idx uint32
	Val uint32
}

// linux/connector.h: struct cb_msg
type cnMsg struct {
	ID    cbID
	Seq   uint32
	Ack   uint32
	Len   uint16
	Flags uint16
}

// linux/cn_proc.h: struct proc_event.{what,cpu,timestamp_ns}
type procEventHeader struct {
	What      uint32
	CPU       uint32
	Timestamp uint64
}

// linux/cn_proc.h: struct proc_event.fork
type forkProcEvent struct {
	ParentPid  uint32
	ParentTgid uint32
	ChildPid   uint32
	ChildTgid  uint32
}

// linux/cn_proc.h: struct proc_event.exec
type execProcEvent struct {
	ProcessPid  uint32
	ProcessTgid uint32
}

// linux/cn_proc.h: struct proc_event.exit
type exitProcEvent struct {
	ProcessPid  uint32
	ProcessTgid uint32
	ExitCode    uint32
	ExitSignal  uint32
}

// standard netlink header + connector header
type netlinkProcMessage struct {
	Header syscall.NlMsghdr
	Data   cnMsg
}

// NewProcConnector creates a new process Walker.
func NewProcConnector() (pc *ProcConnector, err error) {
	pc = &ProcConnector{
		running:    false,
		activePids: map[int]Process{},
	}

	pc.sockfd, err = syscall.Socket(
		syscall.AF_NETLINK,
		syscall.SOCK_DGRAM,
		syscall.NETLINK_CONNECTOR)
	if err != nil {
		return nil, fmt.Errorf("failed to create Netlink socket: %s", err)
	}

	addr := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: cnIdxProc,
	}

	err = syscall.Bind(pc.sockfd, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to bind Netlink socket: %s", err)
	}

	err = pc.subscribe(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to the proc connector: %s", err)
	}

	// get the initial set of pids before receiving the updates
	dirEntries, err := fs.ReadDirNames("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to get initial list of processes: %s", err)
	}

	for _, filename := range dirEntries {
		pid, err := strconv.Atoi(filename)
		if err != nil {
			/* this is not an error: some files in /proc are not
			 * about processes (e.g. /proc/mounts) */
			continue
		}

		name, cmdline := GetCmdline(pid)

		pc.activePids[pid] = Process{
			Pid:     pid,
			Name:    name,
			Cmdline: cmdline,
		}
	}

	log.Infof("Proc connector successfully initialized (%d processes)", len(pc.activePids))

	go pc.receive()

	pc.running = true

	return pc, nil
}

// GetCmdline is getting the name and command line of a process based on
// /proc/$pid/cmdline
func GetCmdline(pid int) (name, cmdline string) {
	name = ""

	cmdlineBuf, err := fs.ReadFile(path.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err == nil {
		i := bytes.IndexByte(cmdlineBuf, '\000')
		if i == -1 {
			i = len(cmdlineBuf)
		}
		name = string(cmdlineBuf[:i])
		cmdlineBuf = bytes.Replace(cmdlineBuf, []byte{'\000'}, []byte{' '}, -1)
		cmdline = string(cmdlineBuf)
	}
	if name == "" {
		if commBuf, err := fs.ReadFile(path.Join("/proc", strconv.Itoa(pid), "comm")); err != nil {
			name = "[" + strings.TrimSpace(string(commBuf)) + "]"
		} else {
			name = "(unknown)"
		}
	}
	return
}

func (pc *ProcConnector) subscribe(addr *syscall.SockaddrNetlink) error {
	var op uint32
	op = procCnMcastListen
	pc.seq++

	pr := &netlinkProcMessage{}
	plen := binary.Size(pr.Data) + binary.Size(op)
	pr.Header.Len = syscall.NLMSG_HDRLEN + uint32(plen)
	pr.Header.Type = uint16(syscall.NLMSG_DONE)
	pr.Header.Flags = 0
	pr.Header.Seq = pc.seq
	pr.Header.Pid = uint32(os.Getpid())

	pr.Data.ID.Idx = cnIdxProc
	pr.Data.ID.Val = cnValProc

	pr.Data.Len = uint16(binary.Size(op))

	buf := bytes.NewBuffer(make([]byte, 0, pr.Header.Len))
	binary.Write(buf, byteOrder, pr)
	binary.Write(buf, byteOrder, op)

	err := syscall.Sendto(pc.sockfd, buf.Bytes(), 0, addr)
	return err
}

func (pc *ProcConnector) receive() {
	buf := make([]byte, syscall.Getpagesize())

	for {
		nr, _, err := syscall.Recvfrom(pc.sockfd, buf, 0)
		if err != nil {
			log.Errorf("Proc connector failed to receive a message")
			pc.running = false
			return
		}
		if nr < syscall.NLMSG_HDRLEN {
			continue
		}

		msgs, _ := syscall.ParseNetlinkMessage(buf[:nr])
		for _, m := range msgs {
			if m.Header.Type == syscall.NLMSG_DONE {
				pc.handleEvent(m.Data)
			}
		}
	}
}

func (pc *ProcConnector) handleEvent(data []byte) {
	buf := bytes.NewBuffer(data)
	msg := &cnMsg{}
	hdr := &procEventHeader{}

	binary.Read(buf, byteOrder, msg)
	binary.Read(buf, byteOrder, hdr)

	switch hdr.What {
	case procEventFork:
		event := &forkProcEvent{}
		binary.Read(buf, byteOrder, event)
		pid := int(event.ChildTgid)
		tid := int(event.ChildPid)
		if pid != tid {
			return
		}

		name, cmdline := GetCmdline(pid)

		pc.lock.Lock()
		pc.activePids[pid] = Process{
			Pid:     pid,
			Name:    name,
			Cmdline: cmdline,
		}
		pc.lock.Unlock()

	case procEventExec:
		event := &execProcEvent{}
		binary.Read(buf, byteOrder, event)
		pid := int(event.ProcessTgid)

		name, cmdline := GetCmdline(pid)

		pc.lock.Lock()
		pc.activePids[pid] = Process{
			Pid:     pid,
			Name:    name,
			Cmdline: cmdline,
		}
		pc.lock.Unlock()

	case procEventExit:
		event := &exitProcEvent{}
		binary.Read(buf, byteOrder, event)
		pid := int(event.ProcessTgid)
		tid := int(event.ProcessPid)
		if pid != tid {
			return
		}

		pc.lock.Lock()
		defer pc.lock.Unlock()

		if pr, ok := pc.activePids[pid]; ok {
			pc.bufferedPids = append(pc.bufferedPids, pr)
			delete(pc.activePids, pid)
		}

	}
}

// Walk calls f with all active processes and processes that have come and gone
// since the last call to Walk
func (pc *ProcConnector) Walk(f func(pid Process)) {
	pc.lock.RLock()
	defer pc.lock.RUnlock()

	for _, pid := range pc.activePids {
		f(pid)
	}
	for _, pid := range pc.bufferedPids {
		f(pid)
	}
	pc.bufferedPids = pc.bufferedPids[:0]
}

// IsRunning tells whether the proc connector is really working. If the kernel
// does not have CONFIG_PROC_EVENTS=y or if the socket returns some errors,
// Scope should gracefully fall back to the previous method.
func (pc *ProcConnector) IsRunning() bool {
	if pc == nil {
		return false
	}
	return pc.running
}
