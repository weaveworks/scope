package endpoint

import (
	"log"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

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
	conntracker      Conntracker
	natmapper        *NATMapper
	revResolver      *ReverseResolver
	procReader       process.Reader
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

// NewReporter creates a new Reporter that invokes Connections to
// generate a report.Report that contains every discovered (spied) connection
// on the host machine, at the granularity of host and port. That information
// is stored in the Endpoint topology. It optionally enriches that topology
// with process (PID) information.
func NewReporter(hostID, hostName string, includeProcesses bool, procReader process.Reader, useConntrack bool) *Reporter {
	var (
		conntrackModulePresent = ConntrackModulePresent()
		conntracker            Conntracker
		natmapper              NATMapper
		err                    error
	)
	if conntrackModulePresent && useConntrack {
		conntracker, err = NewConntracker(true)
		if err != nil {
			log.Printf("Failed to start conntracker: %v", err)
		}
	}
	if conntrackModulePresent {
		ct, err := NewConntracker(true, "--any-nat")
		if err != nil {
			log.Printf("Failed to start conntracker for natmapper: %v", err)
		}
		natmapper = NewNATMapper(ct)
	}
	return &Reporter{
		hostID:           hostID,
		hostName:         hostName,
		includeProcesses: includeProcesses,
		conntracker:      conntracker,
		natmapper:        &natmapper,
		revResolver:      NewReverseResolver(),
		procReader:       procReader,
	}
}

// Stop stop stop
func (r *Reporter) Stop() {
	if r.conntracker != nil {
		r.conntracker.Stop()
	}
	if r.natmapper != nil {
		r.natmapper.Stop()
	}
	r.revResolver.Stop()
}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	defer func(begin time.Time) {
		SpyDuration.WithLabelValues().Observe(float64(time.Since(begin)))
	}(time.Now())

	hostNodeID := report.MakeHostNodeID(r.hostID)
	rpt := report.MakeReport()

	{
		commonNodeInfo := report.MakeNode().WithMetadata(report.Metadata{
			Procspied: "true",
		})

		err := r.procReader.Connections(func(conn process.Connection) {
			var (
				localPort  = conn.LocalPort
				remotePort = conn.RemotePort
				localAddr  = conn.LocalAddress.String()
				remoteAddr = conn.RemoteAddress.String()
			)
			extraNodeInfo := commonNodeInfo.Copy()
			if conn.Process.PID > 0 {
				extraNodeInfo = extraNodeInfo.WithMetadata(report.Metadata{
					process.PID:       strconv.FormatUint(uint64(conn.Process.PID), 10),
					report.HostNodeID: hostNodeID,
				})
			}
			r.addConnection(&rpt, localAddr, remoteAddr, localPort, remotePort, &extraNodeInfo, &commonNodeInfo)
		})
		if err != nil {
			return rpt, err
		}
	}

	if r.conntracker != nil {
		extraNodeInfo := report.MakeNode().WithMetadata(report.Metadata{
			Conntracked: "true",
		})
		r.conntracker.WalkFlows(func(f Flow) {
			var (
				localPort  = uint16(f.Original.Layer4.SrcPort)
				remotePort = uint16(f.Original.Layer4.DstPort)
				localAddr  = f.Original.Layer3.SrcIP
				remoteAddr = f.Original.Layer3.DstIP
			)
			r.addConnection(&rpt, localAddr, remoteAddr, localPort, remotePort, &extraNodeInfo, &extraNodeInfo)
		})
	}

	if r.natmapper != nil {
		r.natmapper.ApplyNAT(rpt, r.hostID)
	}

	return rpt, nil
}

func (r *Reporter) addConnection(rpt *report.Report, localAddr, remoteAddr string, localPort, remotePort uint16, extraLocalNode, extraRemoteNode *report.Node) {
	localIsClient := int(localPort) > int(remotePort)

	// Update address topology
	{
		var (
			localAddressNodeID  = report.MakeAddressNodeID(r.hostID, localAddr)
			remoteAddressNodeID = report.MakeAddressNodeID(r.hostID, remoteAddr)
			localNode           = report.MakeNodeWith(map[string]string{
				"name": r.hostName,
				Addr:   localAddr,
			})
			remoteNode = report.MakeNodeWith(map[string]string{
				Addr: remoteAddr,
			})
		)

		// In case we have a reverse resolution for the IP, we can use it for
		// the name...
		if revRemoteName, err := r.revResolver.Get(remoteAddr); err == nil {
			remoteNode = remoteNode.WithMetadata(map[string]string{
				"name": revRemoteName,
			})
		}

		if localIsClient {
			// New nodes are merged into the report so we don't need to do any
			// counting here; the merge does it for us.
			localNode = localNode.WithEdge(remoteAddressNodeID, report.EdgeMetadata{
				MaxConnCountTCP: newu64(1),
			})
		} else {
			remoteNode = localNode.WithEdge(localAddressNodeID, report.EdgeMetadata{
				MaxConnCountTCP: newu64(1),
			})
		}

		if extraLocalNode != nil {
			localNode = localNode.Merge(*extraLocalNode)
		}
		if extraRemoteNode != nil {
			remoteNode = remoteNode.Merge(*extraRemoteNode)
		}
		rpt.Address = rpt.Address.AddNode(localAddressNodeID, localNode)
		rpt.Address = rpt.Address.AddNode(remoteAddressNodeID, remoteNode)
	}

	// Update endpoint topology
	if r.includeProcesses {
		var (
			localEndpointNodeID  = report.MakeEndpointNodeID(r.hostID, localAddr, strconv.Itoa(int(localPort)))
			remoteEndpointNodeID = report.MakeEndpointNodeID(r.hostID, remoteAddr, strconv.Itoa(int(remotePort)))

			localNode = report.MakeNodeWith(map[string]string{
				Addr: localAddr,
				Port: strconv.Itoa(int(localPort)),
			})
			remoteNode = report.MakeNodeWith(map[string]string{
				Addr: remoteAddr,
				Port: strconv.Itoa(int(remotePort)),
			})
		)

		// In case we have a reverse resolution for the IP, we can use it for
		// the name...
		if revRemoteName, err := r.revResolver.Get(remoteAddr); err == nil {
			remoteNode = remoteNode.WithMetadata(map[string]string{
				"name": revRemoteName,
			})
		}

		if localIsClient {
			// New nodes are merged into the report so we don't need to do any
			// counting here; the merge does it for us.
			localNode = localNode.WithEdge(remoteEndpointNodeID, report.EdgeMetadata{
				MaxConnCountTCP: newu64(1),
			})
		} else {
			remoteNode = remoteNode.WithEdge(localEndpointNodeID, report.EdgeMetadata{
				MaxConnCountTCP: newu64(1),
			})
		}

		if extraLocalNode != nil {
			localNode = localNode.Merge(*extraLocalNode)
		}
		if extraRemoteNode != nil {
			remoteNode = remoteNode.Merge(*extraRemoteNode)
		}
		rpt.Endpoint = rpt.Endpoint.AddNode(localEndpointNodeID, localNode)
		rpt.Endpoint = rpt.Endpoint.AddNode(remoteEndpointNodeID, remoteNode)
	}
}

func newu64(i uint64) *uint64 {
	return &i
}
