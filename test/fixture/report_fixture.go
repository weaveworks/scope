package fixture

import (
	"time"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
)

// This is an example Report:
//   2 hosts with probes installed - client & server.
var (
	Now = time.Now()

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

	Client1Name      = "/usr/bin/curl"
	Client2Name      = "/usr/bin/curl"
	ServerName       = "apache"
	NonContainerName = "bash"

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
	ClientContainerName   = "client"
	ServerContainerID     = "5e4d3c2b1a"
	ClientContainerNodeID = report.MakeContainerNodeID(ClientContainerID)
	ServerContainerNodeID = report.MakeContainerNodeID(ServerContainerID)

	ClientContainerImageID     = "imageid123"
	ServerContainerImageID     = "imageid456"
	ClientContainerImageNodeID = report.MakeContainerImageNodeID(ClientContainerImageID)
	ServerContainerImageNodeID = report.MakeContainerImageNodeID(ServerContainerImageID)
	ClientContainerImageName   = "image/client"
	ServerContainerImageName   = "image/server"

	KubernetesNamespace = "ping"
	ClientPodID         = "ping/pong-a"
	ServerPodID         = "ping/pong-b"
	ClientPodNodeID     = report.MakePodNodeID(KubernetesNamespace, "pong-a")
	ServerPodNodeID     = report.MakePodNodeID(KubernetesNamespace, "pong-b")
	ServiceID           = "ping/pongservice"
	ServiceNodeID       = report.MakeServiceNodeID(KubernetesNamespace, "pongservice")

	ClientProcess1CPUMetric    = report.MakeMetric().Add(Now, 0.01).WithFirst(Now.Add(-1 * time.Second))
	ClientProcess1MemoryMetric = report.MakeMetric().Add(Now, 0.02).WithFirst(Now.Add(-2 * time.Second))

	ClientContainerCPUMetric    = report.MakeMetric().Add(Now, 0.03).WithFirst(Now.Add(-3 * time.Second))
	ClientContainerMemoryMetric = report.MakeMetric().Add(Now, 0.04).WithFirst(Now.Add(-4 * time.Second))

	ServerContainerCPUMetric    = report.MakeMetric().Add(Now, 0.05).WithFirst(Now.Add(-5 * time.Second))
	ServerContainerMemoryMetric = report.MakeMetric().Add(Now, 0.06).WithFirst(Now.Add(-6 * time.Second))

	ClientHostCPUMetric    = report.MakeMetric().Add(Now, 0.07).WithFirst(Now.Add(-7 * time.Second))
	ClientHostMemoryMetric = report.MakeMetric().Add(Now, 0.08).WithFirst(Now.Add(-8 * time.Second))
	ClientHostLoad1Metric  = report.MakeMetric().Add(Now, 0.09).WithFirst(Now.Add(-9 * time.Second))
	ClientHostLoad5Metric  = report.MakeMetric().Add(Now, 0.10).WithFirst(Now.Add(-10 * time.Second))
	ClientHostLoad15Metric = report.MakeMetric().Add(Now, 0.11).WithFirst(Now.Add(-11 * time.Second))

	ServerHostCPUMetric    = report.MakeMetric().Add(Now, 0.12).WithFirst(Now.Add(-12 * time.Second))
	ServerHostMemoryMetric = report.MakeMetric().Add(Now, 0.13).WithFirst(Now.Add(-13 * time.Second))
	ServerHostLoad1Metric  = report.MakeMetric().Add(Now, 0.14).WithFirst(Now.Add(-14 * time.Second))
	ServerHostLoad5Metric  = report.MakeMetric().Add(Now, 0.15).WithFirst(Now.Add(-15 * time.Second))
	ServerHostLoad15Metric = report.MakeMetric().Add(Now, 0.16).WithFirst(Now.Add(-16 * time.Second))

	Report = report.Report{
		ID: "test-report",
		Endpoint: report.Topology{
			Nodes: report.Nodes{
				// Node is arbitrary. We're free to put only precisely what we
				// care to test into the fixture. Just be sure to include the bits
				// that the mapping funcs extract :)
				Client54001NodeID: report.MakeNode().WithID(Client54001NodeID).WithTopology(report.Endpoint).WithLatests(map[string]string{
					endpoint.Addr:      ClientIP,
					endpoint.Port:      ClientPort54001,
					process.PID:        Client1PID,
					report.HostNodeID:  ClientHostNodeID,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(10),
					EgressByteCount:   newu64(100),
				}),

				Client54002NodeID: report.MakeNode().WithID(Client54002NodeID).WithTopology(report.Endpoint).WithLatests(map[string]string{
					endpoint.Addr:      ClientIP,
					endpoint.Port:      ClientPort54002,
					process.PID:        Client2PID,
					report.HostNodeID:  ClientHostNodeID,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(20),
					EgressByteCount:   newu64(200),
				}),

				Server80NodeID: report.MakeNode().WithID(Server80NodeID).WithTopology(report.Endpoint).WithLatests(map[string]string{
					endpoint.Addr:      ServerIP,
					endpoint.Port:      ServerPort,
					process.PID:        ServerPID,
					report.HostNodeID:  ServerHostNodeID,
					endpoint.Procspied: True,
				}),

				NonContainerNodeID: report.MakeNode().WithID(NonContainerNodeID).WithTopology(report.Endpoint).WithLatests(map[string]string{
					endpoint.Addr:      ServerIP,
					endpoint.Port:      NonContainerClientPort,
					process.PID:        NonContainerPID,
					report.HostNodeID:  ServerHostNodeID,
					endpoint.Procspied: True,
				}).WithAdjacent(GoogleEndpointNodeID),

				// Probe pseudo nodes
				UnknownClient1NodeID: report.MakeNode().WithID(UnknownClient1NodeID).WithTopology(report.Endpoint).WithLatests(map[string]string{
					endpoint.Addr:      UnknownClient1IP,
					endpoint.Port:      UnknownClient1Port,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(30),
					EgressByteCount:   newu64(300),
				}),

				UnknownClient2NodeID: report.MakeNode().WithID(UnknownClient2NodeID).WithTopology(report.Endpoint).WithLatests(map[string]string{
					endpoint.Addr:      UnknownClient2IP,
					endpoint.Port:      UnknownClient2Port,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(40),
					EgressByteCount:   newu64(400),
				}),

				UnknownClient3NodeID: report.MakeNode().WithID(UnknownClient3NodeID).WithTopology(report.Endpoint).WithLatests(map[string]string{
					endpoint.Addr:      UnknownClient3IP,
					endpoint.Port:      UnknownClient3Port,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(50),
					EgressByteCount:   newu64(500),
				}),

				RandomClientNodeID: report.MakeNode().WithID(RandomClientNodeID).WithTopology(report.Endpoint).WithLatests(map[string]string{
					endpoint.Addr:      RandomClientIP,
					endpoint.Port:      RandomClientPort,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(60),
					EgressByteCount:   newu64(600),
				}),

				GoogleEndpointNodeID: report.MakeNode().WithID(GoogleEndpointNodeID).WithTopology(report.Endpoint).WithLatests(map[string]string{
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
					process.Name:       Client1Name,
					docker.ContainerID: ClientContainerID,
					report.HostNodeID:  ClientHostNodeID,
				}).WithID(ClientProcess1NodeID).WithTopology(report.Process).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ClientHostNodeID)).
					Add("container", report.MakeStringSet(ClientContainerNodeID)).
					Add("container_image", report.MakeStringSet(ClientContainerImageNodeID)),
				).WithMetrics(report.Metrics{
					process.CPUUsage:    ClientProcess1CPUMetric,
					process.MemoryUsage: ClientProcess1MemoryMetric,
				}),
				ClientProcess2NodeID: report.MakeNodeWith(map[string]string{
					process.PID:        Client2PID,
					process.Name:       Client2Name,
					docker.ContainerID: ClientContainerID,
					report.HostNodeID:  ClientHostNodeID,
				}).WithID(ClientProcess2NodeID).WithTopology(report.Process).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ClientHostNodeID)).
					Add("container", report.MakeStringSet(ClientContainerNodeID)).
					Add("container_image", report.MakeStringSet(ClientContainerImageNodeID)),
				),
				ServerProcessNodeID: report.MakeNodeWith(map[string]string{
					process.PID:        ServerPID,
					process.Name:       ServerName,
					docker.ContainerID: ServerContainerID,
					report.HostNodeID:  ServerHostNodeID,
				}).WithID(ServerProcessNodeID).WithTopology(report.Process).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ServerHostNodeID)).
					Add("container", report.MakeStringSet(ServerContainerNodeID)).
					Add("container_image", report.MakeStringSet(ServerContainerImageNodeID)),
				),
				NonContainerProcessNodeID: report.MakeNodeWith(map[string]string{
					process.PID:       NonContainerPID,
					process.Name:      NonContainerName,
					report.HostNodeID: ServerHostNodeID,
				}).WithID(NonContainerProcessNodeID).WithTopology(report.Process).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ServerHostNodeID)),
				),
			},
		},
		Container: report.Topology{
			Nodes: report.Nodes{
				ClientContainerNodeID: report.MakeNodeWith(map[string]string{
					docker.ContainerID:                            ClientContainerID,
					docker.ContainerName:                          ClientContainerName,
					docker.ImageID:                                ClientContainerImageID,
					report.HostNodeID:                             ClientHostNodeID,
					docker.LabelPrefix + "io.kubernetes.pod.name": ClientPodID,
					kubernetes.PodID:                              ClientPodID,
					kubernetes.Namespace:                          KubernetesNamespace,
					docker.ContainerState:                         docker.StateRunning,
				}).WithID(ClientContainerNodeID).WithTopology(report.Container).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ClientHostNodeID)).
					Add("container_image", report.MakeStringSet(ClientContainerImageNodeID)).
					Add("pod", report.MakeStringSet(ClientPodID)),
				).WithMetrics(report.Metrics{
					docker.CPUTotalUsage: ClientContainerCPUMetric,
					docker.MemoryUsage:   ClientContainerMemoryMetric,
				}),
				ServerContainerNodeID: report.MakeNodeWith(map[string]string{
					docker.ContainerID:                                        ServerContainerID,
					docker.ContainerName:                                      "task-name-5-server-aceb93e2f2b797caba01",
					docker.ContainerState:                                     docker.StateRunning,
					docker.ImageID:                                            ServerContainerImageID,
					report.HostNodeID:                                         ServerHostNodeID,
					docker.LabelPrefix + detailed.AmazonECSContainerNameLabel: "server",
					docker.LabelPrefix + "foo1":                               "bar1",
					docker.LabelPrefix + "foo2":                               "bar2",
					docker.LabelPrefix + "io.kubernetes.pod.name":             ServerPodID,
					kubernetes.PodID:                                          ServerPodID,
					kubernetes.Namespace:                                      KubernetesNamespace,
				}).WithID(ServerContainerNodeID).WithTopology(report.Container).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ServerHostNodeID)).
					Add("container_image", report.MakeStringSet(ServerContainerImageNodeID)).
					Add("pod", report.MakeStringSet(ServerPodID)),
				).WithMetrics(report.Metrics{
					docker.CPUTotalUsage: ServerContainerCPUMetric,
					docker.MemoryUsage:   ServerContainerMemoryMetric,
				}),
			},
		},
		ContainerImage: report.Topology{
			Nodes: report.Nodes{
				ClientContainerImageNodeID: report.MakeNodeWith(map[string]string{
					docker.ImageID:    ClientContainerImageID,
					docker.ImageName:  ClientContainerImageName,
					report.HostNodeID: ClientHostNodeID,
				}).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ClientHostNodeID)),
				).WithID(ClientContainerImageNodeID).WithTopology(report.ContainerImage),
				ServerContainerImageNodeID: report.MakeNodeWith(map[string]string{
					docker.ImageID:              ServerContainerImageID,
					docker.ImageName:            ServerContainerImageName,
					report.HostNodeID:           ServerHostNodeID,
					docker.LabelPrefix + "foo1": "bar1",
					docker.LabelPrefix + "foo2": "bar2",
				}).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ServerHostNodeID)),
				).WithID(ServerContainerImageNodeID).WithTopology(report.ContainerImage),
			},
		},
		Host: report.Topology{
			Nodes: report.Nodes{
				ClientHostNodeID: report.MakeNodeWith(map[string]string{
					"host_name":       ClientHostName,
					"os":              "Linux",
					report.HostNodeID: ClientHostNodeID,
				}).WithID(ClientHostNodeID).WithTopology(report.Host).WithSets(report.EmptySets.
					Add(host.LocalNetworks, report.MakeStringSet("10.10.10.0/24")),
				).WithMetrics(report.Metrics{
					host.CPUUsage:    ClientHostCPUMetric,
					host.MemoryUsage: ClientHostMemoryMetric,
					host.Load1:       ClientHostLoad1Metric,
					host.Load5:       ClientHostLoad5Metric,
					host.Load15:      ClientHostLoad15Metric,
				}),
				ServerHostNodeID: report.MakeNodeWith(map[string]string{
					"host_name":       ServerHostName,
					"os":              "Linux",
					report.HostNodeID: ServerHostNodeID,
				}).WithID(ServerHostNodeID).WithTopology(report.Host).WithSets(report.EmptySets.
					Add(host.LocalNetworks, report.MakeStringSet("10.10.10.0/24")),
				).WithMetrics(report.Metrics{
					host.CPUUsage:    ServerHostCPUMetric,
					host.MemoryUsage: ServerHostMemoryMetric,
					host.Load1:       ServerHostLoad1Metric,
					host.Load5:       ServerHostLoad5Metric,
					host.Load15:      ServerHostLoad15Metric,
				}),
			},
		},
		Pod: report.Topology{
			Nodes: report.Nodes{
				ClientPodNodeID: report.MakeNodeWith(map[string]string{
					kubernetes.PodID:           ClientPodID,
					kubernetes.PodName:         "pong-a",
					kubernetes.Namespace:       KubernetesNamespace,
					kubernetes.PodContainerIDs: ClientContainerID,
					kubernetes.ServiceIDs:      ServiceID,
				}).WithID(ClientPodNodeID).WithTopology(report.Pod).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ClientHostNodeID)).
					Add("service", report.MakeStringSet(ServiceID)),
				),
				ServerPodNodeID: report.MakeNodeWith(map[string]string{
					kubernetes.PodID:           ServerPodID,
					kubernetes.PodName:         "pong-b",
					kubernetes.Namespace:       KubernetesNamespace,
					kubernetes.PodContainerIDs: ServerContainerID,
					kubernetes.ServiceIDs:      ServiceID,
				}).WithID(ServerPodNodeID).WithTopology(report.Pod).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ServerHostNodeID)).
					Add("service", report.MakeStringSet(ServiceID)),
				),
			},
		},
		Service: report.Topology{
			Nodes: report.Nodes{
				ServiceNodeID: report.MakeNodeWith(map[string]string{
					kubernetes.ServiceID:   ServiceID,
					kubernetes.ServiceName: "pongservice",
					kubernetes.Namespace:   "ping",
				}).WithID(ServiceNodeID).WithTopology(report.Service),
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
