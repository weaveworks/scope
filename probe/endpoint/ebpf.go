package endpoint

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/common/fs"
	"github.com/weaveworks/scope/probe/endpoint/procspy"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/tcptracer-bpf/pkg/tracer"
)

// An ebpfConnection represents a TCP connection
type ebpfConnection struct {
	tuple            fourTuple
	networkNamespace string
	incoming         bool
	pid              int
}

// EbpfTracker contains the sets of open and closed TCP connections.
// Closed connections are kept in the `closedConnections` slice for one iteration of `walkConnections`.
type EbpfTracker struct {
	sync.Mutex
	tracer                   *tracer.Tracer
	readyToHandleConnections bool
	dead                     bool
	lastTimestampV4          uint64

	openConnections   map[fourTuple]ebpfConnection
	closedConnections []ebpfConnection
}

var releaseRegex = regexp.MustCompile(`^(\d+)\.(\d+).*$`)

func isKernelSupported() error {
	release, _, err := host.GetKernelReleaseAndVersion()
	if err != nil {
		return err
	}

	releaseParts := releaseRegex.FindStringSubmatch(release)
	if len(releaseParts) != 3 {
		return fmt.Errorf("got invalid release version %q (expected format '4.4[.2-1]')", release)
	}

	major, err := strconv.Atoi(releaseParts[1])
	if err != nil {
		return err
	}

	minor, err := strconv.Atoi(releaseParts[2])
	if err != nil {
		return err
	}

	if major > 4 {
		return nil
	}

	if major < 4 || minor < 4 {
		return fmt.Errorf("got kernel %s but need kernel >=4.4", release)
	}

	return nil
}

func newEbpfTracker() (*EbpfTracker, error) {
	if err := isKernelSupported(); err != nil {
		return nil, fmt.Errorf("kernel not supported: %v", err)
	}

	tracker := &EbpfTracker{
		openConnections: map[fourTuple]ebpfConnection{},
	}

	tracer, err := tracer.NewTracer(tracker.tcpEventCbV4, tracker.tcpEventCbV6, tracker.lostCb)
	if err != nil {
		return nil, err
	}

	tracker.tracer = tracer
	tracer.Start()

	return tracker, nil
}

func (t *EbpfTracker) tcpEventCbV4(e tracer.TcpV4) {
	if t.lastTimestampV4 > e.Timestamp {
		// A kernel bug can cause the timestamps to be wrong (e.g. on Ubuntu with Linux 4.4.0-47.68)
		// Upgrading the kernel will fix the problem. For further info see:
		// https://github.com/iovisor/bcc/issues/790#issuecomment-263704235
		// https://github.com/weaveworks/scope/issues/2334
		log.Errorf("tcp tracer received event with timestamp %v even though the last timestamp was %v. Stopping the eBPF tracker.", e.Timestamp, t.lastTimestampV4)
		t.dead = true
		t.stop()
	}

	t.lastTimestampV4 = e.Timestamp

	if e.Type == tracer.EventFdInstall {
		t.handleFdInstall(e.Type, int(e.Pid), int(e.Fd))
	} else {
		tuple := fourTuple{e.SAddr.String(), e.DAddr.String(), e.SPort, e.DPort}
		t.handleConnection(e.Type, tuple, int(e.Pid), strconv.Itoa(int(e.NetNS)))
	}
}

func (t *EbpfTracker) tcpEventCbV6(e tracer.TcpV6) {
	// TODO: IPv6 not supported in Scope
}

func (t *EbpfTracker) lostCb(count uint64) {
	log.Errorf("tcp tracer lost %d events. Stopping the eBPF tracker", count)
	t.dead = true
	t.stop()
}

