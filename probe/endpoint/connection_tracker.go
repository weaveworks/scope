// +build linux

package endpoint

import (
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/typetypetype/conntrack"

	"github.com/weaveworks/scope/probe/endpoint/procspy"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

type connectionTracker struct {
	conf            ReporterConfig
	flowWalker      flowWalker // Interface
	ebpfTracker     *EbpfTracker
	reverseResolver *reverseResolver

	// time of the previous ebpf failure, or zero if it didn't fail
	ebpfLastFailureTime time.Time
}

func newConnectionTracker(conf ReporterConfig) connectionTracker {
	ct := connectionTracker{
		conf:            conf,
		reverseResolver: newReverseResolver(),
	}
	if conf.UseEbpfConn {
		et, err := newEbpfTracker()
		if err == nil {
			ct.ebpfTracker = et
			go feedEBPFInitialState(conf, et)
			return ct
		}
		log.Warnf("Error setting up the eBPF tracker, falling back to proc scanning: %v", err)
	}
	ct.useProcfs()
	return ct
}

func flowToTuple(f conntrack.Conn) (ft fourTuple) {
	if f.Orig.Dst.Equal(f.Reply.Src) {
		return makeFourTuple(f.Orig.Src, f.Orig.Dst, uint16(f.Orig.SrcPort), uint16(f.Orig.DstPort))
	}
	// Handle DNAT-ed connections in the initial state
	return makeFourTuple(f.Orig.Dst, f.Orig.Src, uint16(f.Orig.DstPort), uint16(f.Orig.SrcPort))
}

func (t *connectionTracker) useProcfs() {
	t.ebpfTracker = nil
	if t.conf.WalkProc && t.conf.Scanner == nil {
		t.conf.Scanner = procspy.NewConnectionScanner(t.conf.ProcessCache, t.conf.SpyProcs)
	}
	if t.flowWalker == nil {
		t.flowWalker = newConntrackFlowWalker(t.conf.UseConntrack, t.conf.ProcRoot, t.conf.BufferSize, false /* natOnly */)
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

		if ebpfLastFailureTime.After(time.Now().Add(-1 * time.Minute)) {
			// Multiple failures in the last minute, fall back to proc parsing
			log.Warnf("ebpf tracker died again, gently falling back to proc scanning")
			t.useProcfs()
		} else {
			// Tolerable failure rate, restart the tracker
			log.Warnf("ebpf tracker died, restarting it")
			err := t.ebpfTracker.restart()
			if err == nil {
				feedEBPFInitialState(t.conf, t.ebpfTracker)
				t.performEbpfTrack(rpt, hostNodeID)
				return
			}
			log.Warnf("could not restart ebpf tracker, falling back to proc scanning: %v", err)
			t.useProcfs()
		}
	}

	// consult the flowWalker for short-lived (conntracked) connections
	seenTuples := map[string]fourTuple{}
	t.flowWalker.walkFlows(func(f conntrack.Conn, alive bool) {
		tuple := flowToTuple(f)
		seenTuples[tuple.key()] = tuple
		t.addConnection(rpt, false, tuple, 0, nil, nil)
	})

	if t.conf.WalkProc && t.conf.Scanner != nil {
		t.performWalkProc(rpt, hostNodeID, seenTuples)
	}
}

func existingFlowsFromConntrack(conf ReporterConfig) map[string]fourTuple {
	seenTuples := map[string]fourTuple{}
	if !conf.UseConntrack {
		// log.Warnf("Not using conntrack: disabled")
	} else if err := IsConntrackSupported(conf.ProcRoot); err != nil {
		log.Warnf("Not using conntrack: not supported by the kernel: %s", err)
	} else if existingFlows, err := conntrack.ConnectionsSize(conf.BufferSize); err != nil {
		log.Errorf("conntrack existingConnections error: %v", err)
	} else {
		for _, f := range existingFlows {
			if (f.Status & conntrack.IPS_NAT_MASK) == 0 {
				continue
			}
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

// feedEBPFInitialState runs conntrack and proc parsing synchronously only
// once to initialize ebpfTracker
// This is run on a background goroutine during initial setup, so does
// not take *connectionTracker which could change under it
func feedEBPFInitialState(conf ReporterConfig, ebpfTracker *EbpfTracker) {
	var processCache *process.CachingWalker
	walker := process.NewWalker(conf.ProcRoot, true)
	processCache = process.NewCachingWalker(walker)
	processCache.Tick()

	scanner := procspy.NewSyncConnectionScanner(processCache, conf.SpyProcs)

	// Consult conntrack to get the initial state
	seenTuples := existingFlowsFromConntrack(conf)

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

	ebpfTracker.feedInitialConnections(conns, seenTuples, processesWaitingInAccept, report.MakeHostNodeID(conf.HostID))
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

func (t *connectionTracker) addConnection(rpt *report.Report, incoming bool, ft fourTuple, namespaceID uint32, extraFromNode, extraToNode map[string]string) {
	if incoming {
		ft = reverse(ft)
		extraFromNode, extraToNode = extraToNode, extraFromNode
	}
	var (
		fromAddr = net.IP(ft.fromAddr[:])
		fromNode = t.makeEndpointNode(namespaceID, fromAddr, ft.fromPort, extraFromNode)
		toAddr   = net.IP(ft.toAddr[:])
		toNode   = t.makeEndpointNode(namespaceID, toAddr, ft.toPort, extraToNode)
	)
	rpt.Endpoint.AddNode(fromNode.WithAdjacent(toNode.ID))
	rpt.Endpoint.AddNode(toNode)
	t.addDNS(rpt, fromAddr.String())
	t.addDNS(rpt, toAddr.String())
}

func (t *connectionTracker) makeEndpointNode(namespaceID uint32, addr net.IP, port uint16, extra map[string]string) report.Node {
	node := report.MakeNodeWith(report.MakeEndpointNodeIDB(t.conf.HostID, namespaceID, addr, port), nil)
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

func connectionTuple(conn *procspy.Connection, seenTuples map[string]fourTuple) (fourTuple, uint32, bool) {
	tuple := makeFourTuple(conn.LocalAddress, conn.RemoteAddress, conn.LocalPort, conn.RemotePort)

	// If we've already seen this connection, we should know the direction
	// (or have already figured it out), so we normalize and use the
	// canonical direction. Otherwise, we can use a port-heuristic to guess
	// the direction.
	canonical, ok := seenTuples[tuple.key()]
	incoming := (ok && canonical != tuple) || (!ok && tuple.fromPort < tuple.toPort)
	return tuple, conn.Proc.NetNamespaceID, incoming
}
