package endpoint

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/scope/probe/endpoint/procspy"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/tcptracer-bpf/pkg/tracer"
)

// An ebpfConnection represents a TCP connection
type ebpfConnection struct {
	tuple            fourTuple
	networkNamespace string
	incoming         bool
	pid              int
}

type eventTracker interface {
	handleConnection(ev tracer.EventType, tuple fourTuple, pid int, networkNamespace string)
	walkConnections(f func(ebpfConnection))
	feedInitialConnections(ci procspy.ConnIter, seenTuples map[string]fourTuple, hostNodeID string)
	isReadyToHandleConnections() bool
	stop()
}

var ebpfTracker *EbpfTracker

// EbpfTracker contains the sets of open and closed TCP connections.
// Closed connections are kept in the `closedConnections` slice for one iteration of `walkConnections`.
type EbpfTracker struct {
	sync.Mutex
	tracer                   *tracer.Tracer
	readyToHandleConnections bool
	dead                     bool

	openConnections   map[string]ebpfConnection
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

func newEbpfTracker(useEbpfConn bool) (eventTracker, error) {
	if !useEbpfConn {
		return nil, errors.New("ebpf tracker not enabled")
	}

	if err := isKernelSupported(); err != nil {
		return nil, fmt.Errorf("kernel not supported: %v", err)
	}

	t, err := tracer.NewTracer(tcpEventCbV4, tcpEventCbV6)
	if err != nil {
		return nil, err
	}

	tracker := &EbpfTracker{
		openConnections: map[string]ebpfConnection{},
		tracer:          t,
	}

	ebpfTracker = tracker
	return tracker, nil
}

var lastTimestampV4 uint64

func tcpEventCbV4(e tracer.TcpV4) {
	if lastTimestampV4 > e.Timestamp {
		log.Errorf("ERROR: late event!\n")
	}

	lastTimestampV4 = e.Timestamp

	tuple := fourTuple{e.SAddr.String(), e.DAddr.String(), e.SPort, e.DPort}
	ebpfTracker.handleConnection(e.Type, tuple, int(e.Pid), strconv.Itoa(int(e.NetNS)))
}

func tcpEventCbV6(e tracer.TcpV6) {
	// TODO: IPv6 not supported in Scope
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
		conn := ebpfConnection{
			incoming:         false,
			tuple:            tuple,
			pid:              pid,
			networkNamespace: networkNamespace,
		}
		t.openConnections[tuple.String()] = conn
	case tracer.EventAccept:
		conn := ebpfConnection{
			incoming:         true,
			tuple:            tuple,
			pid:              pid,
			networkNamespace: networkNamespace,
		}
		t.openConnections[tuple.String()] = conn
	case tracer.EventClose:
		if deadConn, ok := t.openConnections[tuple.String()]; ok {
			delete(t.openConnections, tuple.String())
			t.closedConnections = append(t.closedConnections, deadConn)
		} else {
			log.Debugf("EbpfTracker: unmatched close event: %s pid=%d netns=%s", tuple.String(), pid, networkNamespace)
		}
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

func (t *EbpfTracker) feedInitialConnections(conns procspy.ConnIter, seenTuples map[string]fourTuple, hostNodeID string) {
	t.readyToHandleConnections = true
	for conn := conns.Next(); conn != nil; conn = conns.Next() {
		var (
			namespaceID string
			tuple       = fourTuple{
				conn.LocalAddress.String(),
				conn.RemoteAddress.String(),
				conn.LocalPort,
				conn.RemotePort,
			}
		)

		if conn.Proc.NetNamespaceID > 0 {
			namespaceID = strconv.FormatUint(conn.Proc.NetNamespaceID, 10)
		}

		// We can use a port-heuristic to guess the direction.
		// We assume that tuple.fromPort < tuple.toPort is a connect event (outgoing)
		canonical, ok := seenTuples[tuple.key()]
		if (ok && canonical != tuple) || (!ok && tuple.fromPort < tuple.toPort) {
			t.handleConnection(tracer.EventConnect, tuple, int(conn.Proc.PID), namespaceID)
		} else {
			t.handleConnection(tracer.EventAccept, tuple, int(conn.Proc.PID), namespaceID)
		}
	}
}

func (t *EbpfTracker) isReadyToHandleConnections() bool {
	return t.readyToHandleConnections
}

func (t *EbpfTracker) stop() {
	// TODO: implement proper stopping logic
	//
	// Even if we stop the go routine, it's not enough since we disabled the
	// async proc parser. We leave this uninmplemented for now because:
	//
	// * Ebpf parsing is optional (need to be enabled explicitly with
	// --probe.ebpf.connections=true), if a user enables it we assume they
	// check on the logs whether it works or not
	//
	// * It's unlikely that the ebpf tracker stops working if it started
	// successfully and if it does, we probaby want it to fail hard
}
