package endpoint

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/report"
)

// Reporter generates Reports containing the Endpoint topology.
type Reporter struct {
	hostID           string
	hostName         string
	includeProcesses bool
	includeNAT       bool
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
func NewReporter(hostID, hostName string, includeProcesses bool) *Reporter {
	return &Reporter{
		hostID:           hostID,
		hostName:         hostName,
		includeProcesses: includeProcesses,
		includeNAT:       conntrackModulePresent(),
	}
}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	defer func(begin time.Time) {
		SpyDuration.WithLabelValues().Observe(float64(time.Since(begin)))
	}(time.Now())

	rpt := report.MakeReport()
	conns, err := procspy.Connections(r.includeProcesses)
	if err != nil {
		return rpt, err
	}

	for conn := conns.Next(); conn != nil; conn = conns.Next() {
		r.addConnection(&rpt, conn)
	}

	if r.includeNAT {
		err = applyNAT(rpt, r.hostID)
	}

	return rpt, err
}

func (r *Reporter) addConnection(rpt *report.Report, c *procspy.Connection) {
	var (
		scopedLocal  = report.MakeAddressNodeID(r.hostID, c.LocalAddress.String())
		scopedRemote = report.MakeAddressNodeID(r.hostID, c.RemoteAddress.String())
		key          = report.MakeAdjacencyID(scopedLocal)
		edgeKey      = report.MakeEdgeID(scopedLocal, scopedRemote)
	)

	rpt.Address.Adjacency[key] = rpt.Address.Adjacency[key].Add(scopedRemote)

	if _, ok := rpt.Address.NodeMetadatas[scopedLocal]; !ok {
		rpt.Address.NodeMetadatas[scopedLocal] = report.NodeMetadata{
			"name": r.hostName,
			"addr": c.LocalAddress.String(),
		}
	}

	// Count the TCP connection.
	edgeMeta := rpt.Address.EdgeMetadatas[edgeKey]
	edgeMeta.WithConnCountTCP = true
	edgeMeta.MaxConnCountTCP++
	rpt.Address.EdgeMetadatas[edgeKey] = edgeMeta

	if c.Proc.PID > 0 {
		var (
			scopedLocal  = report.MakeEndpointNodeID(r.hostID, c.LocalAddress.String(), strconv.Itoa(int(c.LocalPort)))
			scopedRemote = report.MakeEndpointNodeID(r.hostID, c.RemoteAddress.String(), strconv.Itoa(int(c.RemotePort)))
			key          = report.MakeAdjacencyID(scopedLocal)
			edgeKey      = report.MakeEdgeID(scopedLocal, scopedRemote)
		)

		rpt.Endpoint.Adjacency[key] = rpt.Endpoint.Adjacency[key].Add(scopedRemote)

		if _, ok := rpt.Endpoint.NodeMetadatas[scopedLocal]; !ok {
			// First hit establishes NodeMetadata for scoped local address + port
			md := report.NodeMetadata{
				"addr": c.LocalAddress.String(),
				"port": strconv.Itoa(int(c.LocalPort)),
				"pid":  fmt.Sprintf("%d", c.Proc.PID),
			}

			rpt.Endpoint.NodeMetadatas[scopedLocal] = md
		}
		// Count the TCP connection.
		edgeMeta := rpt.Endpoint.EdgeMetadatas[edgeKey]
		edgeMeta.WithConnCountTCP = true
		edgeMeta.MaxConnCountTCP++
		rpt.Endpoint.EdgeMetadatas[edgeKey] = edgeMeta
	}
}
