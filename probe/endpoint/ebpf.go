// +build linux

package endpoint

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
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
	tracer          *tracer.Tracer
	ready           bool
	stopping        bool
	dead            bool
	lastTimestampV4 uint64

	// debugBPF specifies if EbpfTracker must be started in debug mode. This
	// allows to easily debug issues like:
	// https://github.com/weaveworks/scope/issues/2650
	//
	// Scope could be started this way:
	//   $ sudo WEAVESCOPE_DOCKER_ARGS="-e SCOPE_DEBUG_BPF=1" ./scope launch
	//
	// Then, EbpfTracker could be tricked into restarting with:
	//   $ echo stop | sudo tee /proc/$(pidof scope-probe)/root/var/run/scope/debug-bpf
	debugBPF bool

	openConnections   map[fourTuple]ebpfConnection
	closedConnections map[fourTuple]ebpfConnection
	closedDuringInit  map[fourTuple]struct{}
}

// releaseRegex should match all possible variations of a common Linux
// version string:
//   - 4.1
//   - 4.22-foo
//   - 4.1.2-foo
//   - 4.1.2-33.44+bar
//   - etc.
// For example, on a Ubuntu system the vendor specific release part
// (after the first `-`) could look like:
// '<ABI number>.<upload number>-<flavour>' or
// '<ABI number>-<flavour>'
// See https://wiki.ubuntu.com/Kernel/FAQ
var releaseRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.?(\d*)-?(\d*)(.*)$`)

func isKernelSupported() error {
	release, version, err := host.GetKernelReleaseAndVersion()
	if err != nil {
		return err
	}

	releaseParts := releaseRegex.FindStringSubmatch(release)
	if len(releaseParts) != 6 {
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

	if strings.Contains(version, "Ubuntu") {
		// Check for specific Ubuntu kernel versions with
		// known issues.

		abiNumber, err := strconv.Atoi(releaseParts[4])
		if err != nil {
			// By now we know it's at least kernel 4.4 and
			// not "119-ish", so allow it.
			return nil
		}
		if major == 4 && minor == 4 && abiNumber >= 119 && abiNumber < 127 {
			// https://github.com/weaveworks/scope/issues/3131
			// https://bugs.launchpad.net/ubuntu/+source/linux/+bug/1763454
			return fmt.Errorf("got Ubuntu kernel %s with known bug", release)
		}
	}

	return nil
}

func newEbpfTracker() (*EbpfTracker, error) {
	if err := isKernelSupported(); err != nil {
		return nil, fmt.Errorf("kernel not supported: %v", err)
	}

	var debugBPF bool
	if os.Getenv("SCOPE_DEBUG_BPF") != "" {
		log.Infof("ebpf tracker started in debug mode")
		debugBPF = true
	}

	tracker := &EbpfTracker{
		debugBPF: debugBPF,
	}
	if err := tracker.restart(); err != nil {
		return nil, err
	}

	return tracker, nil
}

// TCPEventV4 handles IPv4 TCP events from the eBPF tracer
func (t *EbpfTracker) TCPEventV4(e tracer.TcpV4) {
	if t.debugBPF {
		debugBPFFile := "/var/run/scope/debug-bpf"
		b, err := ioutil.ReadFile("/var/run/scope/debug-bpf")
		if err == nil && strings.TrimSpace(string(b[:])) == "stop" {
			os.Remove(debugBPFFile)
			log.Warnf("ebpf tracker stopped as requested by user")
			t.stop()
			return
		}
	}

	if t.lastTimestampV4 > e.Timestamp {
		// A kernel bug can cause the timestamps to be wrong (e.g. on Ubuntu with Linux 4.4.0-47.68)
		// Upgrading the kernel will fix the problem. For further info see:
		// https://github.com/iovisor/bcc/issues/790#issuecomment-263704235
		// https://github.com/weaveworks/scope/issues/2334
		log.Errorf("tcp tracer received event with timestamp %v even though the last timestamp was %v. Stopping the eBPF tracker.", e.Timestamp, t.lastTimestampV4)
		t.stop()
		return
	}

	t.lastTimestampV4 = e.Timestamp

	if e.Type == tracer.EventFdInstall {
		t.handleFdInstall(e.Type, int(e.Pid), int(e.Fd))
	} else {
		tuple := fourTuple{e.SAddr.String(), e.DAddr.String(), e.SPort, e.DPort}
		t.handleConnection(e.Type, tuple, int(e.Pid), strconv.Itoa(int(e.NetNS)))
	}
}

// TCPEventV6 handles IPv6 TCP events from the eBPF tracer. This is
// currently a no-op.
func (t *EbpfTracker) TCPEventV6(e tracer.TcpV6) {
	// TODO: IPv6 not supported in Scope
}

// LostV4 handles IPv4 TCP event misses from the eBPF tracer.
func (t *EbpfTracker) LostV4(count uint64) {
	log.Errorf("tcp tracer lost %d events. Stopping the eBPF tracker", count)
	t.stop()
}

// LostV6 handles IPv4 TCP event misses from the eBPF tracer. This is
// currently a no-op.
func (t *EbpfTracker) LostV6(count uint64) {
	// TODO: IPv6 not supported in Scope
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
	if !process.IsProcInAccept("/proc", strconv.Itoa(pid)) {
		t.tracer.RemoveFdInstallWatcher(uint32(pid))
	}
	tuple, netns, ok := tupleFromPidFd(pid, fd)
	log.Debugf("EbpfTracker: got fd-install event: pid=%d fd=%d -> tuple=%s netns=%s ok=%v", pid, fd, tuple, netns, ok)
	if !ok {
		return
	}

	t.Lock()
	defer t.Unlock()

	t.openConnections[tuple] = ebpfConnection{
		incoming:         true,
		tuple:            tuple,
		pid:              pid,
		networkNamespace: netns,
	}
}

func (t *EbpfTracker) handleConnection(ev tracer.EventType, tuple fourTuple, pid int, networkNamespace string) {
	t.Lock()
	defer t.Unlock()

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
		if !t.ready {
			t.closedDuringInit[tuple] = struct{}{}
		}
		if deadConn, ok := t.openConnections[tuple]; ok {
			delete(t.openConnections, tuple)
			// mask the source port on closed connections so we only report one for each destination
			if deadConn.incoming {
				tuple.fromPort = 0
			} else {
				tuple.toPort = 0
			}
			t.closedConnections[tuple] = deadConn
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
	t.closedConnections = map[fourTuple]ebpfConnection{}
}

func (t *EbpfTracker) feedInitialConnections(conns procspy.ConnIter, seenTuples map[string]fourTuple, processesWaitingInAccept []int, hostNodeID string) {
	t.Lock()
	for conn := conns.Next(); conn != nil; conn = conns.Next() {
		tuple, namespaceID, incoming := connectionTuple(conn, seenTuples)
		if _, ok := t.closedDuringInit[tuple]; !ok {
			if _, ok := t.openConnections[tuple]; !ok {
				t.openConnections[tuple] = ebpfConnection{
					incoming:         incoming,
					tuple:            tuple,
					pid:              int(conn.Proc.PID),
					networkNamespace: namespaceID,
				}
			}
		}
	}
	t.closedDuringInit = nil
	t.ready = true
	t.Unlock()

	for _, p := range processesWaitingInAccept {
		t.tracer.AddFdInstallWatcher(uint32(p))
		log.Debugf("EbpfTracker: install fd-install watcher: pid=%d", p)
	}
}

func (t *EbpfTracker) isDead() bool {
	t.Lock()
	defer t.Unlock()
	return t.dead
}

func (t *EbpfTracker) stop() {
	t.Lock()
	alreadyDead := t.dead || t.stopping
	t.stopping = true
	t.Unlock()

	// Do not call tracer.Stop() in this thread, otherwise tracer.Stop() will
	// deadlock waiting for this thread to pick up the next event.
	go func() {
		if !alreadyDead && t.tracer != nil {
			t.tracer.Stop()
			t.tracer = nil
		}

		// Only advertise the tracer as dead after the tracer is fully stopped so that
		// restart() is not called in parallel in another thread.
		t.Lock()
		t.stopping = false
		t.dead = true
		t.Unlock()
	}()
}

func (t *EbpfTracker) restart() error {
	t.Lock()
	defer t.Unlock()

	t.dead = false
	t.ready = false

	t.openConnections = map[fourTuple]ebpfConnection{}
	t.closedDuringInit = map[fourTuple]struct{}{}
	t.closedConnections = map[fourTuple]ebpfConnection{}

	tracer, err := tracer.NewTracer(t)
	if err != nil {
		return err
	}

	t.tracer = tracer
	tracer.Start()

	return nil
}
