package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/report"
)

// spy invokes procspy.Connections to generate a report.Report that contains
// every discovered (spied) connection on the host machine, at the granularity
// of host and port. It optionally enriches that topology with process (PID)
// information.
func spy(
	hostID, hostName string,
	includeProcesses bool,
) report.Report {
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
		addConnection(&r, conn, hostID, hostName)
	}

	return r
}

func addConnection(
	r *report.Report,
	c *procspy.Connection,
	hostID, hostName string,
) {
	var (
		scopedLocal  = report.MakeAddressNodeID(hostID, c.LocalAddress.String())
		scopedRemote = report.MakeAddressNodeID(hostID, c.RemoteAddress.String())
		key          = report.MakeAdjacencyID(hostID, scopedLocal)
		edgeKey      = report.MakeEdgeID(scopedLocal, scopedRemote)
	)

	r.Network.Adjacency[key] = r.Network.Adjacency[key].Add(scopedRemote)

	if _, ok := r.Network.NodeMetadatas[scopedLocal]; !ok {
		r.Network.NodeMetadatas[scopedLocal] = report.NodeMetadata{
			"name": hostName,
		}
	}

	// Count the TCP connection.
	edgeMeta := r.Network.EdgeMetadatas[edgeKey]
	edgeMeta.WithConnCountTCP = true
	edgeMeta.MaxConnCountTCP++
	r.Network.EdgeMetadatas[edgeKey] = edgeMeta

	if c.Proc.PID > 0 {
		var (
			scopedLocal  = report.MakeEndpointNodeID(hostID, c.LocalAddress.String(), strconv.Itoa(int(c.LocalPort)))
			scopedRemote = report.MakeEndpointNodeID(hostID, c.RemoteAddress.String(), strconv.Itoa(int(c.RemotePort)))
			key          = report.MakeAdjacencyID(hostID, scopedLocal)
			edgeKey      = report.MakeEdgeID(scopedLocal, scopedRemote)
		)

		r.Endpoint.Adjacency[key] = r.Endpoint.Adjacency[key].Add(scopedRemote)

		if _, ok := r.Endpoint.NodeMetadatas[scopedLocal]; !ok {
			// First hit establishes NodeMetadata for scoped local address + port
			md := report.NodeMetadata{
				"pid":    fmt.Sprintf("%d", c.Proc.PID),
				"name":   c.Proc.Name,
				"domain": hostID,
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
