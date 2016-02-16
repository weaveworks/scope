package endpoint

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/weaveworks/scope/probe/endpoint/procspy"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Node metadata keys.
const (
	Addr        = "addr" // typically IPv4
	Port        = "port"
	Conntracked = "conntracked"
	Procspied   = "procspied"
)

// Reporter generates Reports containing the Endpoint topology.
type Reporter struct {
	hostID           string
	hostName         string
	includeProcesses bool
	includeNAT       bool
	flowWalker       flowWalker // interface
	scanner          procspy.ConnectionScanner
	natMapper        natMapper
	reverseResolver  *reverseResolver
}

// SpyDuration is an exported prometheus metric
var SpyDuration = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Namespace: "scope",
		Subsystem: "probe",
		Name:      "spy_time_nanoseconds",
		Help:      "Total time spent spying on active connections.",
		MaxAge:    10 * time.Second, // like statsd
	},
	[]string{},
)

// NewReporter creates a new Reporter that invokes procspy.Connections to
// generate a report.Report that contains every discovered (spied) connection
// on the host machine, at the granularity of host and port. That information
// is stored in the Endpoint topology. It optionally enriches that topology
// with process (PID) information.
func NewReporter(hostID, hostName string, includeProcesses bool, useConntrack bool, scanner procspy.ConnectionScanner) *Reporter {
	return &Reporter{
		hostID:           hostID,
		hostName:         hostName,
		includeProcesses: includeProcesses,
		flowWalker:       newConntrackFlowWalker(useConntrack),
		natMapper:        makeNATMapper(newConntrackFlowWalker(useConntrack, "--any-nat")),
		reverseResolver:  newReverseResolver(),
		scanner:          scanner,
	}
}

// Name of this reporter, for metrics gathering
func (Reporter) Name() string { return "Endpoint" }

// Stop stop stop
func (r *Reporter) Stop() {
	r.flowWalker.stop()
	r.natMapper.stop()
	r.reverseResolver.stop()
	r.scanner.Stop()
}

type fourTuple struct {
	fromAddr, toAddr string
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

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	defer func(begin time.Time) {
		SpyDuration.WithLabelValues().Observe(float64(time.Since(begin)))
	}(time.Now())

	hostNodeID := report.MakeHostNodeID(r.hostID)
	rpt := report.MakeReport()
	seenTuples := map[string]fourTuple{}

	// Consult the flowWalker for short-live connections
	{
		extraNodeInfo := report.MakeNode().WithLatests(map[string]string{
			Conntracked: "true",
		})
		r.flowWalker.walkFlows(func(f flow) {
			tuple := fourTuple{
				f.Original.Layer3.SrcIP,
				f.Original.Layer3.DstIP,
				uint16(f.Original.Layer4.SrcPort),
				uint16(f.Original.Layer4.DstPort),
			}
			seenTuples[tuple.key()] = tuple
			r.addConnection(&rpt, tuple, &extraNodeInfo, &extraNodeInfo)
		})
	}

	{
		conns, err := r.scanner.Connections(r.includeProcesses)
		if err != nil {
			return rpt, err
		}
		extraNodeInfo := report.MakeNode().WithLatests(map[string]string{
			Procspied: "true",
		})
		for conn := conns.Next(); conn != nil; conn = conns.Next() {
			var (
				tuple = fourTuple{
					conn.LocalAddress.String(),
					conn.RemoteAddress.String(),
					conn.LocalPort,
					conn.RemotePort,
				}
				toNodeInfo, fromNodeInfo = extraNodeInfo.Copy(), extraNodeInfo.Copy()
			)
			if conn.Proc.PID > 0 {
				fromNodeInfo = fromNodeInfo.WithLatests(map[string]string{
					process.PID:       strconv.FormatUint(uint64(conn.Proc.PID), 10),
					report.HostNodeID: hostNodeID,
				})
			}

			// If we've already seen this connection, we should know the direction
			// (or have already figured it out), so we normalize and use the
			// canonical direction. Otherwise, we can use a port-heuristic to guess
			// the direction.
			canonical, ok := seenTuples[tuple.key()]
			if (ok && canonical != tuple) || (!ok && tuple.fromPort < tuple.toPort) {
				tuple.reverse()
				toNodeInfo, fromNodeInfo = fromNodeInfo, toNodeInfo
			}
			r.addConnection(&rpt, tuple, &fromNodeInfo, &toNodeInfo)
		}
	}

	r.natMapper.applyNAT(rpt, r.hostID)
	return rpt, nil
}

func (r *Reporter) addConnection(rpt *report.Report, t fourTuple, extraFromNode, extraToNode *report.Node) {
	// Update address topology
	{
		var (
			fromAddressNodeID = report.MakeAddressNodeID(r.hostID, t.fromAddr)
			toAddressNodeID   = report.MakeAddressNodeID(r.hostID, t.toAddr)
			fromNode          = report.MakeNodeWith(map[string]string{Addr: t.fromAddr}).WithEdge(toAddressNodeID, report.EdgeMetadata{})
			toNode            = report.MakeNodeWith(map[string]string{Addr: t.toAddr})
		)

		// In case we have a reverse resolution for the IP, we can use it for
		// the name...
		if toNames, err := r.reverseResolver.get(t.toAddr); err == nil {
			toNode = toNode.WithSet("name", report.MakeStringSet(toNames...))
		}

		if extraFromNode != nil {
			fromNode = fromNode.Merge(*extraFromNode)
		}
		if extraToNode != nil {
			toNode = toNode.Merge(*extraToNode)
		}
		rpt.Address = rpt.Address.AddNode(fromAddressNodeID, fromNode)
		rpt.Address = rpt.Address.AddNode(toAddressNodeID, toNode)
	}

	// Update endpoint topology
	if r.includeProcesses {
		var (
			fromEndpointNodeID = report.MakeEndpointNodeID(r.hostID, t.fromAddr, strconv.Itoa(int(t.fromPort)))
			toEndpointNodeID   = report.MakeEndpointNodeID(r.hostID, t.toAddr, strconv.Itoa(int(t.toPort)))

			fromNode = report.MakeNodeWith(map[string]string{
				Addr: t.fromAddr,
				Port: strconv.Itoa(int(t.fromPort)),
			}).WithEdge(toEndpointNodeID, report.EdgeMetadata{})
			toNode = report.MakeNodeWith(map[string]string{
				Addr: t.toAddr,
				Port: strconv.Itoa(int(t.toPort)),
			})
		)

		// In case we have a reverse resolution for the IP, we can use it for
		// the name...
		if toNames, err := r.reverseResolver.get(t.toAddr); err == nil {
			toNode = toNode.WithSet("name", report.MakeStringSet(toNames...))
		}

		if extraFromNode != nil {
			fromNode = fromNode.Merge(*extraFromNode)
		}
		if extraToNode != nil {
			toNode = toNode.Merge(*extraToNode)
		}
		rpt.Endpoint = rpt.Endpoint.AddNode(fromEndpointNodeID, fromNode)
		rpt.Endpoint = rpt.Endpoint.AddNode(toEndpointNodeID, toNode)
	}
}

func newu64(i uint64) *uint64 {
	return &i
}
