package test

import (
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

// This is an example Report:
//   2 hosts with probes installed - client & server.
var (
	ClientHostID  = "client.hostname.com"
	ServerHostID  = "server.hostname.com"
	UnknownHostID = ""

	ClientIP        = "10.10.10.20"
	ServerIP        = "192.168.1.1"
	ClientPort54001 = "54001"
	ClientPort54002 = "54002"
	ServerPort      = "80"

	ClientHostName = ClientHostID
	ServerHostName = ServerHostID

	ClientPID       = "10001"
	ServerPID       = "215"
	NonContainerPID = "1234"

	ClientHostNodeID = report.MakeHostNodeID(ClientHostID)
	ServerHostNodeID = report.MakeHostNodeID(ServerHostID)

	Client54001NodeID    = report.MakeEndpointNodeID(ClientHostID, ClientIP, ClientPort54001) // curl (1)
	Client54002NodeID    = report.MakeEndpointNodeID(ClientHostID, ClientIP, ClientPort54002) // curl (2)
	Server80NodeID       = report.MakeEndpointNodeID(ServerHostID, ServerIP, ServerPort)      // apache
	UnknownClient1NodeID = report.MakeEndpointNodeID(ServerHostID, "10.10.10.10", "54010")    // we want to ensure two unknown clients, connnected
	UnknownClient2NodeID = report.MakeEndpointNodeID(ServerHostID, "10.10.10.10", "54020")    // to the same server, are deduped.
	UnknownClient3NodeID = report.MakeEndpointNodeID(ServerHostID, "10.10.10.11", "54020")    // Check this one isn't deduped
	RandomClientNodeID   = report.MakeEndpointNodeID(ServerHostID, "51.52.53.54", "12345")    // this should become an internet node

	ClientProcessNodeID       = report.MakeProcessNodeID(ClientHostID, ClientPID)
	ServerProcessNodeID       = report.MakeProcessNodeID(ServerHostID, ServerPID)
	NonContainerProcessNodeID = report.MakeProcessNodeID(ServerHostID, NonContainerPID)

	ClientContainerID     = "a1b2c3d4e5"
	ServerContainerID     = "5e4d3c2b1a"
	ClientContainerNodeID = report.MakeContainerNodeID(ClientHostID, ClientContainerID)
	ServerContainerNodeID = report.MakeContainerNodeID(ServerHostID, ServerContainerID)

	ClientContainerImageID     = "imageid123"
	ServerContainerImageID     = "imageid456"
	ClientContainerImageNodeID = report.MakeContainerNodeID(ClientHostID, ClientContainerImageID)
	ServerContainerImageNodeID = report.MakeContainerNodeID(ServerHostID, ServerContainerImageID)

	ClientAddressNodeID   = report.MakeAddressNodeID(ClientHostID, "10.10.10.20")
	ServerAddressNodeID   = report.MakeAddressNodeID(ServerHostID, "192.168.1.1")
	UnknownAddress1NodeID = report.MakeAddressNodeID(ServerHostID, "10.10.10.10")
	UnknownAddress2NodeID = report.MakeAddressNodeID(ServerHostID, "10.10.10.11")
	RandomAddressNodeID   = report.MakeAddressNodeID(ServerHostID, "51.52.53.54") // this should become an internet node

	Report = report.Report{
		Endpoint: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(Client54001NodeID): report.MakeIDList(Server80NodeID),
				report.MakeAdjacencyID(Client54002NodeID): report.MakeIDList(Server80NodeID),
				report.MakeAdjacencyID(Server80NodeID): report.MakeIDList(
					Client54001NodeID, Client54002NodeID, UnknownClient1NodeID, UnknownClient2NodeID,
					UnknownClient3NodeID, RandomClientNodeID),
			},
			NodeMetadatas: report.NodeMetadatas{
				// NodeMetadata is arbitrary. We're free to put only precisely what we
				// care to test into the fixture. Just be sure to include the bits
				// that the mapping funcs extract :)
				Client54001NodeID: report.NodeMetadata{
					"addr":            ClientIP,
					"port":            ClientPort54001,
					"pid":             ClientPID,
					report.HostNodeID: ClientHostNodeID,
				},
				Client54002NodeID: report.NodeMetadata{
					"addr":            ClientIP,
					"port":            ClientPort54002,
					"pid":             ClientPID, // should be same as above!
					report.HostNodeID: ClientHostNodeID,
				},
				Server80NodeID: report.NodeMetadata{
					"addr":            ServerIP,
					"port":            ServerPort,
					"pid":             ServerPID,
					report.HostNodeID: ServerHostNodeID,
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(Client54001NodeID, Server80NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 100,
					BytesEgress:  10,
				},
				report.MakeEdgeID(Client54002NodeID, Server80NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 200,
					BytesEgress:  20,
				},

				report.MakeEdgeID(Server80NodeID, Client54001NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 10,
					BytesEgress:  100,
				},
				report.MakeEdgeID(Server80NodeID, Client54002NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 20,
					BytesEgress:  200,
				},
				report.MakeEdgeID(Server80NodeID, UnknownClient1NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 30,
					BytesEgress:  300,
				},
				report.MakeEdgeID(Server80NodeID, UnknownClient2NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 40,
					BytesEgress:  400,
				},
				report.MakeEdgeID(Server80NodeID, UnknownClient3NodeID): report.EdgeMetadata{
					WithBytes:    true,
					BytesIngress: 50,
					BytesEgress:  500,
				},
			},
		},
		Process: report.Topology{
			Adjacency: report.Adjacency{},
			NodeMetadatas: report.NodeMetadatas{
				ClientProcessNodeID: report.NodeMetadata{
					"pid":              ClientPID,
					"comm":             "curl",
					docker.ContainerID: ClientContainerID,
					report.HostNodeID:  ClientHostNodeID,
				},
				ServerProcessNodeID: report.NodeMetadata{
					"pid":              ServerPID,
					"comm":             "apache",
					docker.ContainerID: ServerContainerID,
					report.HostNodeID:  ServerHostNodeID,
				},
				NonContainerProcessNodeID: report.NodeMetadata{
					"pid":             NonContainerPID,
					"comm":            "bash",
					report.HostNodeID: ServerHostNodeID,
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{},
		},
		Container: report.Topology{
			NodeMetadatas: report.NodeMetadatas{
				ClientContainerNodeID: report.NodeMetadata{
					docker.ContainerID:   ClientContainerID,
					docker.ContainerName: "client",
					docker.ImageID:       ClientContainerImageID,
					report.HostNodeID:    ClientHostNodeID,
				},
				ServerContainerNodeID: report.NodeMetadata{
					docker.ContainerID:   ServerContainerID,
					docker.ContainerName: "server",
					docker.ImageID:       ServerContainerImageID,
					report.HostNodeID:    ServerHostNodeID,
				},
			},
		},
		ContainerImage: report.Topology{
			NodeMetadatas: report.NodeMetadatas{
				ClientContainerImageNodeID: report.NodeMetadata{
					docker.ImageID:    ClientContainerImageID,
					docker.ImageName:  "client_image",
					report.HostNodeID: ClientHostNodeID,
				},
				ServerContainerImageNodeID: report.NodeMetadata{
					docker.ImageID:    ServerContainerImageID,
					docker.ImageName:  "server_image",
					report.HostNodeID: ServerHostNodeID,
				},
			},
		},
		Address: report.Topology{
			Adjacency: report.Adjacency{
				report.MakeAdjacencyID(ClientAddressNodeID): report.MakeIDList(ServerAddressNodeID),
				report.MakeAdjacencyID(ServerAddressNodeID): report.MakeIDList(
					ClientAddressNodeID, UnknownAddress1NodeID, UnknownAddress2NodeID, RandomAddressNodeID), // no backlinks to unknown/random
			},
			NodeMetadatas: report.NodeMetadatas{
				ClientAddressNodeID: report.NodeMetadata{
					"addr":            ClientIP,
					report.HostNodeID: ClientHostNodeID,
				},
				ServerAddressNodeID: report.NodeMetadata{
					"addr":            ServerIP,
					report.HostNodeID: ServerHostNodeID,
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{
				report.MakeEdgeID(ClientAddressNodeID, ServerAddressNodeID): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  3,
				},
				report.MakeEdgeID(ServerAddressNodeID, ClientAddressNodeID): report.EdgeMetadata{
					WithConnCountTCP: true,
					MaxConnCountTCP:  3,
				},
			},
		},
		Host: report.Topology{
			Adjacency: report.Adjacency{},
			NodeMetadatas: report.NodeMetadatas{
				ClientHostNodeID: report.NodeMetadata{
					"host_name":       ClientHostName,
					"local_networks":  "10.10.10.0/24",
					"os":              "Linux",
					"load":            "0.01 0.01 0.01",
					report.HostNodeID: ClientHostNodeID,
				},
				ServerHostNodeID: report.NodeMetadata{
					"host_name":       ServerHostName,
					"local_networks":  "10.10.10.0/24",
					"os":              "Linux",
					"load":            "0.01 0.01 0.01",
					report.HostNodeID: ServerHostNodeID,
				},
			},
			EdgeMetadatas: report.EdgeMetadatas{},
		},
	}
)

func init() {
	if err := Report.Validate(); err != nil {
		panic(err)
	}
}