func tupleFromPidFd(pid int, fd int) (tuple fourTuple, netns string, ok bool) {
	// read /proc/$pid/ns/net
	//
	// probe/endpoint/procspy/proc_linux.go supports Linux < 3.8 but we
	// don't need that here since ebpf-enabled kernels will be > 3.8
	netnsIno, err := procspy.ReadNetnsFromPID(pid)
	if err != nil {
		log.Debugf("netns proc file for pid %d disappeared before we could read it: %v", pid, err)
		return fourTuple{}, "", false
	}
	netns = fmt.Sprintf("%d", netnsIno)

	// find /proc/$pid/fd/$fd's ino
	fdFilename := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
	var statFdFile syscall.Stat_t
	if err := fs.Stat(fdFilename, &statFdFile); err != nil {
		log.Debugf("proc file %q disappeared before we could read it", fdFilename)
		return fourTuple{}, "", false
	}

	if statFdFile.Mode&syscall.S_IFMT != syscall.S_IFSOCK {
		log.Errorf("file %q is not a socket", fdFilename)
		return fourTuple{}, "", false
	}
	ino := statFdFile.Ino

	// read both /proc/pid/net/{tcp,tcp6}
	buf := bytes.NewBuffer(make([]byte, 0, 5000))
	if _, err := procspy.ReadTCPFiles(pid, buf); err != nil {
		log.Debugf("TCP proc file for pid %d disappeared before we could read it: %v", pid, err)
		return fourTuple{}, "", false
	}

	// find /proc/$pid/fd/$fd's ino in /proc/pid/net/tcp
	pn := procspy.NewProcNet(buf.Bytes())
	for {
		n := pn.Next()
		if n == nil {
			log.Debugf("connection for proc file %q not found. buf=%q", fdFilename, buf.String())
			break
		}
		if n.Inode == ino {
			return fourTuple{n.LocalAddress.String(), n.RemoteAddress.String(), n.LocalPort, n.RemotePort}, netns, true
		}
	}

	return fourTuple{}, "", false
}

func (t *EbpfTracker) handleFdInstall(ev tracer.EventType, pid int, fd int) {
	tuple, netns, ok := tupleFromPidFd(pid, fd)
	log.Debugf("EbpfTracker: got fd-install event: pid=%d fd=%d -> tuple=%s netns=%s ok=%v", pid, fd, tuple, netns, ok)
	if !ok {
		return
	}
	t.openConnections[tuple] = ebpfConnection{
		incoming:         true,
		tuple:            tuple,
		pid:              pid,
		networkNamespace: netns,
	}
	if !process.IsProcInAccept("/proc", strconv.Itoa(pid)) {
		t.tracer.RemoveFdInstallWatcher(uint32(pid))
	}
}

func (t *EbpfTracker) handleConnection(ev tracer.EventType, tuple fourTuple, pid int, networkNamespace string) {
	t.Lock()
	defer t.Unlock()

	if !t.isReadyToHandleConnections() {
		return
	}

	log.Debugf("handleConnection(%v, [%v:%v --> %v:%v], pid=%v, netNS=%v)",
		ev, tuple.fromAddr, tuple.fromPort, tuple.toAddr, tuple.toPort, pid, networkNamespace)

	switch ev {
	case tracer.EventConnect:
		t.openConnections[tuple] = ebpfConnection{
			incoming:         false,
			tuple:            tuple,
			pid:              pid,
			networkNamespace: networkNamespace,
		}
	case tracer.EventAccept:
		t.openConnections[tuple] = ebpfConnection{
			incoming:         true,
			tuple:            tuple,
			pid:              pid,
			networkNamespace: networkNamespace,
		}
	case tracer.EventClose:
		if deadConn, ok := t.openConnections[tuple]; ok {
			delete(t.openConnections, tuple)
			t.closedConnections = append(t.closedConnections, deadConn)
		} else {
			log.Debugf("EbpfTracker: unmatched close event: %s pid=%d netns=%s", tuple, pid, networkNamespace)
		}
	default:
		log.Debugf("EbpfTracker: unknown event: %s (%d)", ev, ev)
	}
}

// walkConnections calls f with all open connections and connections that have come and gone
// since the last call to walkConnections
func (t *EbpfTracker) walkConnections(f func(ebpfConnection)) {
	t.Lock()
	defer t.Unlock()

	for _, connection := range t.openConnections {
		f(connection)
	}
	for _, connection := range t.closedConnections {
		f(connection)
	}
	t.closedConnections = t.closedConnections[:0]
}

func (t *EbpfTracker) feedInitialConnections(conns procspy.ConnIter, seenTuples map[string]fourTuple, processesWaitingInAccept []int, hostNodeID string) {
	t.readyToHandleConnections = true
	for conn := conns.Next(); conn != nil; conn = conns.Next() {
		tuple, namespaceID, incoming := connectionTuple(conn, seenTuples)
		if incoming {
			t.handleConnection(tracer.EventAccept, tuple, int(conn.Proc.PID), namespaceID)
		} else {
			t.handleConnection(tracer.EventConnect, tuple, int(conn.Proc.PID), namespaceID)
		}
	}
	for _, p := range processesWaitingInAccept {
		t.tracer.AddFdInstallWatcher(uint32(p))
		log.Debugf("EbpfTracker: install fd-install watcher: pid=%d", p)
	}
}

func (t *EbpfTracker) isReadyToHandleConnections() bool {
	return t.readyToHandleConnections
}

func (t *EbpfTracker) isDead() bool {
	return t.dead
}

func (t *EbpfTracker) stop() {
	if t.tracer != nil {
		t.tracer.Stop()
	}
	t.dead = true
}
