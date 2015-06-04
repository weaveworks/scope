package report_test

import (
	"github.com/weaveworks/scope/report"
)

var reportFixture = report.Report{
	Endpoint: report.Topology{
		Adjacency: report.Adjacency{
			report.MakeAdjacencyID(clientHostID, client54001EndpointNodeID): report.MakeIDList(server80EndpointNodeID),
			report.MakeAdjacencyID(clientHostID, client54002EndpointNodeID): report.MakeIDList(server80EndpointNodeID),
			report.MakeAdjacencyID(serverHostID, server80EndpointNodeID):    report.MakeIDList(client54001EndpointNodeID, client54002EndpointNodeID, unknown1EndpointNodeID, unknown2EndpointNodeID, unknown3EndpointNodeID),
		},
		NodeMetadatas: report.NodeMetadatas{
			client54001EndpointNodeID: report.NodeMetadata{
				"process_node_id": report.MakeProcessNodeID(clientHostID, "4242"),
				"address_node_id": report.MakeAddressNodeID(clientHostID, clientAddress),
			},
			client54002EndpointNodeID: report.NodeMetadata{
				//"process_node_id": report.MakeProcessNodeID(clientHostID, "4242"), // leave it out, to test a branch in Render
				"address_node_id": report.MakeAddressNodeID(clientHostID, clientAddress),
			},
			server80EndpointNodeID: report.NodeMetadata{
				"process_node_id": report.MakeProcessNodeID(serverHostID, "215"),
				"address_node_id": report.MakeAddressNodeID(serverHostID, serverAddress),
			},

			"process-not-available": report.NodeMetadata{},                                  // for TestProcess{PID,Name,Container[Name]}
			"process-badly-linked":  report.NodeMetadata{"process_node_id": "none"},         // for TestProcess{PID,Name,Container[Name]}
			"process-no-container":  report.NodeMetadata{"process_node_id": "no-container"}, // for TestProcessContainer[Name]
			"address-not-available": report.NodeMetadata{},                                  // for TestAddressHostname
			"address-badly-linked":  report.NodeMetadata{"address_node_id": "none"},         // for TestAddressHostname
		},
		EdgeMetadatas: report.EdgeMetadatas{
			report.MakeEdgeID(client54001EndpointNodeID, server80EndpointNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  10,  // src -> dst
				BytesIngress: 100, // src <- dst
			},
			report.MakeEdgeID(client54002EndpointNodeID, server80EndpointNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  20,
				BytesIngress: 200,
			},
			report.MakeEdgeID(server80EndpointNodeID, client54001EndpointNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  100,
				BytesIngress: 10,
			},
			report.MakeEdgeID(server80EndpointNodeID, client54002EndpointNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  200,
				BytesIngress: 20,
			},
			report.MakeEdgeID(server80EndpointNodeID, unknown1EndpointNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  400,
				BytesIngress: 40,
			},
			report.MakeEdgeID(server80EndpointNodeID, unknown2EndpointNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  800,
				BytesIngress: 80,
			},
			report.MakeEdgeID(server80EndpointNodeID, unknown3EndpointNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  1600,
				BytesIngress: 160,
			},
		},
	},
	Address: report.Topology{
		Adjacency: report.Adjacency{
			report.MakeAdjacencyID(clientHostID, clientAddressNodeID): report.MakeIDList(serverAddressNodeID),
			report.MakeAdjacencyID(serverHostID, serverAddressNodeID): report.MakeIDList(clientAddressNodeID, unknownAddressNodeID),
		},
		NodeMetadatas: report.NodeMetadatas{
			clientAddressNodeID: report.NodeMetadata{
				"host_name": "client.host.com",
			},
			serverAddressNodeID: report.NodeMetadata{},

			"no-host-name": report.NodeMetadata{},
		},
		EdgeMetadatas: report.EdgeMetadatas{
			report.MakeEdgeID(clientAddressNodeID, serverAddressNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  10 + 20 + 1,
				BytesIngress: 100 + 200 + 2,
			},
			report.MakeEdgeID(serverAddressNodeID, clientAddressNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  100 + 200 + 3,
				BytesIngress: 10 + 20 + 4,
			},
			report.MakeEdgeID(serverAddressNodeID, unknownAddressNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  400 + 800 + 1600 + 5,
				BytesIngress: 40 + 80 + 160 + 6,
			},
		},
	},
	Process: report.Topology{
		Adjacency: report.Adjacency{},
		NodeMetadatas: report.NodeMetadatas{
			report.MakeProcessNodeID(clientHostID, "4242"): report.NodeMetadata{
				"host_name":             "client.host.com",
				"pid":                   "4242",
				"process_name":          "curl",
				"docker_container_id":   "a1b2c3d4e5",
				"docker_container_name": "fixture-container",
				"docker_image_id":       "0000000000",
				"docker_image_name":     "fixture/container:latest",
			},
			report.MakeProcessNodeID(serverHostID, "215"): report.NodeMetadata{
				"pid":          "215",
				"process_name": "apache",
			},

			"no-container": report.NodeMetadata{},
		},
		EdgeMetadatas: report.EdgeMetadatas{},
	},
	Host: report.Topology{
		Adjacency: report.Adjacency{},
		NodeMetadatas: report.NodeMetadatas{
			report.MakeHostNodeID(clientHostID): report.NodeMetadata{
				"host_name":      clientHostName,
				"local_networks": "10.10.10.0/24",
				"os":             "OS/2",
				"load":           "0.11 0.22 0.33",
			},
			report.MakeHostNodeID(serverHostID): report.NodeMetadata{
				"host_name":      serverHostName,
				"local_networks": "10.10.10.0/24",
				"os":             "Linux",
				"load":           "0.01 0.01 0.01",
			},
		},
		EdgeMetadatas: report.EdgeMetadatas{},
	},
}

var (
	clientHostID     = "client.host.com"
	clientHostName   = clientHostID
	clientHostNodeID = report.MakeHostNodeID(clientHostID)
	clientAddress    = "10.10.10.20"
	serverHostID     = "server.host.com"
	serverHostName   = serverHostID
	serverHostNodeID = report.MakeHostNodeID(serverHostID)
	serverAddress    = "10.10.10.1"
	unknownHostID    = ""              // by definition, we don't know it
	unknownAddress   = "172.16.93.112" // will be a pseudonode, no corresponding host

	client54001EndpointNodeID = report.MakeEndpointNodeID(clientHostID, clientAddress, "54001") // i.e. curl
	client54002EndpointNodeID = report.MakeEndpointNodeID(clientHostID, clientAddress, "54002") // also curl
	server80EndpointNodeID    = report.MakeEndpointNodeID(serverHostID, serverAddress, "80")    // i.e. apache
	unknown1EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, unknownAddress, "10001")
	unknown2EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, unknownAddress, "10002")
	unknown3EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, unknownAddress, "10003")

	clientAddressNodeID  = report.MakeAddressNodeID(clientHostID, clientAddress)
	serverAddressNodeID  = report.MakeAddressNodeID(serverHostID, serverAddress)
	unknownAddressNodeID = report.MakeAddressNodeID(unknownHostID, unknownAddress)
)
