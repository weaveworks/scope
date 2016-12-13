package endpoint

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/scope/probe/endpoint/procspy"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Node metadata keys.
const (
	Addr            = "addr" // typically IPv4
	Port            = "port"
	Conntracked     = "conntracked"
	EBPF            = "eBPF"
	Procspied       = "procspied"
	ReverseDNSNames = "reverse_dns_names"
	SnoopedDNSNames = "snooped_dns_names"
)

// ReporterConfig are the config options for the endpoint reporter.
type ReporterConfig struct {
	HostID       string
	HostName     string
	SpyProcs     bool
	UseConntrack bool
	WalkProc     bool
	UseEbpfConn  bool
	ProcRoot     string
	BufferSize   int
	Scanner      procspy.ConnectionScanner
	DNSSnooper   *DNSSnooper
}

// Reporter generates Reports containing the Endpoint topology.
type Reporter struct {
	conf            ReporterConfig
	flowWalker      flowWalker // interface
	ebpfTracker     eventTracker
	natMapper       natMapper
	reverseResolver *reverseResolver
}

// SpyDuration is an exported prometheus metric
var SpyDuration = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Namespace: "scope",
		Subsystem: "probe",
		Name:      "spy_duration_seconds",
		Help:      "Time in seconds spent spying on active connections.",
		MaxAge:    10 * time.Second, // like statsd
	},
	[]string{},
)

// NewReporter creates a new Reporter that invokes procspy.Connections to
// generate a report.Report that contains every discovered (spied) connection
// on the host machine, at the granularity of host and port. That information
// is stored in the Endpoint topology. It optionally enriches that topology
// with process (PID) information.
func NewReporter(conf ReporterConfig) *Reporter {
	return &Reporter{
		conf:            conf,
		flowWalker:      newConntrackFlowWalker(conf.UseConntrack, conf.ProcRoot, conf.BufferSize),
		ebpfTracker:     newEbpfTracker(conf.UseEbpfConn),
		natMapper:       makeNATMapper(newConntrackFlowWalker(conf.UseConntrack, conf.ProcRoot, conf.BufferSize, "--any-nat")),
		reverseResolver: newReverseResolver(),
	}
}

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "Endpoint" }

// Stop stop stop
func (r *Reporter) Stop() {
	r.flowWalker.stop()
	r.natMapper.stop()
	r.reverseResolver.stop()
	r.conf.Scanner.Stop()
	r.ebpfTracker.stop()
}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	defer func(begin time.Time) {
		SpyDuration.WithLabelValues().Observe(time.Since(begin).Seconds())
	}(time.Now())

	hostNodeID := report.MakeHostNodeID(r.conf.HostID)
	rpt := report.MakeReport()

	seenTuples := map[string]fourTuple{}

	// Consult the flowWalker for short-lived connections
	// With eBPF, this is used only in the first round to build seenTuples for WalkProc
	if r.conf.WalkProc || !r.conf.UseEbpfConn {
		extraNodeInfo := map[string]string{
			Conntracked: "true",
		}
		r.flowWalker.walkFlows(func(f flow) {
			tuple := fourTuple{
				f.Original.Layer3.SrcIP,
				f.Original.Layer3.DstIP,
				uint16(f.Original.Layer4.SrcPort),
				uint16(f.Original.Layer4.DstPort),
			}
			// Handle DNAT-ed short-lived connections.
			// The NAT mapper won't help since it only runs periodically,
			// missing the short-lived connections.
			if f.Original.Layer3.DstIP != f.Reply.Layer3.SrcIP {
				tuple = fourTuple{
					f.Reply.Layer3.DstIP,
					f.Reply.Layer3.SrcIP,
					uint16(f.Reply.Layer4.DstPort),
					uint16(f.Reply.Layer4.SrcPort),
				}
			}

			seenTuples[tuple.key()] = tuple
			r.addConnection(&rpt, tuple, "", extraNodeInfo, extraNodeInfo)
		})
	}

	if r.conf.WalkProc {
		conns, err := r.conf.Scanner.Connections(r.conf.SpyProcs)
		defer r.procParsingSwitcher()
		if err != nil {
			return rpt, err
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
			canonical, ok := seenTuples[tuple.key()]
			if (ok && canonical != tuple) || (!ok && tuple.fromPort < tuple.toPort) {
				r.feedToEbpf(tuple, true, int(conn.Proc.PID), namespaceID)
				r.addConnection(&rpt, reverse(tuple), namespaceID, toNodeInfo, fromNodeInfo)
			} else {
				r.feedToEbpf(tuple, false, int(conn.Proc.PID), namespaceID)
				r.addConnection(&rpt, tuple, namespaceID, fromNodeInfo, toNodeInfo)
			}

		}
	}

	// eBPF
	if r.conf.UseEbpfConn && !r.ebpfTracker.hasDied() {
		r.ebpfTracker.walkConnections(func(e ebpfConnection) {
			fromNodeInfo := map[string]string{
				Procspied: "true",
				EBPF:      "true",
			}
			toNodeInfo := map[string]string{
				Procspied: "true",
				EBPF:      "true",
			}
			if e.pid > 0 {
				fromNodeInfo[process.PID] = strconv.Itoa(e.pid)
				fromNodeInfo[report.HostNodeID] = hostNodeID
			}
			log.Debugf("Report: ebpfTracker %v (%v) (%v)", e.tuple, e.pid, e.incoming)

			if e.incoming {
				r.addConnection(&rpt, reverse(e.tuple), e.networkNamespace, toNodeInfo, fromNodeInfo)
			} else {
				r.addConnection(&rpt, e.tuple, e.networkNamespace, fromNodeInfo, toNodeInfo)
			}

		})
	}

	r.natMapper.applyNAT(rpt, r.conf.HostID)
	return rpt, nil
}

