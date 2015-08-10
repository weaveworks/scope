package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

func main() {
	var (
		publish         = flag.String("publish", fmt.Sprintf("localhost:%d", xfer.AppPort), "publish target")
		publishInterval = flag.Duration("publish.interval", 1*time.Second, "publish (output) interval")
		hostCount       = flag.Int("hostcount", 10, "Number of demo hosts to generate")
	)
	flag.Parse()

	publisher, err := xfer.NewHTTPPublisher(*publish, "demoprobe")
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())
	for range time.Tick(*publishInterval) {
		if err := publisher.Publish(demoReport(*hostCount)); err != nil {
			log.Print(err)
		}
	}
}

func demoReport(nodeCount int) report.Report {
	r := report.MakeReport()

	// Make up some plausible IPv4 numbers.
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
	hosts = append(hosts, []string{"1.2.3.4", "2.3.4.5"}...) // Some non-local ones, too.

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
			srcPortID        = report.MakeEndpointNodeID("", src, strconv.Itoa(srcPort))
			dstPortID        = report.MakeEndpointNodeID("", dst, strconv.Itoa(c.dstPort))
			srcID            = report.MakeAdjacencyID(srcPortID)
			dstID            = report.MakeAdjacencyID(dstPortID)
			srcAddressID     = report.MakeAddressNodeID("", src)
			dstAddressID     = report.MakeAddressNodeID("", dst)
			nodeSrcAddressID = report.MakeAdjacencyID(srcAddressID)
			nodeDstAddressID = report.MakeAdjacencyID(dstAddressID)
		)

		// Endpoint topology
		if _, ok := r.Endpoint.NodeMetadatas[srcPortID]; !ok {
			r.Endpoint.NodeMetadatas[srcPortID] = report.MakeNodeMetadataWith(map[string]string{
				process.PID: "4000",
				"name":      c.srcProc,
				"domain":    "node-" + src,
			})
		}
		r.Endpoint.Adjacency[srcID] = r.Endpoint.Adjacency[srcID].Add(dstPortID)
		if _, ok := r.Endpoint.NodeMetadatas[dstPortID]; !ok {
			r.Endpoint.NodeMetadatas[dstPortID] = report.MakeNodeMetadataWith(map[string]string{
				process.PID: "4000",
				"name":      c.dstProc,
				"domain":    "node-" + dst,
			})
		}
		r.Endpoint.Adjacency[dstID] = r.Endpoint.Adjacency[dstID].Add(srcPortID)
		var (
			edgeKeyEgress  = report.MakeEdgeID(srcPortID, dstPortID)
			edgeKeyIngress = report.MakeEdgeID(dstPortID, srcPortID)
		)
		r.Endpoint.EdgeMetadatas[edgeKeyEgress] = report.EdgeMetadata{
			MaxConnCountTCP: newu64(uint64(rand.Intn(100) + 10)),
		}
		r.Endpoint.EdgeMetadatas[edgeKeyIngress] = report.EdgeMetadata{
			MaxConnCountTCP: newu64(uint64(rand.Intn(100) + 10)),
		}

		// Address topology
		if _, ok := r.Address.NodeMetadatas[srcAddressID]; !ok {
			r.Address.NodeMetadatas[srcAddressID] = report.MakeNodeMetadataWith(map[string]string{
				docker.Name: src,
			})
		}
		r.Address.Adjacency[nodeSrcAddressID] = r.Address.Adjacency[nodeSrcAddressID].Add(dstAddressID)
		if _, ok := r.Address.NodeMetadatas[dstAddressID]; !ok {
			r.Address.NodeMetadatas[dstAddressID] = report.MakeNodeMetadataWith(map[string]string{
				docker.Name: dst,
			})
		}
		r.Address.Adjacency[nodeDstAddressID] = r.Address.Adjacency[nodeDstAddressID].Add(srcAddressID)

		// Host data
		r.Host.NodeMetadatas["hostX"] = report.MakeNodeMetadataWith(map[string]string{
			"ts":             time.Now().UTC().Format(time.RFC3339Nano),
			"host_name":      "host-x",
			"local_networks": localNet.String(),
			"os":             "linux",
		})
	}

	return r
}

func newu64(value uint64) *uint64 { return &value }
