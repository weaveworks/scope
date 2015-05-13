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
		Process: report.Topology{
			Adjacency: report.Adjacency{
				"hostA|;192.168.1.1;12345": []string{";192.168.1.2;80"},
				"hostA|;192.168.1.1;12346": []string{";192.168.1.2;80"},
				"hostA|;192.168.1.1;8888":  []string{";1.2.3.4;22"},
				"hostB|;192.168.1.2;80":    []string{";192.168.1.1;12345"},
			},
			EdgeMetadatas: report.EdgeMetadatas{
				";192.168.1.1;12345|;192.168.1.2;80": report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  200,
				},
				";192.168.1.1;12346|;192.168.1.2;80": report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  201,
				},
				";192.168.1.1;8888|;1.2.3.4;80": report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      200,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  202,
				},
				";192.168.1.2;80|;192.168.1.1;12345": report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      0,
					BytesIngress:     12,
					WithConnCountTCP: true,
					MaxConnCountTCP:  203,
				},
			},
			NodeMetadatas: report.NodeMetadatas{
				";192.168.1.1;12345": report.NodeMetadata{
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
				";192.168.1.1;12346": report.NodeMetadata{ // <-- same as :12345
					"pid":    "23128",
					"name":   "curl",
					"domain": "node-a.local",
				},
				";192.168.1.1;8888": report.NodeMetadata{
					"pid":    "55100",
					"name":   "ssh",
					"domain": "node-a.local",
				},
				";192.168.1.2;80": report.NodeMetadata{
					"pid":    "215",
					"name":   "apache",
					"domain": "node-b.local",
				},
			},
		},

		Network: report.Topology{
			Adjacency: report.Adjacency{
				"hostA|;192.168.1.1": []string{";192.168.1.2", ";1.2.3.4"},
				"hostB|;192.168.1.2": []string{";192.168.1.1"},
			},
			EdgeMetadatas: report.EdgeMetadatas{
				";192.168.1.1|;192.168.1.2": report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      12,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  14,
				},
				";192.168.1.1|;1.2.3.4": report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      200,
					BytesIngress:     0,
					WithConnCountTCP: true,
					MaxConnCountTCP:  15,
				},
				";192.168.1.2|;192.168.1.1": report.EdgeMetadata{
					WithBytes:        true,
					BytesEgress:      0,
					BytesIngress:     12,
					WithConnCountTCP: true,
					MaxConnCountTCP:  16,
				},
			},
			NodeMetadatas: report.NodeMetadatas{
				";192.168.1.1": report.NodeMetadata{
					"name": "host-a",
				},
				";192.168.1.2": report.NodeMetadata{
					"name": "host-b",
				},
			},
		},

		HostMetadatas: report.HostMetadatas{
			"hostA": report.HostMetadata{
				Hostname:    "node-a.local",
				LocalNets:   []*net.IPNet{localNet},
				OS:          "Linux",
				LoadOne:     3.1415,
				LoadFive:    2.7182,
				LoadFifteen: 1.6180,
			},
			"hostB": report.HostMetadata{
				Hostname:  "node-b.local",
				LocalNets: []*net.IPNet{localNet},
				OS:        "Linux",
			},
		},
	}
	return testReport.SquashRemote()
}