func (r *Reporter) addConnection(rpt *report.Report, t fourTuple, namespaceID string, extraFromNode, extraToNode map[string]string) {
	var (
		fromNode = r.makeEndpointNode(namespaceID, t.fromAddr, t.fromPort, extraFromNode)
		toNode   = r.makeEndpointNode(namespaceID, t.toAddr, t.toPort, extraToNode)
	)
	rpt.Endpoint = rpt.Endpoint.AddNode(fromNode.WithEdge(toNode.ID, report.EdgeMetadata{}))
	rpt.Endpoint = rpt.Endpoint.AddNode(toNode)
}

func (r *Reporter) makeEndpointNode(namespaceID string, addr string, port uint16, extra map[string]string) report.Node {
	portStr := strconv.Itoa(int(port))
	node := report.MakeNodeWith(
		report.MakeEndpointNodeID(r.conf.HostID, namespaceID, addr, portStr),
		map[string]string{Addr: addr, Port: portStr})
	if names := r.conf.DNSSnooper.CachedNamesForIP(addr); len(names) > 0 {
		node = node.WithSet(SnoopedDNSNames, report.MakeStringSet(names...))
	}
	if names, err := r.reverseResolver.get(addr); err == nil && len(names) > 0 {
		node = node.WithSet(ReverseDNSNames, report.MakeStringSet(names...))
	}
	if extra != nil {
		node = node.WithLatests(extra)
	}
	return node
}

func newu64(i uint64) *uint64 {
	return &i
}

// procParsingSwitcher make sure that if eBPF tracking is enabled,
// connections coming from /proc parsing are only walked once.
func (r *Reporter) procParsingSwitcher() {
	if r.conf.WalkProc && r.conf.UseEbpfConn {
		r.conf.WalkProc = false
		r.ebpfTracker.initialize()

		r.flowWalker.stop()
	}
}

// if the eBPF tracker is enabled, feed the existing connections into it
// incoming connections correspond to "accept" events
// outgoing connections correspond to "connect" events
func (r Reporter) feedToEbpf(tuple fourTuple, incoming bool, pid int, namespaceID string) {
	if r.conf.UseEbpfConn && !r.ebpfTracker.isInitialized() {
		tcpEventType := "connect"

		if incoming {
			tcpEventType = "accept"
		}

		r.ebpfTracker.handleConnection(tcpEventType, tuple, pid, namespaceID)
	}
}
