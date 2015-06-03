package main

import (
	"fmt"
	"log"
	"time"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/report"
)

// spy invokes procspy.Connections to generate a report.Report that contains
// every discovered (spied) connection on the host machine, at the granularity
// of host and port. It optionally enriches that topology with process (PID)
// information.
func spy(hostID, hostName string, includeProcesses bool) report.Report {
	defer func(begin time.Time) {
		spyDuration.WithLabelValues().Observe(float64(time.Since(begin)))
	}(time.Now())

	r := report.MakeReport()

	conns, err := procspy.Connections(includeProcesses)
	if err != nil {
		log.Printf("spy connections: %v", err)
		return r
	}
	for conn := conns.Next(); conn != nil; conn = conns.Next() {
		r = addConnection(r, conn, hostID, hostName)
	}

	return r
}

func addConnection(r report.Report, c *procspy.Connection, hostID, hostName string) report.Report {
	var (
		addressNodeID  = report.MakeAddressNodeID(hostID, c.LocalAddress.String())
		endpointNodeID = report.MakeEndpointNodeID(hostID, c.LocalAddress.String(), fmt.Sprint(c.LocalPort))
	)
	{
		var (
			remoteNodeID = report.MakeAddressNodeID(hostID, c.RemoteAddress.String())
			adjacencyID  = report.MakeAdjacencyID(hostID, addressNodeID)
			edgeID       = report.MakeEdgeID(addressNodeID, remoteNodeID)
		)
		r.Address.Adjacency[adjacencyID] = r.Address.Adjacency[adjacencyID].Add(remoteNodeID)
		if _, ok := r.Address.NodeMetadatas[addressNodeID]; !ok {
			r.Address.NodeMetadatas[addressNodeID] = report.NodeMetadata{
				"host_id":       hostID,
				"host_name":     hostName,
				"local_address": c.LocalAddress.String(),
				// Can't put remote information here, meaingfully, as the same
				// endpoint may appear in multiple connections with different
				// remotes.
			}
		}
		md := r.Address.EdgeMetadatas[edgeID]
		md.WithConnCountTCP = true
		md.MaxConnCountTCP++
		r.Address.EdgeMetadatas[edgeID] = md
	}
	{
		var (
			remoteNodeID = report.MakeEndpointNodeID(hostID, c.RemoteAddress.String(), fmt.Sprint(c.RemotePort))
			adjacencyID  = report.MakeAdjacencyID(hostID, endpointNodeID)
			edgeID       = report.MakeEdgeID(endpointNodeID, remoteNodeID)
		)
		r.Endpoint.Adjacency[adjacencyID] = r.Endpoint.Adjacency[adjacencyID].Add(remoteNodeID)
		if _, ok := r.Endpoint.NodeMetadatas[endpointNodeID]; !ok {
			r.Endpoint.NodeMetadatas[endpointNodeID] = report.NodeMetadata{
				"host_id":   hostID,
				"host_name": hostName,
				"address":   c.LocalAddress.String(),
				"port":      fmt.Sprintf("%d", c.LocalPort),
				// Can't put remote information here, meaingfully, as the same
				// endpoint may appear in multiple connections with different
				// remotes.
			}
		}
		md := r.Endpoint.EdgeMetadatas[edgeID]
		md.WithConnCountTCP = true
		md.MaxConnCountTCP++
		r.Endpoint.EdgeMetadatas[edgeID] = md
	}
	if c.Proc.PID > 0 {
		var (
			pidStr        = fmt.Sprint(c.Proc.PID)
			processNodeID = report.MakeProcessNodeID(hostID, pidStr)
		)
		if _, ok := r.Process.NodeMetadatas[processNodeID]; !ok {
			r.Process.NodeMetadatas[processNodeID] = report.NodeMetadata{
				"pid":          fmt.Sprintf("%d", c.Proc.PID),
				"process_name": c.Proc.Name,
				"host_id":      hostID,
				"host_name":    hostName,
			}
		}
		// We don't currently have enough info to build process-to-process
		// edges. But we do want to make a foreign-key-like association from
		// the endpoint to this process...
		r.Endpoint.NodeMetadatas[endpointNodeID]["process_node_id"] = processNodeID
		// That works as it's one-to-one. We could make a similar many-to-one
		// relationship from r.Address to this process, but it's more
		// complicated and less useful.
	}
	return r
}
