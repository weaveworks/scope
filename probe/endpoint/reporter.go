package endpoint

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/probe/tag"
	"github.com/weaveworks/scope/report"
)

type reporter struct {
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
// on the host machine, at the granularity of host and port. It optionally
// enriches that topology with process (PID) information.
func NewReporter(hostID, hostName string, includeProcesses bool) tag.Reporter {
	return &reporter{
		hostID:           hostID,
		hostName:         hostName,
		includeProcesses: includeProcesses,
		includeNAT:       conntrackModulePresent(),
	}
}

func (rep *reporter) Report() (report.Report, error) {
	defer func(begin time.Time) {
		SpyDuration.WithLabelValues().Observe(float64(time.Since(begin)))
	}(time.Now())

	r := report.MakeReport()
	conns, err := procspy.Connections(rep.includeProcesses)
	if err != nil {
		return r, err
	}

	for conn := conns.Next(); conn != nil; conn = conns.Next() {
		rep.addConnection(&r, conn)
	}

	if rep.includeNAT {
		err = applyNAT(r, rep.hostID)
	}

	return r, err
}

func (rep *reporter) addConnection(r *report.Report, c *procspy.Connection) {
	var (
		scopedLocal  = report.MakeAddressNodeID(rep.hostID, c.LocalAddress.String())
		scopedRemote = report.MakeAddressNodeID(rep.hostID, c.RemoteAddress.String())
		key          = report.MakeAdjacencyID(scopedLocal)
		edgeKey      = report.MakeEdgeID(scopedLocal, scopedRemote)
	)

	r.Address.Adjacency[key] = r.Address.Adjacency[key].Add(scopedRemote)

	if _, ok := r.Address.NodeMetadatas[scopedLocal]; !ok {
		r.Address.NodeMetadatas[scopedLocal] = report.NodeMetadata{
			"name": rep.hostName,
			"addr": c.LocalAddress.String(),
		}
	}

	// Count the TCP connection.
	edgeMeta := r.Address.EdgeMetadatas[edgeKey]
	edgeMeta.WithConnCountTCP = true
	edgeMeta.MaxConnCountTCP++
	r.Address.EdgeMetadatas[edgeKey] = edgeMeta

	if c.Proc.PID > 0 {
		var (
			scopedLocal  = report.MakeEndpointNodeID(rep.hostID, c.LocalAddress.String(), strconv.Itoa(int(c.LocalPort)))
			scopedRemote = report.MakeEndpointNodeID(rep.hostID, c.RemoteAddress.String(), strconv.Itoa(int(c.RemotePort)))
			key          = report.MakeAdjacencyID(scopedLocal)
			edgeKey      = report.MakeEdgeID(scopedLocal, scopedRemote)
		)

		r.Endpoint.Adjacency[key] = r.Endpoint.Adjacency[key].Add(scopedRemote)

		if _, ok := r.Endpoint.NodeMetadatas[scopedLocal]; !ok {
			// First hit establishes NodeMetadata for scoped local address + port
			md := report.NodeMetadata{
				"addr": c.LocalAddress.String(),
				"port": strconv.Itoa(int(c.LocalPort)),
				"pid":  fmt.Sprintf("%d", c.Proc.PID),
			}

			r.Endpoint.NodeMetadatas[scopedLocal] = md
		}
		// Count the TCP connection.
		edgeMeta := r.Endpoint.EdgeMetadatas[edgeKey]
		edgeMeta.WithConnCountTCP = true
		edgeMeta.MaxConnCountTCP++
		r.Endpoint.EdgeMetadatas[edgeKey] = edgeMeta
	}
}
