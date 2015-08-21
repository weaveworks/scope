package endpoint

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Node metadata keys.
const (
	Addr = "addr" // typically IPv4
	Port = "port"
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
		localIsClient       = int(c.LocalPort) > int(c.RemotePort)
		localAddressNodeID  = report.MakeAddressNodeID(r.hostID, c.LocalAddress.String())
		remoteAddressNodeID = report.MakeAddressNodeID(r.hostID, c.RemoteAddress.String())
		adjacencyID         = ""
		edgeID              = ""
	)

	if localIsClient {
		adjacencyID = report.MakeAdjacencyID(localAddressNodeID)
		rpt.Address.Adjacency[adjacencyID] = rpt.Address.Adjacency[adjacencyID].Add(remoteAddressNodeID)

		edgeID = report.MakeEdgeID(localAddressNodeID, remoteAddressNodeID)
	} else {
		adjacencyID = report.MakeAdjacencyID(remoteAddressNodeID)
		rpt.Address.Adjacency[adjacencyID] = rpt.Address.Adjacency[adjacencyID].Add(localAddressNodeID)

		edgeID = report.MakeEdgeID(remoteAddressNodeID, localAddressNodeID)
	}

	if _, ok := rpt.Address.NodeMetadatas[localAddressNodeID]; !ok {
		rpt.Address.NodeMetadatas[localAddressNodeID] = report.MakeNodeMetadataWith(map[string]string{
			"name": r.hostName,
			Addr:   c.LocalAddress.String(),
		})
	}

	countTCPConnection(rpt.Address.EdgeMetadatas, edgeID)

	if c.Proc.PID > 0 {
		var (
			localEndpointNodeID  = report.MakeEndpointNodeID(r.hostID, c.LocalAddress.String(), strconv.Itoa(int(c.LocalPort)))
			remoteEndpointNodeID = report.MakeEndpointNodeID(r.hostID, c.RemoteAddress.String(), strconv.Itoa(int(c.RemotePort)))
			adjacencyID          = ""
			edgeID               = ""
		)

		if localIsClient {
			adjacencyID = report.MakeAdjacencyID(localEndpointNodeID)
			rpt.Endpoint.Adjacency[adjacencyID] = rpt.Endpoint.Adjacency[adjacencyID].Add(remoteEndpointNodeID)

			edgeID = report.MakeEdgeID(localEndpointNodeID, remoteEndpointNodeID)
		} else {
			adjacencyID = report.MakeAdjacencyID(remoteEndpointNodeID)
			rpt.Endpoint.Adjacency[adjacencyID] = rpt.Endpoint.Adjacency[adjacencyID].Add(localEndpointNodeID)

			edgeID = report.MakeEdgeID(remoteEndpointNodeID, localEndpointNodeID)
		}

		if _, ok := rpt.Endpoint.NodeMetadatas[localEndpointNodeID]; !ok {
			// First hit establishes NodeMetadata for scoped local address + port
			md := report.MakeNodeMetadataWith(map[string]string{
				Addr:        c.LocalAddress.String(),
				Port:        strconv.Itoa(int(c.LocalPort)),
				process.PID: fmt.Sprint(c.Proc.PID),
			})

			rpt.Endpoint.NodeMetadatas[localEndpointNodeID] = md
		}

		countTCPConnection(rpt.Endpoint.EdgeMetadatas, edgeID)
	}
}

func countTCPConnection(mds report.EdgeMetadatas, key string) {
	md := mds[key]
	if md.MaxConnCountTCP == nil {
		md.MaxConnCountTCP = new(uint64)
	}
	*md.MaxConnCountTCP++
	mds[key] = md
}
