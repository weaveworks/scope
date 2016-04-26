package render_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

var (
	serverHostID     = "host1"
	serverHostNodeID = report.MakeHostNodeID(serverHostID)

	randomIP             = "3.4.5.6"
	randomPort           = "56789"
	randomEndpointNodeID = report.MakeEndpointNodeID(serverHostID, randomIP, randomPort)

	serverIP             = "192.168.1.1"
	serverPort           = "80"
	serverEndpointNodeID = report.MakeEndpointNodeID(serverHostID, serverIP, serverPort)

	containerID     = "a1b2c3d4e5"
	containerIP     = "192.168.0.1"
	containerName   = "foo"
	containerNodeID = report.MakeContainerNodeID(containerID)

	rpt = report.Report{
		Endpoint: report.Topology{
			Nodes: report.Nodes{
				randomEndpointNodeID: report.MakeNodeWith(randomEndpointNodeID, map[string]string{
					endpoint.Addr:        randomIP,
					endpoint.Port:        randomPort,
					endpoint.Conntracked: "true",
				}).
					WithAdjacent(serverEndpointNodeID).WithTopology(report.Endpoint),

				serverEndpointNodeID: report.MakeNodeWith(serverEndpointNodeID, map[string]string{
					endpoint.Addr:        serverIP,
					endpoint.Port:        serverPort,
					endpoint.Conntracked: "true",
				}).
					WithTopology(report.Endpoint),
			},
		},
		Container: report.Topology{
			Nodes: report.Nodes{
				containerNodeID: report.MakeNodeWith(containerNodeID, map[string]string{
					docker.ContainerID:   containerID,
					docker.ContainerName: containerName,
					report.HostNodeID:    serverHostNodeID,
				}).
					WithSets(report.EmptySets.
						Add(docker.ContainerIPs, report.MakeStringSet(containerIP)).
						Add(docker.ContainerPorts, report.MakeStringSet(fmt.Sprintf("%s:%s->%s/tcp", serverIP, serverPort, serverPort))),
					).WithTopology(report.Container),
			},
		},
		Host: report.Topology{
			Nodes: report.Nodes{
				serverHostNodeID: report.MakeNodeWith(serverHostNodeID, map[string]string{
					report.HostNodeID: serverHostNodeID,
				}).
					WithSets(report.EmptySets.
						Add(host.LocalNetworks, report.MakeStringSet("192.168.0.0/16")),
					).WithTopology(report.Host),
			},
		},
	}
)

func TestShortLivedInternetNodeConnections(t *testing.T) {
	have := Prune(render.ContainerWithImageNameRenderer.Render(rpt, render.FilterNoop))

	// Conntracked-only connections from the internet should be assigned to the internet pseudonode
	internet, ok := have[render.IncomingInternetID]
	if !ok {
		t.Fatal("Expected output to have an incoming internet node")
	}

	if !internet.Adjacency.Contains(report.MakeContainerNodeID(containerID)) {
		t.Errorf("Expected internet node to have adjacency to %s, but only had %v", report.MakeContainerNodeID(containerID), internet.Adjacency)
	}
}
