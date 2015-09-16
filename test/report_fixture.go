package test

import (
	"time"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

// This is an example Report:
//   2 hosts with probes installed - client & server.
var (
	ClientHostID  = "client.hostname.com"
	ServerHostID  = "server.hostname.com"
	UnknownHostID = ""

	ClientIP         = "10.10.10.20"
	ServerIP         = "192.168.1.1"
	UnknownClient1IP = "10.10.10.10"
	UnknownClient2IP = "10.10.10.10"
	UnknownClient3IP = "10.10.10.11"
	RandomClientIP   = "51.52.53.54"
	GoogleIP         = "8.8.8.8"

	ClientPort54001        = "54001"
	ClientPort54002        = "54002"
	ServerPort             = "80"
	UnknownClient1Port     = "54010"
	UnknownClient2Port     = "54020"
	UnknownClient3Port     = "54020"
	RandomClientPort       = "12345"
	GooglePort             = "80"
	NonContainerClientPort = "46789"

	ClientHostName = ClientHostID
	ServerHostName = ServerHostID

	Client1PID      = "10001"
	Client2PID      = "30020"
	ServerPID       = "215"
	NonContainerPID = "1234"

	Client1Comm      = "curl"
	Client2Comm      = "curl"
	ServerComm       = "apache"
	NonContainerComm = "bash"

	True = "true"

	ClientHostNodeID = report.MakeHostNodeID(ClientHostID)
	ServerHostNodeID = report.MakeHostNodeID(ServerHostID)

	Client54001NodeID    = report.MakeEndpointNodeID(ClientHostID, ClientIP, ClientPort54001)            // curl (1)
	Client54002NodeID    = report.MakeEndpointNodeID(ClientHostID, ClientIP, ClientPort54002)            // curl (2)
	Server80NodeID       = report.MakeEndpointNodeID(ServerHostID, ServerIP, ServerPort)                 // apache
	UnknownClient1NodeID = report.MakeEndpointNodeID(ServerHostID, UnknownClient1IP, UnknownClient1Port) // we want to ensure two unknown clients, connnected
	UnknownClient2NodeID = report.MakeEndpointNodeID(ServerHostID, UnknownClient2IP, UnknownClient2Port) // to the same server, are deduped.
	UnknownClient3NodeID = report.MakeEndpointNodeID(ServerHostID, UnknownClient3IP, UnknownClient3Port) // Check this one isn't deduped
	RandomClientNodeID   = report.MakeEndpointNodeID(ServerHostID, RandomClientIP, RandomClientPort)     // this should become an internet node
	NonContainerNodeID   = report.MakeEndpointNodeID(ServerHostID, ServerIP, NonContainerClientPort)
	GoogleEndpointNodeID = report.MakeEndpointNodeID(ServerHostID, GoogleIP, GooglePort)

	ClientProcess1NodeID      = report.MakeProcessNodeID(ClientHostID, Client1PID)
	ClientProcess2NodeID      = report.MakeProcessNodeID(ClientHostID, Client2PID)
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
	ClientContainerImageName   = "image/client"
	ServerContainerImageName   = "image/server"

	ClientAddressNodeID   = report.MakeAddressNodeID(ClientHostID, ClientIP)
	ServerAddressNodeID   = report.MakeAddressNodeID(ServerHostID, ServerIP)
	UnknownAddress1NodeID = report.MakeAddressNodeID(ServerHostID, UnknownClient1IP)
	UnknownAddress2NodeID = report.MakeAddressNodeID(ServerHostID, UnknownClient2IP)
	UnknownAddress3NodeID = report.MakeAddressNodeID(ServerHostID, UnknownClient3IP)
	RandomAddressNodeID   = report.MakeAddressNodeID(ServerHostID, RandomClientIP) // this should become an internet node

	Report = report.Report{
		Endpoint: report.Topology{
			Nodes: report.Nodes{
				// Node is arbitrary. We're free to put only precisely what we
				// care to test into the fixture. Just be sure to include the bits
				// that the mapping funcs extract :)
				Client54001NodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:      ClientIP,
					endpoint.Port:      ClientPort54001,
					process.PID:        Client1PID,
					report.HostNodeID:  ClientHostNodeID,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(10),
					EgressByteCount:   newu64(100),
				}),

				Client54002NodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:      ClientIP,
					endpoint.Port:      ClientPort54002,
					process.PID:        Client2PID,
					report.HostNodeID:  ClientHostNodeID,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(20),
					EgressByteCount:   newu64(200),
				}),

				Server80NodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:      ServerIP,
					endpoint.Port:      ServerPort,
					process.PID:        ServerPID,
					report.HostNodeID:  ServerHostNodeID,
					endpoint.Procspied: True,
				}),

				NonContainerNodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:      ServerIP,
					endpoint.Port:      NonContainerClientPort,
					process.PID:        NonContainerPID,
					report.HostNodeID:  ServerHostNodeID,
					endpoint.Procspied: True,
				}).WithAdjacent(GoogleEndpointNodeID),

				// Probe pseudo nodes
				UnknownClient1NodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:      UnknownClient1IP,
					endpoint.Port:      UnknownClient1Port,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(30),
					EgressByteCount:   newu64(300),
				}),

				UnknownClient2NodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:      UnknownClient2IP,
					endpoint.Port:      UnknownClient2Port,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(40),
					EgressByteCount:   newu64(400),
				}),

				UnknownClient3NodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:      UnknownClient3IP,
					endpoint.Port:      UnknownClient3Port,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(50),
					EgressByteCount:   newu64(500),
				}),

				RandomClientNodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:      RandomClientIP,
					endpoint.Port:      RandomClientPort,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(60),
					EgressByteCount:   newu64(600),
				}),

				GoogleEndpointNodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:      GoogleIP,
					endpoint.Port:      GooglePort,
					endpoint.Procspied: True,
				}),
			},
		},
		Process: report.Topology{
			Nodes: report.Nodes{
				ClientProcess1NodeID: report.MakeNodeWith(map[string]string{
					process.PID:        Client1PID,
					"comm":             Client1Comm,
					docker.ContainerID: ClientContainerID,
					report.HostNodeID:  ClientHostNodeID,
				}),
				ClientProcess2NodeID: report.MakeNodeWith(map[string]string{
					process.PID:        Client2PID,
					"comm":             Client2Comm,
					docker.ContainerID: ClientContainerID,
					report.HostNodeID:  ClientHostNodeID,
				}),
				ServerProcessNodeID: report.MakeNodeWith(map[string]string{
					process.PID:        ServerPID,
					"comm":             ServerComm,
					docker.ContainerID: ServerContainerID,
					report.HostNodeID:  ServerHostNodeID,
				}),
				NonContainerProcessNodeID: report.MakeNodeWith(map[string]string{
					process.PID:       NonContainerPID,
					"comm":            NonContainerComm,
					report.HostNodeID: ServerHostNodeID,
				}),
			},
		},
		Container: report.Topology{
			Nodes: report.Nodes{
				ClientContainerNodeID: report.MakeNodeWith(map[string]string{
					docker.ContainerID:   ClientContainerID,
					docker.ContainerName: "client",
					docker.ImageID:       ClientContainerImageID,
					report.HostNodeID:    ClientHostNodeID,
				}),
				ServerContainerNodeID: report.MakeNodeWith(map[string]string{
					docker.ContainerID:                                      ServerContainerID,
					docker.ContainerName:                                    "task-name-5-server-aceb93e2f2b797caba01",
					docker.ImageID:                                          ServerContainerImageID,
					report.HostNodeID:                                       ServerHostNodeID,
					docker.LabelPrefix + render.AmazonECSContainerNameLabel: "server",
					docker.LabelPrefix + "foo1":                             "bar1",
					docker.LabelPrefix + "foo2":                             "bar2",
				}),
			},
		},
		ContainerImage: report.Topology{
			Nodes: report.Nodes{
				ClientContainerImageNodeID: report.MakeNodeWith(map[string]string{
					docker.ImageID:    ClientContainerImageID,
					docker.ImageName:  ClientContainerImageName,
					report.HostNodeID: ClientHostNodeID,
				}),
				ServerContainerImageNodeID: report.MakeNodeWith(map[string]string{
					docker.ImageID:              ServerContainerImageID,
					docker.ImageName:            ServerContainerImageName,
					report.HostNodeID:           ServerHostNodeID,
					docker.LabelPrefix + "foo1": "bar1",
					docker.LabelPrefix + "foo2": "bar2",
				}),
			},
		},
		Address: report.Topology{
			Nodes: report.Nodes{
				ClientAddressNodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:     ClientIP,
					report.HostNodeID: ClientHostNodeID,
				}).WithEdge(ServerAddressNodeID, report.EdgeMetadata{
					MaxConnCountTCP: newu64(3),
				}),

				ServerAddressNodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr:     ServerIP,
					report.HostNodeID: ServerHostNodeID,
				}),

				UnknownAddress1NodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr: UnknownClient1IP,
				}).WithAdjacency(report.MakeIDList(ServerAddressNodeID)),

				UnknownAddress2NodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr: UnknownClient2IP,
				}).WithAdjacency(report.MakeIDList(ServerAddressNodeID)),

				UnknownAddress3NodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr: UnknownClient3IP,
				}).WithAdjacency(report.MakeIDList(ServerAddressNodeID)),

				RandomAddressNodeID: report.MakeNode().WithMetadata(map[string]string{
					endpoint.Addr: RandomClientIP,
				}).WithAdjacency(report.MakeIDList(ServerAddressNodeID)),
			},
		},
		Host: report.Topology{
			Nodes: report.Nodes{
				ClientHostNodeID: report.MakeNodeWith(map[string]string{
					"host_name":       ClientHostName,
					"local_networks":  "10.10.10.0/24",
					"os":              "Linux",
					"load":            "0.01 0.01 0.01",
					report.HostNodeID: ClientHostNodeID,
				}),
				ServerHostNodeID: report.MakeNodeWith(map[string]string{
					"host_name":       ServerHostName,
					"local_networks":  "10.10.10.0/24",
					"os":              "Linux",
					"load":            "0.01 0.01 0.01",
					report.HostNodeID: ServerHostNodeID,
				}),
			},
		},
		Sampling: report.Sampling{
			Count: 1024,
			Total: 4096,
		},
		Window: 2 * time.Second,
	}
)

func init() {
	if err := Report.Validate(); err != nil {
		panic(err)
	}
}

func newu64(value uint64) *uint64 { return &value }
