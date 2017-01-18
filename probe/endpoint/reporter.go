package endpoint

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/weaveworks/scope/probe/endpoint/conntrack"
	"github.com/weaveworks/scope/probe/endpoint/procspy"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Node metadata keys.
const (
	Addr            = "addr" // typically IPv4
	Port            = "port"
	Conntracked     = "conntracked"
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
	ProcRoot     string
	BufferSize   int
	Scanner      procspy.ConnectionScanner
	DNSSnooper   *DNSSnooper
}

// Reporter generates Reports containing the Endpoint topology.
type Reporter struct {
	conf            ReporterConfig
	flowWalker      flowWalker // interface
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
}

type fourTuple struct {
	fromAddr, toAddr net.IP
	fromPort, toPort uint16
}

// key is a sortable direction-independent key for tuples, used to look up a
// fourTuple, when you are unsure of it's direction.
func (t fourTuple) key() string {
	key := []string{
		fmt.Sprintf("%s:%d", t.fromAddr, t.fromPort),
		fmt.Sprintf("%s:%d", t.toAddr, t.toPort),
	}
	sort.Strings(key)
	return strings.Join(key, " ")
}

// reverse flips the direction of the tuple
func (t *fourTuple) reverse() {
	t.fromAddr, t.fromPort, t.toAddr, t.toPort = t.toAddr, t.toPort, t.fromAddr, t.fromPort
}

// compares 2 fourTuples for equality
func (t *fourTuple) equal(x *fourTuple) bool {
	return t.fromPort == x.fromPort && t.toPort == x.fromPort && t.fromAddr.Equal(x.fromAddr) && t.toAddr.Equal(x.toAddr)
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
	{
		extraNodeInfo := map[string]string{
			Conntracked: "true",
		}
		r.flowWalker.walkFlows(func(f conntrack.Flow) {
			tuple := fourTuple{
				f.Original.Layer3.SrcIP,
				f.Original.Layer3.DstIP,
				f.Original.Layer4.SrcPort,
				f.Original.Layer4.DstPort,
			}
			// Handle DNAT-ed short-lived connections.
			// The NAT mapper won't help since it only runs periodically,
			// missing the short-lived connections.
			if !f.Original.Layer3.DstIP.Equal(f.Reply.Layer3.SrcIP) {
				tuple = fourTuple{
					f.Reply.Layer3.DstIP,
					f.Reply.Layer3.SrcIP,
					f.Reply.Layer4.DstPort,
					f.Reply.Layer4.SrcPort,
				}
			}

			seenTuples[tuple.key()] = tuple
			r.addConnection(&rpt, tuple, "", extraNodeInfo, extraNodeInfo)
		})
	}

	if r.conf.WalkProc {
		conns, err := r.conf.Scanner.Connections(r.conf.SpyProcs)
		if err != nil {
			return rpt, err
		}
		for conn := conns.Next(); conn != nil; conn = conns.Next() {
			var (
				namespaceID string
				tuple       = fourTuple{
					conn.LocalAddress,
					conn.RemoteAddress,
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
			if (ok && !canonical.equal(&tuple)) || (!ok && tuple.fromPort < tuple.toPort) {
				tuple.reverse()
				toNodeInfo, fromNodeInfo = fromNodeInfo, toNodeInfo
			}
			r.addConnection(&rpt, tuple, namespaceID, fromNodeInfo, toNodeInfo)
		}
	}

	r.natMapper.applyNAT(rpt, r.conf.HostID)
	return rpt, nil
}

func (r *Reporter) addConnection(rpt *report.Report, t fourTuple, namespaceID string, extraFromNode, extraToNode map[string]string) {
	var (
		fromNode = r.makeEndpointNode(namespaceID, t.fromAddr.String(), t.fromPort, extraFromNode)
		toNode   = r.makeEndpointNode(namespaceID, t.toAddr.String(), t.toPort, extraToNode)
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
