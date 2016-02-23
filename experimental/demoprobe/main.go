package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/appclient"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

func main() {
	var (
		publish         = flag.String("publish", fmt.Sprintf("localhost:%d", xfer.AppPort), "publish target")
		publishInterval = flag.Duration("publish.interval", 1*time.Second, "publish (output) interval")
		hostCount       = flag.Int("hostcount", 10, "Number of demo hosts to generate")
	)
	flag.Parse()

	client, err := appclient.NewAppClient(appclient.ProbeConfig{
		Token:    "demoprobe",
		ProbeID:  "demoprobe",
		Insecure: false,
	}, *publish, *publish, nil)
	if err != nil {
		log.Fatal(err)
	}
	rp := appclient.NewReportPublisher(client)

	rand.Seed(time.Now().UnixNano())
	for range time.Tick(*publishInterval) {
		if err := rp.Publish(demoReport(*hostCount)); err != nil {
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
			c            = procPool[rand.Intn(len(procPool))]
			src          = hosts[rand.Intn(len(hosts))]
			dst          = hosts[rand.Intn(len(hosts))]
			srcPort      = rand.Intn(50000) + 10000
			srcPortID    = report.MakeEndpointNodeID("", src, strconv.Itoa(srcPort))
			dstPortID    = report.MakeEndpointNodeID("", dst, strconv.Itoa(c.dstPort))
			srcAddressID = report.MakeAddressNodeID("", src)
			dstAddressID = report.MakeAddressNodeID("", dst)
		)

		// Endpoint topology
		r.Endpoint = r.Endpoint.AddNode(srcPortID, report.MakeNode().WithLatests(map[string]string{
			process.PID: "4000",
			"name":      c.srcProc,
			"domain":    "node-" + src,
		}).WithEdge(dstPortID, report.EdgeMetadata{}))
		r.Endpoint = r.Endpoint.AddNode(dstPortID, report.MakeNode().WithLatests(map[string]string{
			process.PID: "4000",
			"name":      c.dstProc,
			"domain":    "node-" + dst,
		}).WithEdge(srcPortID, report.EdgeMetadata{}))

		// Address topology
		r.Address = r.Address.AddNode(srcAddressID, report.MakeNode().WithLatests(map[string]string{
			docker.Name: src,
		}).WithAdjacent(dstAddressID))
		r.Address = r.Address.AddNode(srcAddressID, report.MakeNode().WithLatests(map[string]string{
			docker.Name: dst,
		}).WithAdjacent(srcAddressID))

		// Host data
		r.Host = r.Host.AddNode("hostX", report.MakeNodeWith(map[string]string{
			"ts":             time.Now().UTC().Format(time.RFC3339Nano),
			"host_name":      "host-x",
			"local_networks": localNet.String(),
			"os":             "linux",
		}))
	}

	return r
}

func newu64(value uint64) *uint64 { return &value }
