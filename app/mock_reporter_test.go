package main

import (
	"net"

	"github.com/weaveworks/scope/report"
)

// StaticReport is used as know test data in api tests.
type StaticReport struct{}

func (s StaticReport) Report() report.Report {
	_, localNet, err := net.ParseCIDR("192.168.1.1/24")
	if err != nil {
		panic(err.Error())
	}

	var testReport = report.Report{
		Endpoint: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID("hostA", report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345")): report.MakeIDList(report.MakeEndpointNodeID("hostB", "192.168.1.2", "80")),
				report.MakeAdjacencyID("hostA", report.MakeEndpointNodeID("hostA", "192.168.1.1", "12346")): report.MakeIDList(report.MakeEndpointNodeID("hostB", "192.168.1.2", "80")),
				report.MakeAdjacencyID("hostA", report.MakeEndpointNodeID("hostA", "192.168.1.1", "8888")):  report.MakeIDList(report.MakeEndpointNodeID("", "1.2.3.4", "22")),
				report.MakeAdjacencyID("hostB", report.MakeEndpointNodeID("hostB", "192.168.1.2", "80")):    report.MakeIDList(report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345")),
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345"), report.MakeEndpointNodeID("hostB", "192.168.1.2", "80")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  200,
				},
				report.MakeEdgeID(report.MakeEndpointNodeID("hostA", "192.168.1.1", "12346"), report.MakeEndpointNodeID("hostB", "192.168.1.2", "80")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  201,
				},
				report.MakeEdgeID(report.MakeEndpointNodeID("hostA", "192.168.1.1", "8888"), report.MakeEndpointNodeID("", "1.2.3.4", "80")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      200,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  202,
				},
				report.MakeEdgeID(report.MakeEndpointNodeID("hostB", "192.168.1.2", "80"), report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      0,
					BytesIngress:     12,
					WithConnCountTCP: true,
					MaxConnCountTCP:  203,
				},
			},
			NodeMetadatas: report.NodeMetadatas{
				report.MakeEndpointNodeID("hostA", "192.168.1.1", "12345"): report.NodeMetadata{
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
				report.MakeEndpointNodeID("hostA", "192.168.1.1", "12346"): report.NodeMetadata{ // <-- same as :12345
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
				report.MakeEndpointNodeID("hostA", "192.168.1.1", "8888"): report.NodeMetadata{
					"pid":    "55100",
					"name":   "ssh",
					"domain": "node-a.local",
				},
				report.MakeEndpointNodeID("hostB", "192.168.1.2", "80"): report.NodeMetadata{
					"pid":    "215",
					"name":   "apache",
					"domain": "node-b.local",
				},
			},
		},

		Address: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID("hostA", report.MakeAddressNodeID("hostA", "192.168.1.1")): report.MakeIDList(report.MakeAddressNodeID("hostB", "192.168.1.2"), report.MakeAddressNodeID("", "1.2.3.4")),
				report.MakeAdjacencyID("hostB", report.MakeAddressNodeID("hostB", "192.168.1.2")): report.MakeIDList(report.MakeAddressNodeID("hostA", "192.168.1.1")),
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(report.MakeAddressNodeID("hostA", "192.168.1.1"), report.MakeAddressNodeID("hostB", "192.168.1.2")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  14,
				},
				report.MakeEdgeID(report.MakeAddressNodeID("hostA", "192.168.1.1"), report.MakeAddressNodeID("", "1.2.3.4")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      200,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  15,
				},
				report.MakeEdgeID(report.MakeAddressNodeID("hostB", "192.168.1.2"), report.MakeAddressNodeID("hostA", "192.168.1.1")): report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      0,
					BytesIngress:     12,
					WithConnCountTCP: true,
					MaxConnCountTCP:  16,
				},
			},
			NodeMetadatas: report.NodeMetadatas{
				report.MakeAddressNodeID("hostA", "192.168.1.1"): report.NodeMetadata{
					"name": "host-a",
				},
				report.MakeAddressNodeID("hostB", "192.168.1.2"): report.NodeMetadata{
					"name": "host-b",
				},
			},
		},

		Host: report.Topology{
			Adjacency:     report.Adjacency{},
			EdgeMetadatas: report.EdgeMetadatas{},
			NodeMetadatas: report.NodeMetadatas{
				report.MakeHostNodeID("hostA"): report.NodeMetadata{
					"host_name":      "node-a.local",
					"os":             "Linux",
					"local_networks": localNet.String(),
					"load":           "3.14 2.71 1.61",
				},
				report.MakeHostNodeID("hostB"): report.NodeMetadata{
					"host_name":      "node-b.local",
					"os":             "Linux",
					"local_networks": localNet.String(),
				},
			},
		},
	}
	return testReport.SquashRemote()
}
