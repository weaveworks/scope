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

	Client54001NodeID    = report.MakeEndpointNodeID(ClientHostID, "", ClientIP, ClientPort54001)            // curl (1)
	Client54002NodeID    = report.MakeEndpointNodeID(ClientHostID, "", ClientIP, ClientPort54002)            // curl (2)
	Server80NodeID       = report.MakeEndpointNodeID(ServerHostID, "", ServerIP, ServerPort)                 // apache
	UnknownClient1NodeID = report.MakeEndpointNodeID(ServerHostID, "", UnknownClient1IP, UnknownClient1Port) // we want to ensure two unknown clients, connnected
	UnknownClient2NodeID = report.MakeEndpointNodeID(ServerHostID, "", UnknownClient2IP, UnknownClient2Port) // to the same server, are deduped.
	UnknownClient3NodeID = report.MakeEndpointNodeID(ServerHostID, "", UnknownClient3IP, UnknownClient3Port) // Check this one isn't deduped
	RandomClientNodeID   = report.MakeEndpointNodeID(ServerHostID, "", RandomClientIP, RandomClientPort)     // this should become an internet node
	NonContainerNodeID   = report.MakeEndpointNodeID(ServerHostID, "", ServerIP, NonContainerClientPort)
	GoogleEndpointNodeID = report.MakeEndpointNodeID(ServerHostID, "", GoogleIP, GooglePort)

	ClientProcess1NodeID      = report.MakeProcessNodeID(ClientHostID, Client1PID)
	ClientProcess2NodeID      = report.MakeProcessNodeID(ClientHostID, Client2PID)
	ServerProcessNodeID       = report.MakeProcessNodeID(ServerHostID, ServerPID)
	NonContainerProcessNodeID = report.MakeProcessNodeID(ServerHostID, NonContainerPID)

	ClientContainerID     = "a1b2c3d4e5"
	ClientContainerName   = "client"
	ServerContainerID     = "5e4d3c2b1a"
	ServerContainerName   = "task-name-5-server-aceb93e2f2b797caba01"
	ClientContainerNodeID = report.MakeContainerNodeID(ClientContainerID)
	ServerContainerNodeID = report.MakeContainerNodeID(ServerContainerID)

	ClientContainerHostname = ClientContainerName + ".hostname.com"
	ServerContainerHostname = ServerContainerName + ".hostname.com"

	ClientContainerImageID     = "imageid123"
	ServerContainerImageID     = "imageid456"
	ClientContainerImageNodeID = report.MakeContainerImageNodeID(ClientContainerImageID)
	ServerContainerImageNodeID = report.MakeContainerImageNodeID(ServerContainerImageID)
	ClientContainerImageName   = "image/client"
	ServerContainerImageName   = "image/server"

	KubernetesNamespace = "ping"
	ClientPodUID        = "5d4c3b2a1"
	ServerPodUID        = "i9h8g7f6e"
	ClientPodNodeID     = report.MakePodNodeID(ClientPodUID)
	ServerPodNodeID     = report.MakePodNodeID(ServerPodUID)
	ServiceName         = "pongservice"
	ServiceUID          = "service1234"
	ServiceNodeID       = report.MakeServiceNodeID(ServiceUID)

	ClientProcess1CPUMetric    = report.MakeSingletonMetric(Now.Add(-1*time.Second), 0.01)
	ClientProcess1MemoryMetric = report.MakeSingletonMetric(Now.Add(-2*time.Second), 0.02)

	ClientContainerCPUMetric    = report.MakeSingletonMetric(Now.Add(-3*time.Second), 0.03)
	ClientContainerMemoryMetric = report.MakeSingletonMetric(Now.Add(-4*time.Second), 0.04)

	ServerContainerCPUMetric    = report.MakeSingletonMetric(Now.Add(-5*time.Second), 0.05)
	ServerContainerMemoryMetric = report.MakeSingletonMetric(Now.Add(-6*time.Second), 0.06)

	ClientHostCPUMetric    = report.MakeSingletonMetric(Now.Add(-7*time.Second), 0.07)
	ClientHostMemoryMetric = report.MakeSingletonMetric(Now.Add(-8*time.Second), 0.08)
	ClientHostLoad1Metric  = report.MakeSingletonMetric(Now.Add(-9*time.Second), 0.09)

	ServerHostCPUMetric    = report.MakeSingletonMetric(Now.Add(-12*time.Second), 0.12)
	ServerHostMemoryMetric = report.MakeSingletonMetric(Now.Add(-13*time.Second), 0.13)
	ServerHostLoad1Metric  = report.MakeSingletonMetric(Now.Add(-14*time.Second), 0.14)

	Report = report.Report{
		ID: "test-report",
		Endpoint: report.Topology{
			Nodes: report.Nodes{
				// Node is arbitrary. We're free to put only precisely what we
				// care to test into the fixture. Just be sure to include the bits
				// that the mapping funcs extract :)
				Client54001NodeID: report.MakeNode(Client54001NodeID).WithTopology(report.Endpoint).WithConsts(map[string]string{
					endpoint.Addr:      ClientIP,
					endpoint.Port:      ClientPort54001,
					process.PID:        Client1PID,
					report.HostNodeID:  ClientHostNodeID,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(10),
					EgressByteCount:   newu64(100),
				}),

				Client54002NodeID: report.MakeNode(Client54002NodeID).WithTopology(report.Endpoint).WithConsts(map[string]string{
					endpoint.Addr:      ClientIP,
					endpoint.Port:      ClientPort54002,
					process.PID:        Client2PID,
					report.HostNodeID:  ClientHostNodeID,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(20),
					EgressByteCount:   newu64(200),
				}),

				Server80NodeID: report.MakeNode(Server80NodeID).WithTopology(report.Endpoint).WithConsts(map[string]string{
					endpoint.Addr:      ServerIP,
					endpoint.Port:      ServerPort,
					process.PID:        ServerPID,
					report.HostNodeID:  ServerHostNodeID,
					endpoint.Procspied: True,
				}),

				NonContainerNodeID: report.MakeNode(NonContainerNodeID).WithTopology(report.Endpoint).WithConsts(map[string]string{
					endpoint.Addr:      ServerIP,
					endpoint.Port:      NonContainerClientPort,
					process.PID:        NonContainerPID,
					report.HostNodeID:  ServerHostNodeID,
					endpoint.Procspied: True,
				}).WithAdjacent(GoogleEndpointNodeID),

				// Probe pseudo nodes
				UnknownClient1NodeID: report.MakeNode(UnknownClient1NodeID).WithTopology(report.Endpoint).WithConsts(map[string]string{
					endpoint.Addr:      UnknownClient1IP,
					endpoint.Port:      UnknownClient1Port,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(30),
					EgressByteCount:   newu64(300),
				}),

				UnknownClient2NodeID: report.MakeNode(UnknownClient2NodeID).WithTopology(report.Endpoint).WithConsts(map[string]string{
					endpoint.Addr:      UnknownClient2IP,
					endpoint.Port:      UnknownClient2Port,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(40),
					EgressByteCount:   newu64(400),
				}),

				UnknownClient3NodeID: report.MakeNode(UnknownClient3NodeID).WithTopology(report.Endpoint).WithConsts(map[string]string{
					endpoint.Addr:      UnknownClient3IP,
					endpoint.Port:      UnknownClient3Port,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(50),
					EgressByteCount:   newu64(500),
				}),

				RandomClientNodeID: report.MakeNode(RandomClientNodeID).WithTopology(report.Endpoint).WithConsts(map[string]string{
					endpoint.Addr:      RandomClientIP,
					endpoint.Port:      RandomClientPort,
					endpoint.Procspied: True,
				}).WithEdge(Server80NodeID, report.EdgeMetadata{
					EgressPacketCount: newu64(60),
					EgressByteCount:   newu64(600),
				}),

				GoogleEndpointNodeID: report.MakeNode(GoogleEndpointNodeID).WithTopology(report.Endpoint).WithConsts(map[string]string{
					endpoint.Addr:      GoogleIP,
					endpoint.Port:      GooglePort,
					endpoint.Procspied: True,
				}),
			},
		},
		Process: report.Topology{
			Nodes: report.Nodes{
				ClientProcess1NodeID: report.MakeNodeWith(ClientProcess1NodeID, map[string]string{
					process.PID:        Client1PID,
					process.Name:       Client1Name,
					docker.ContainerID: ClientContainerID,
					report.HostNodeID:  ClientHostNodeID,
				}).
					WithTopology(report.Process).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ClientHostNodeID)).
					Add("container", report.MakeStringSet(ClientContainerNodeID)),
				).WithMetrics(report.Metrics{
					process.CPUUsage:    ClientProcess1CPUMetric,
					process.MemoryUsage: ClientProcess1MemoryMetric,
				}),
				ClientProcess2NodeID: report.MakeNodeWith(ClientProcess2NodeID, map[string]string{
					process.PID:        Client2PID,
					process.Name:       Client2Name,
					docker.ContainerID: ClientContainerID,
					report.HostNodeID:  ClientHostNodeID,
				}).
					WithTopology(report.Process).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ClientHostNodeID)).
					Add("container", report.MakeStringSet(ClientContainerNodeID)),
				),
				ServerProcessNodeID: report.MakeNodeWith(ServerProcessNodeID, map[string]string{
					process.PID:        ServerPID,
					process.Name:       ServerName,
					docker.ContainerID: ServerContainerID,
					report.HostNodeID:  ServerHostNodeID,
				}).
					WithTopology(report.Process).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ServerHostNodeID)).
					Add("container", report.MakeStringSet(ServerContainerNodeID)),
				),
				NonContainerProcessNodeID: report.MakeNodeWith(NonContainerProcessNodeID, map[string]string{
					process.PID:       NonContainerPID,
					process.Name:      NonContainerName,
					report.HostNodeID: ServerHostNodeID,
				}).
					WithTopology(report.Process).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ServerHostNodeID)),
				),
			},
			MetadataTemplates: process.MetadataTemplates,
			MetricTemplates:   process.MetricTemplates,
		}.WithShape(report.Square).WithLabel("process", "processes"),
		Container: report.Topology{
			Nodes: report.Nodes{
				ClientContainerNodeID: report.MakeNodeWith(

					ClientContainerNodeID, map[string]string{
						docker.ContainerID:                           ClientContainerID,
						docker.ContainerName:                         ClientContainerName,
						docker.ContainerHostname:                     ClientContainerHostname,
						docker.ImageID:                               ClientContainerImageID,
						report.HostNodeID:                            ClientHostNodeID,
						docker.LabelPrefix + "io.kubernetes.pod.uid": ClientPodUID,
						kubernetes.Namespace:                         KubernetesNamespace,
						docker.ContainerState:                        docker.StateRunning,
						docker.ContainerStateHuman:                   docker.StateRunning,
					}).
					WithTopology(report.Container).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ClientHostNodeID)).
					Add("container_image", report.MakeStringSet(ClientContainerImageNodeID)).
					Add("pod", report.MakeStringSet(ClientPodNodeID)),
				).WithMetrics(report.Metrics{
					docker.CPUTotalUsage: ClientContainerCPUMetric,
					docker.MemoryUsage:   ClientContainerMemoryMetric,
				}),
				ServerContainerNodeID: report.MakeNodeWith(

					ServerContainerNodeID, map[string]string{
						docker.ContainerID:                                        ServerContainerID,
						docker.ContainerName:                                      ServerContainerName,
						docker.ContainerHostname:                                  ServerContainerHostname,
						docker.ContainerState:                                     docker.StateRunning,
						docker.ContainerStateHuman:                                docker.StateRunning,
						docker.ImageID:                                            ServerContainerImageID,
						report.HostNodeID:                                         ServerHostNodeID,
						docker.LabelPrefix + detailed.AmazonECSContainerNameLabel: "server",
						docker.LabelPrefix + "foo1":                               "bar1",
						docker.LabelPrefix + "foo2":                               "bar2",
						docker.LabelPrefix + "io.kubernetes.pod.uid":              ServerPodUID,
						kubernetes.Namespace:                                      KubernetesNamespace,
					}).
					WithTopology(report.Container).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ServerHostNodeID)).
					Add("container_image", report.MakeStringSet(ServerContainerImageNodeID)).
					Add("pod", report.MakeStringSet(ServerPodNodeID)),
				).WithMetrics(report.Metrics{
					docker.CPUTotalUsage: ServerContainerCPUMetric,
					docker.MemoryUsage:   ServerContainerMemoryMetric,
				}),
			},
			MetadataTemplates: docker.ContainerMetadataTemplates,
			MetricTemplates:   docker.ContainerMetricTemplates,
		}.WithShape(report.Hexagon).WithLabel("container", "containers"),
		ContainerImage: report.Topology{
			Nodes: report.Nodes{
				ClientContainerImageNodeID: report.MakeNodeWith(ClientContainerImageNodeID, map[string]string{
					docker.ImageID:    ClientContainerImageID,
					docker.ImageName:  ClientContainerImageName,
					report.HostNodeID: ClientHostNodeID,
				}).
					WithParents(report.EmptySets.
						Add("host", report.MakeStringSet(ClientHostNodeID)),
					).WithTopology(report.ContainerImage),
				ServerContainerImageNodeID: report.MakeNodeWith(ServerContainerImageNodeID, map[string]string{
					docker.ImageID:              ServerContainerImageID,
					docker.ImageName:            ServerContainerImageName,
					report.HostNodeID:           ServerHostNodeID,
					docker.LabelPrefix + "foo1": "bar1",
					docker.LabelPrefix + "foo2": "bar2",
				}).
					WithParents(report.EmptySets.
						Add("host", report.MakeStringSet(ServerHostNodeID)),
					).WithTopology(report.ContainerImage),
			},
			MetadataTemplates: docker.ContainerImageMetadataTemplates,
		}.WithShape(report.Hexagon).WithLabel("image", "images"),
		Host: report.Topology{
			Nodes: report.Nodes{
				ClientHostNodeID: report.MakeNodeWith(

					ClientHostNodeID, map[string]string{
						"host_name":       ClientHostName,
						"os":              "Linux",
						report.HostNodeID: ClientHostNodeID,
					}).
					WithTopology(report.Host).WithSets(report.EmptySets.
					Add(host.LocalNetworks, report.MakeStringSet("10.10.10.0/24")),
				).WithMetrics(report.Metrics{
					host.CPUUsage:    ClientHostCPUMetric,
					host.MemoryUsage: ClientHostMemoryMetric,
					host.Load1:       ClientHostLoad1Metric,
				}),
				ServerHostNodeID: report.MakeNodeWith(

					ServerHostNodeID, map[string]string{
						"host_name":       ServerHostName,
						"os":              "Linux",
						report.HostNodeID: ServerHostNodeID,
					}).
					WithTopology(report.Host).WithSets(report.EmptySets.
					Add(host.LocalNetworks, report.MakeStringSet("10.10.10.0/24")),
				).WithMetrics(report.Metrics{
					host.CPUUsage:    ServerHostCPUMetric,
					host.MemoryUsage: ServerHostMemoryMetric,
					host.Load1:       ServerHostLoad1Metric,
				}),
			},
			MetadataTemplates: host.MetadataTemplates,
			MetricTemplates:   host.MetricTemplates,
		}.WithShape(report.Circle).WithLabel("host", "hosts"),
		Pod: report.Topology{
			Nodes: report.Nodes{
				ClientPodNodeID: report.MakeNodeWith(
					ClientPodNodeID, map[string]string{
						kubernetes.Name:      "pong-a",
						kubernetes.Namespace: KubernetesNamespace,
						report.HostNodeID:    ClientHostNodeID,
					}).
					WithTopology(report.Pod).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ClientHostNodeID)).
					Add("service", report.MakeStringSet(ServiceNodeID)),
				),
				ServerPodNodeID: report.MakeNodeWith(
					ServerPodNodeID, map[string]string{
						kubernetes.Name:      "pong-b",
						kubernetes.Namespace: KubernetesNamespace,
						kubernetes.State:     "running",
						report.HostNodeID:    ServerHostNodeID,
					}).
					WithTopology(report.Pod).WithParents(report.EmptySets.
					Add("host", report.MakeStringSet(ServerHostNodeID)).
					Add("service", report.MakeStringSet(ServiceNodeID)),
				),
			},
			MetadataTemplates: kubernetes.PodMetadataTemplates,
		}.WithShape(report.Heptagon).WithLabel("pod", "pods"),
		Service: report.Topology{
			Nodes: report.Nodes{
				ServiceNodeID: report.MakeNodeWith(

					ServiceNodeID, map[string]string{
						kubernetes.Name:      "pongservice",
						kubernetes.Namespace: "ping",
					}).
					WithTopology(report.Service),
			},
		}.WithShape(report.Heptagon).WithLabel("service", "services"),
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
