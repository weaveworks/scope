package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/weaveworks/scope/scope/report"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// DemoReport makes up a report.
func DemoReport(nodeCount int) report.Report {
	r := report.NewReport()

	// Make up some plausible IPv4 numbers
	hosts := []string{}
	ip := [4]int{192, 168, 1, 1}
	for _ = range make([]struct{}, nodeCount) {
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
	connectionCount := nodeCount * 2
	for i := 0; i < connectionCount; i++ {
		var (
			c                = procPool[rand.Intn(len(procPool))]
			src              = hosts[rand.Intn(len(hosts))]
			dst              = hosts[rand.Intn(len(hosts))]
			srcPort          = rand.Intn(50000) + 10000
			srcPortID        = fmt.Sprintf("%s%s%s%d", report.ScopeDelim, src, report.ScopeDelim, srcPort)
			dstPortID        = fmt.Sprintf("%s%s%s%d", report.ScopeDelim, dst, report.ScopeDelim, c.dstPort)
			srcID            = "hostX" + report.IDDelim + srcPortID
			dstID            = "hostX" + report.IDDelim + dstPortID
			srcAddressID     = fmt.Sprintf("%s%s", report.ScopeDelim, src)
			dstAddressID     = fmt.Sprintf("%s%s", report.ScopeDelim, dst)
			nodeSrcAddressID = "hostX" + report.IDDelim + srcAddressID
			nodeDstAddressID = "hostX" + report.IDDelim + dstAddressID
		)

		// Process topology
		if _, ok := r.Process.NodeMetadatas[srcPortID]; !ok {
			r.Process.NodeMetadatas[srcPortID] = report.NodeMetadata{
				"pid":    "4000",
				"name":   c.srcProc,
				"domain": "node-" + src,
			}
		}
		r.Process.Adjacency[srcID] = r.Process.Adjacency[srcID].Add(dstPortID)
		if _, ok := r.Process.NodeMetadatas[dstPortID]; !ok {
			r.Process.NodeMetadatas[dstPortID] = report.NodeMetadata{
				"pid":    "4000",
				"name":   c.dstProc,
				"domain": "node-" + dst,
			}
		}
		r.Process.Adjacency[dstID] = r.Process.Adjacency[dstID].Add(srcPortID)
		var (
			edgeKeyEgress  = srcPortID + report.IDDelim + dstPortID
			edgeKeyIngress = dstPortID + report.IDDelim + srcPortID
		)
		r.Process.EdgeMetadatas[edgeKeyEgress] = report.EdgeMetadata{
			WithConnCountTCP: true,
			MaxConnCountTCP:  uint(rand.Intn(100) + 10),
		}
		r.Process.EdgeMetadatas[edgeKeyIngress] = report.EdgeMetadata{
			WithConnCountTCP: true,
			MaxConnCountTCP:  uint(rand.Intn(100) + 10),
		}

		// Network topology
		if _, ok := r.Network.NodeMetadatas[srcAddressID]; !ok {
			r.Network.NodeMetadatas[srcAddressID] = report.NodeMetadata{
				"name": src,
			}
		}
		r.Network.Adjacency[nodeSrcAddressID] = r.Network.Adjacency[nodeSrcAddressID].Add(dstAddressID)
		if _, ok := r.Network.NodeMetadatas[dstAddressID]; !ok {
			r.Network.NodeMetadatas[dstAddressID] = report.NodeMetadata{
				"name": dst,
			}
		}
		r.Network.Adjacency[nodeDstAddressID] = r.Network.Adjacency[nodeDstAddressID].Add(srcAddressID)

		// Host data
		r.HostMetadatas["hostX"] = report.HostMetadata{
			Timestamp: time.Now().UTC(),
			Hostname:  "host-x",
			LocalNets: []*net.IPNet{localNet},
			OS:        "linux",
		}
	}

	return r
}
