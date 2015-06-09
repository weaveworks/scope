package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/weaveworks/scope/report"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// DemoReport makes up a report.
func DemoReport(nodeCount int) report.Report {
	r := report.MakeReport()

	// Make up some plausible IPv4 numbers
	hosts := []string{}
	ip := [4]int{192, 168, 1, 1}
	for range make([]struct{}, nodeCount) {
		hosts = append(hosts, fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3]))
		ip[3]++
		if ip[3] > 200 {
			ip[2]++
			ip[3] = 1
		}
	}
	// Some non-local ones.
	hosts = append(hosts, []string{"1.2.3.4", "2.3.4.5"}...)

	_, localNet, err := net.ParseCIDR("192.168.0.0/16")
	if err != nil {
		panic(err)
	}

	type conn struct {
		srcProc, dstProc string
		dstPort          int
	}
	procPool := []conn{
		{srcProc: "curl", dstPort: 80, dstProc: "apache"},
		{srcProc: "wget", dstPort: 80, dstProc: "apache"},
		{srcProc: "curl", dstPort: 80, dstProc: "nginx"},
		{srcProc: "curl", dstPort: 8080, dstProc: "app1"},
		{srcProc: "nginx", dstPort: 8080, dstProc: "app1"},
		{srcProc: "nginx", dstPort: 8080, dstProc: "app2"},
		{srcProc: "nginx", dstPort: 8080, dstProc: "app3"},
	}
	connectionCount := nodeCount * 8
	for i := 0; i < connectionCount; i++ {
		var (
			c                = procPool[rand.Intn(len(procPool))]
			src              = hosts[rand.Intn(len(hosts))]
			dst              = hosts[rand.Intn(len(hosts))]
			srcPort          = rand.Intn(50000) + 10000
			srcPortID        = report.MakeEndpointNodeID("", src, strconv.Itoa(srcPort))
			dstPortID        = report.MakeEndpointNodeID("", dst, strconv.Itoa(c.dstPort))
			srcID            = report.MakeAdjacencyID("hostX", srcPortID)
			dstID            = report.MakeAdjacencyID("hostX", dstPortID)
			srcAddressID     = report.MakeAddressNodeID("", src)
			dstAddressID     = report.MakeAddressNodeID("", dst)
			nodeSrcAddressID = report.MakeAdjacencyID("hostX", srcAddressID)
			nodeDstAddressID = report.MakeAdjacencyID("hostX", dstAddressID)
		)

		// Endpoint topology
		if _, ok := r.Endpoint.NodeMetadatas[srcPortID]; !ok {
			r.Endpoint.NodeMetadatas[srcPortID] = report.NodeMetadata{
				"pid":    "4000",
				"name":   c.srcProc,
				"domain": "node-" + src,
			}
		}
		r.Endpoint.Adjacency[srcID] = r.Endpoint.Adjacency[srcID].Add(dstPortID)
		if _, ok := r.Endpoint.NodeMetadatas[dstPortID]; !ok {
			r.Endpoint.NodeMetadatas[dstPortID] = report.NodeMetadata{
				"pid":    "4000",
				"name":   c.dstProc,
				"domain": "node-" + dst,
			}
		}
		r.Endpoint.Adjacency[dstID] = r.Endpoint.Adjacency[dstID].Add(srcPortID)
		var (
			edgeKeyEgress  = report.MakeEdgeID(srcPortID, dstPortID)
			edgeKeyIngress = report.MakeEdgeID(dstPortID, srcPortID)
		)
		r.Endpoint.EdgeMetadatas[edgeKeyEgress] = report.EdgeMetadata{
			WithConnCountTCP: true,
			MaxConnCountTCP:  uint(rand.Intn(100) + 10),
		}
		r.Endpoint.EdgeMetadatas[edgeKeyIngress] = report.EdgeMetadata{
			WithConnCountTCP: true,
			MaxConnCountTCP:  uint(rand.Intn(100) + 10),
		}

		// Address topology
		if _, ok := r.Address.NodeMetadatas[srcAddressID]; !ok {
			r.Address.NodeMetadatas[srcAddressID] = report.NodeMetadata{
				"name": src,
			}
		}
		r.Address.Adjacency[nodeSrcAddressID] = r.Address.Adjacency[nodeSrcAddressID].Add(dstAddressID)
		if _, ok := r.Address.NodeMetadatas[dstAddressID]; !ok {
			r.Address.NodeMetadatas[dstAddressID] = report.NodeMetadata{
				"name": dst,
			}
		}
		r.Address.Adjacency[nodeDstAddressID] = r.Address.Adjacency[nodeDstAddressID].Add(srcAddressID)

		// Host data
		r.Host.NodeMetadatas["hostX"] = report.NodeMetadata{
			"ts":             time.Now().UTC().Format(time.RFC3339Nano),
			"host_name":      "host-x",
			"local_networks": localNet.String(),
			"os":             "linux",
		}
	}

	return r
}
