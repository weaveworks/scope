package endpoint

import (
	"strconv"

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
	ebpfTracker     eventTracker
	reverseResolver *reverseResolver
}

func newProcfsConnectionTracker(conf connectionTrackerConfig) connectionTracker {
	if conf.WalkProc && conf.Scanner == nil {
		conf.Scanner = procspy.NewConnectionScanner(conf.ProcessCache)
	}
	return connectionTracker{
		conf:            conf,
		flowWalker:      newConntrackFlowWalker(conf.UseConntrack, conf.ProcRoot, conf.BufferSize, "--any-nat"),
		ebpfTracker:     nil,
		reverseResolver: newReverseResolver(),
	}
}

func newConnectionTracker(conf connectionTrackerConfig) connectionTracker {
	if !conf.UseEbpfConn {
		// ebpf off, use proc scanning for connection tracking
		return newProcfsConnectionTracker(conf)
	}
	et, err := newEbpfTracker()
	if err != nil {
		// ebpf failed, fallback to proc scanning for connection tracking
		log.Warnf("Error setting up the eBPF tracker, falling back to proc scanning: %v", err)
		return newProcfsConnectionTracker(conf)
	}
	ct := connectionTracker{
		conf:            conf,
		flowWalker:      nil,
		ebpfTracker:     et,
		reverseResolver: newReverseResolver(),
	}
	go ct.getInitialState()
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

// ReportConnections calls trackers according to the configuration.
func (t *connectionTracker) ReportConnections(rpt *report.Report) {
	hostNodeID := report.MakeHostNodeID(t.conf.HostID)

	if t.ebpfTracker != nil {
		if !t.ebpfTracker.isDead() {
			t.performEbpfTrack(rpt, hostNodeID)
			return
		}
		log.Warnf("ebpf tracker died, gently falling back to proc scanning")
		if t.conf.WalkProc && t.conf.Scanner == nil {
			t.conf.Scanner = procspy.NewConnectionScanner(t.conf.ProcessCache)
		}
		if t.flowWalker == nil {
			t.flowWalker = newConntrackFlowWalker(t.conf.UseConntrack, t.conf.ProcRoot, t.conf.BufferSize, "--any-nat")
		}
		t.ebpfTracker = nil
	}

	// seenTuples contains information about connections seen by conntrack and it will be passed to the /proc parser
	seenTuples := map[string]fourTuple{}
	if t.flowWalker != nil {
		t.performFlowWalk(rpt, &seenTuples)
	}
	// if eBPF was enabled but failed to initialize, Scanner will be nil.
	// We can't recover from this, so don't walk proc in that case.
	// TODO: implement fallback
	if t.conf.WalkProc && t.conf.Scanner != nil {
		t.performWalkProc(rpt, hostNodeID, &seenTuples)
	}
}

func (t *connectionTracker) performFlowWalk(rpt *report.Report, seenTuples *map[string]fourTuple) {
	// Consult the flowWalker for short-lived connections
	extraNodeInfo := map[string]string{
		Conntracked: "true",
	}
	t.flowWalker.walkFlows(func(f flow, alive bool) {
		tuple := flowToTuple(f)
		(*seenTuples)[tuple.key()] = tuple
		t.addConnection(rpt, tuple, "", extraNodeInfo, extraNodeInfo)
	})
}

func (t *connectionTracker) performWalkProc(rpt *report.Report, hostNodeID string, seenTuples *map[string]fourTuple) error {
	conns, err := t.conf.Scanner.Connections(t.conf.SpyProcs)
	if err != nil {
		return err
	}
	for conn := conns.Next(); conn != nil; conn = conns.Next() {
		var (
			namespaceID string
			tuple       = fourTuple{
				conn.LocalAddress.String(),
				conn.RemoteAddress.String(),
				conn.LocalPort,
				conn.RemotePort,
			}
			toNodeInfo   = map[string]string{Procspied: "true"}
			fromNodeInfo = map[string]string{Procspied: "true"}
		)
		if conn.Proc.PID > 0 {
			fromNodeInfo[process.PID] = strconv.FormatUint(uint64(conn.Proc.PID), 10)
			fromNodeInfo[report.HostNodeID] = hostNodeID
		}

		if conn.Proc.NetNamespaceID > 0 {
			namespaceID = strconv.FormatUint(conn.Proc.NetNamespaceID, 10)
		}

		// If we've already seen this connection, we should know the direction
		// (or have already figured it out), so we normalize and use the
		// canonical direction. Otherwise, we can use a port-heuristic to guess
		// the direction.
		canonical, ok := (*seenTuples)[tuple.key()]
		if (ok && canonical != tuple) || (!ok && tuple.fromPort < tuple.toPort) {
			tuple.reverse()
			toNodeInfo, fromNodeInfo = fromNodeInfo, toNodeInfo
		}
		t.addConnection(rpt, tuple, namespaceID, fromNodeInfo, toNodeInfo)
	}
	return nil
}

// getInitialState runs conntrack and proc parsing synchronously only
// once to initialize ebpfTracker
func (t *connectionTracker) getInitialState() {
	var processCache *process.CachingWalker
	processCache = process.NewCachingWalker(process.NewWalker(t.conf.ProcRoot))
	processCache.Tick()

	scanner := procspy.NewSyncConnectionScanner(processCache)
	seenTuples := map[string]fourTuple{}
	// Consult the flowWalker to get the initial state
	if err := IsConntrackSupported(t.conf.ProcRoot); t.conf.UseConntrack && err != nil {
		log.Warnf("Not using conntrack: not supported by the kernel: %s", err)
	} else if existingFlows, err := existingConnections([]string{"--any-nat"}); err != nil {
		log.Errorf("conntrack existingConnections error: %v", err)
	} else {
		for _, f := range existingFlows {
			tuple := flowToTuple(f)
			seenTuples[tuple.key()] = tuple
		}
	}

	conns, err := scanner.Connections(t.conf.SpyProcs)
	if err != nil {
		log.Errorf("Error initializing ebpfTracker while scanning /proc, continuing without initial connections: %s", err)
	}
	scanner.Stop()

	t.ebpfTracker.feedInitialConnections(conns, seenTuples, report.MakeHostNodeID(t.conf.HostID))
}

func (t *connectionTracker) performEbpfTrack(rpt *report.Report, hostNodeID string) error {
	t.ebpfTracker.walkConnections(func(e ebpfConnection) {
		fromNodeInfo := map[string]string{
			EBPF: "true",
		}
		toNodeInfo := map[string]string{
			EBPF: "true",
		}
		if e.pid > 0 {
			fromNodeInfo[process.PID] = strconv.Itoa(e.pid)
			fromNodeInfo[report.HostNodeID] = hostNodeID
		}

		if e.incoming {
			t.addConnection(rpt, reverse(e.tuple), e.networkNamespace, toNodeInfo, fromNodeInfo)
		} else {
			t.addConnection(rpt, e.tuple, e.networkNamespace, fromNodeInfo, toNodeInfo)
		}

	})
	return nil
}

func (t *connectionTracker) addConnection(rpt *report.Report, ft fourTuple, namespaceID string, extraFromNode, extraToNode map[string]string) {
	var (
		fromNode = t.makeEndpointNode(namespaceID, ft.fromAddr, ft.fromPort, extraFromNode)
		toNode   = t.makeEndpointNode(namespaceID, ft.toAddr, ft.toPort, extraToNode)
	)
	rpt.Endpoint = rpt.Endpoint.AddNode(fromNode.WithEdge(toNode.ID, report.EdgeMetadata{}))
	rpt.Endpoint = rpt.Endpoint.AddNode(toNode)
}

func (t *connectionTracker) makeEndpointNode(namespaceID string, addr string, port uint16, extra map[string]string) report.Node {
	portStr := strconv.Itoa(int(port))
	node := report.MakeNodeWith(
		report.MakeEndpointNodeID(t.conf.HostID, namespaceID, addr, portStr),
		map[string]string{Addr: addr, Port: portStr})
	if names := t.conf.DNSSnooper.CachedNamesForIP(addr); len(names) > 0 {
		node = node.WithSet(SnoopedDNSNames, report.MakeStringSet(names...))
	}
	if names, err := t.reverseResolver.get(addr); err == nil && len(names) > 0 {
		node = node.WithSet(ReverseDNSNames, report.MakeStringSet(names...))
	}
	if extra != nil {
		node = node.WithLatests(extra)
	}
	return node
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
