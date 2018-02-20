package endpoint

import (
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/scope/probe/endpoint/procspy"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// connectionTrackerConfig are the config options for the endpoint tracker.
type connectionTrackerConfig struct {
	HostID       string
	HostName     string
	SpyProcs     bool
	UseConntrack bool
	WalkProc     bool
	UseEbpfConn  bool
	ProcRoot     string
	BufferSize   int
	ProcessCache *process.CachingWalker
	Scanner      procspy.ConnectionScanner
	DNSSnooper   *DNSSnooper
}

type connectionTracker struct {
	conf            connectionTrackerConfig
	flowWalker      flowWalker // Interface
	ebpfTracker     *EbpfTracker
	reverseResolver *reverseResolver

	// time of the previous ebpf failure, or zero if it didn't fail
	ebpfLastFailureTime time.Time
}

func newConnectionTracker(conf connectionTrackerConfig) connectionTracker {
	ct := connectionTracker{
		conf:            conf,
		reverseResolver: newReverseResolver(),
	}
	if conf.UseEbpfConn {
		et, err := newEbpfTracker()
		if err == nil {
			ct.ebpfTracker = et
			go ct.getInitialState()
			return ct
		}
		log.Warnf("Error setting up the eBPF tracker, falling back to proc scanning: %v", err)
	}
	ct.useProcfs()
	return ct
}

func flowToTuple(f flow) (ft fourTuple) {
	ft = fourTuple{
		f.Original.Layer3.SrcIP,
		f.Original.Layer3.DstIP,
		uint16(f.Original.Layer4.SrcPort),
		uint16(f.Original.Layer4.DstPort),
	}
	// Handle DNAT-ed connections in the initial state
	if f.Original.Layer3.DstIP != f.Reply.Layer3.SrcIP {
		ft = fourTuple{
			f.Reply.Layer3.DstIP,
			f.Reply.Layer3.SrcIP,
			uint16(f.Reply.Layer4.DstPort),
			uint16(f.Reply.Layer4.SrcPort),
		}
	}
	return ft
}

func (t *connectionTracker) useProcfs() {
	t.ebpfTracker = nil
	if t.conf.WalkProc && t.conf.Scanner == nil {
		t.conf.Scanner = procspy.NewConnectionScanner(t.conf.ProcessCache, t.conf.SpyProcs)
	}
	if t.flowWalker == nil {
		t.flowWalker = newConntrackFlowWalker(t.conf.UseConntrack, t.conf.ProcRoot, t.conf.BufferSize)
	}
}

// ReportConnections calls trackers according to the configuration.
func (t *connectionTracker) ReportConnections(rpt *report.Report) {
	hostNodeID := report.MakeHostNodeID(t.conf.HostID)

	if t.ebpfTracker != nil {
		if !t.ebpfTracker.isDead() {
			t.performEbpfTrack(rpt, hostNodeID)
			return
		}

		// We only restart the EbpfTracker if the failures are not too frequent to
		// avoid repeatitive restarts.

		ebpfLastFailureTime := t.ebpfLastFailureTime
		t.ebpfLastFailureTime = time.Now()

		if ebpfLastFailureTime.After(time.Now().Add(-5 * time.Minute)) {
			// Multiple failures in the last 5 minutes, fall back to proc parsing
			log.Warnf("ebpf tracker died again, gently falling back to proc scanning")
			t.useProcfs()
		} else {
			// Tolerable failure rate, restart the tracker
			log.Warnf("ebpf tracker died, restarting it")
			err := t.ebpfTracker.restart()
			if err == nil {
				go t.getInitialState()
				t.performEbpfTrack(rpt, hostNodeID)
				return
			}
			log.Warnf("could not restart ebpf tracker, falling back to proc scanning: %v", err)
			t.useProcfs()
		}
	}

	// consult the flowWalker for short-lived (conntracked) connections
	seenTuples := map[string]fourTuple{}
	t.flowWalker.walkFlows(func(f flow, alive bool) {
		tuple := flowToTuple(f)
		seenTuples[tuple.key()] = tuple
		t.addConnection(rpt, false, tuple, "", nil, nil)
	})

	if t.conf.WalkProc && t.conf.Scanner != nil {
		t.performWalkProc(rpt, hostNodeID, seenTuples)
	}
}

func (t *connectionTracker) existingFlows() map[string]fourTuple {
	seenTuples := map[string]fourTuple{}
	if !t.conf.UseConntrack {
		// log.Warnf("Not using conntrack: disabled")
	} else if err := IsConntrackSupported(t.conf.ProcRoot); err != nil {
		log.Warnf("Not using conntrack: not supported by the kernel: %s", err)
	} else if existingFlows, err := existingConnections([]string{"--any-nat"}); err != nil {
		log.Errorf("conntrack existingConnections error: %v", err)
	} else {
		for _, f := range existingFlows {
			tuple := flowToTuple(f)
			seenTuples[tuple.key()] = tuple
		}
	}
	return seenTuples
}

func (t *connectionTracker) performWalkProc(rpt *report.Report, hostNodeID string, seenTuples map[string]fourTuple) error {
	conns, err := t.conf.Scanner.Connections()
	if err != nil {
		return err
	}
	for conn := conns.Next(); conn != nil; conn = conns.Next() {
		tuple, namespaceID, incoming := connectionTuple(conn, seenTuples)
		var toNodeInfo, fromNodeInfo map[string]string
		if conn.Proc.PID > 0 {
			fromNodeInfo = map[string]string{
				process.PID:       strconv.FormatUint(uint64(conn.Proc.PID), 10),
				report.HostNodeID: hostNodeID,
			}
		}
		t.addConnection(rpt, incoming, tuple, namespaceID, fromNodeInfo, toNodeInfo)
	}
	return nil
}

// getInitialState runs conntrack and proc parsing synchronously only
// once to initialize ebpfTracker
func (t *connectionTracker) getInitialState() {
	var processCache *process.CachingWalker
	walker := process.NewWalker(t.conf.ProcRoot, true)
	processCache = process.NewCachingWalker(walker)
	processCache.Tick()

	scanner := procspy.NewSyncConnectionScanner(processCache, t.conf.SpyProcs)

	// Consult conntrack to get the initial state
	seenTuples := t.existingFlows()

	conns, err := scanner.Connections()
	if err != nil {
		log.Errorf("Error initializing ebpfTracker while scanning /proc, continuing without initial connections: %s", err)
	}
	scanner.Stop()

	processesWaitingInAccept := []int{}
	processCache.Walk(func(p, prev process.Process) {
		if p.IsWaitingInAccept {
			processesWaitingInAccept = append(processesWaitingInAccept, p.PID)
		}
	})

	t.ebpfTracker.feedInitialConnections(conns, seenTuples, processesWaitingInAccept, report.MakeHostNodeID(t.conf.HostID))
}

func (t *connectionTracker) performEbpfTrack(rpt *report.Report, hostNodeID string) error {
	t.ebpfTracker.walkConnections(func(e ebpfConnection) {
		var toNodeInfo, fromNodeInfo map[string]string
		if e.pid > 0 {
			fromNodeInfo = map[string]string{
				process.PID:       strconv.Itoa(e.pid),
				report.HostNodeID: hostNodeID,
			}
		}
		t.addConnection(rpt, e.incoming, e.tuple, e.networkNamespace, fromNodeInfo, toNodeInfo)
	})
	return nil
}

func (t *connectionTracker) addConnection(rpt *report.Report, incoming bool, ft fourTuple, namespaceID string, extraFromNode, extraToNode map[string]string) {
	if incoming {
		ft = reverse(ft)
		extraFromNode, extraToNode = extraToNode, extraFromNode
	}
	var (
		fromNode = t.makeEndpointNode(namespaceID, ft.fromAddr, ft.fromPort, extraFromNode)
		toNode   = t.makeEndpointNode(namespaceID, ft.toAddr, ft.toPort, extraToNode)
	)
	rpt.Endpoint.AddNode(fromNode.WithAdjacent(toNode.ID))
	rpt.Endpoint.AddNode(toNode)
	t.addDNS(rpt, ft.fromAddr)
	t.addDNS(rpt, ft.toAddr)
}

func (t *connectionTracker) makeEndpointNode(namespaceID string, addr string, port uint16, extra map[string]string) report.Node {
	portStr := strconv.Itoa(int(port))
	node := report.MakeNodeWith(report.MakeEndpointNodeID(t.conf.HostID, namespaceID, addr, portStr), nil)
	if extra != nil {
		node = node.WithLatests(extra)
	}
	return node
}

// Add DNS record for address to report, if not already there
func (t *connectionTracker) addDNS(rpt *report.Report, addr string) {
	if _, found := rpt.DNS[addr]; !found {
		forward := t.conf.DNSSnooper.CachedNamesForIP(addr)
		record := report.DNSRecord{
			Forward: report.MakeStringSet(forward...),
		}
		if names, err := t.reverseResolver.get(addr); err == nil && len(names) > 0 {
			record.Reverse = report.MakeStringSet(names...)
		}
		rpt.DNS[addr] = record
	}
}

func (t *connectionTracker) Stop() error {
	if t.ebpfTracker != nil {
		t.ebpfTracker.stop()
	}
	if t.flowWalker != nil {
		t.flowWalker.stop()
	}
	t.reverseResolver.stop()
	return nil
}

func connectionTuple(conn *procspy.Connection, seenTuples map[string]fourTuple) (fourTuple, string, bool) {
	namespaceID := ""
	tuple := fourTuple{
		conn.LocalAddress.String(),
		conn.RemoteAddress.String(),
		conn.LocalPort,
		conn.RemotePort,
	}
	if conn.Proc.NetNamespaceID > 0 {
		namespaceID = strconv.FormatUint(conn.Proc.NetNamespaceID, 10)
	}

	// If we've already seen this connection, we should know the direction
	// (or have already figured it out), so we normalize and use the
	// canonical direction. Otherwise, we can use a port-heuristic to guess
	// the direction.
	canonical, ok := seenTuples[tuple.key()]
	return tuple, namespaceID, (ok && canonical != tuple) || (!ok && tuple.fromPort < tuple.toPort)
}
