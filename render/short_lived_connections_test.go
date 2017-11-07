package render_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/common/mtime"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/utils"
)

var (
	serverHostID     = "host1"
	serverHostNodeID = report.MakeHostNodeID(serverHostID)

	randomIP             = "3.4.5.6"
	randomPort           = "56789"
	randomEndpointNodeID = report.MakeEndpointNodeID(serverHostID, "", randomIP, randomPort)

	serverIP             = "192.168.1.1"
	serverPort           = "80"
	serverEndpointNodeID = report.MakeEndpointNodeID(serverHostID, "", serverIP, serverPort)

	container1ID     = "11b2c3d4e5"
	container1IP     = "192.168.0.1"
	container1Name   = "foo"
	container1NodeID = report.MakeContainerNodeID(container1ID)

	container1Port           = "16782"
	container1EndpointNodeID = report.MakeEndpointNodeID(serverHostID, "", container1IP, container1Port)

	duplicatedIP             = "192.168.0.2"
	duplicatedPort           = "80"
	duplicatedEndpointNodeID = report.MakeEndpointNodeID(serverHostID, "", duplicatedIP, duplicatedPort)

	container2ID     = "21b2c3d4e5"
	container2IP     = duplicatedIP
	container2Name   = "bar"
	container2NodeID = report.MakeContainerNodeID(container2ID)

	pauseContainerID     = "31b2c3d4e5"
	pauseContainerIP     = duplicatedIP
	pauseContainerName   = "POD"
	pauseContainerNodeID = report.MakeContainerNodeID(pauseContainerID)

	rpt = report.Report{
		Endpoint: report.Topology{
			Nodes: report.Nodes{
				randomEndpointNodeID: report.MakeNode(randomEndpointNodeID).
					WithTopology(report.Endpoint).WithAdjacent(serverEndpointNodeID),

				serverEndpointNodeID: report.MakeNode(serverEndpointNodeID).
					WithTopology(report.Endpoint),

				container1EndpointNodeID: report.MakeNode(container1EndpointNodeID).
					WithTopology(report.Endpoint).WithAdjacent(duplicatedEndpointNodeID),

				duplicatedEndpointNodeID: report.MakeNode(duplicatedEndpointNodeID).
					WithTopology(report.Endpoint),
			},
		},
		Container: report.Topology{
			Nodes: report.Nodes{
				container1NodeID: report.MakeNodeWith(container1NodeID, map[string]string{
					docker.ContainerID:   container1ID,
					docker.ContainerName: container1Name,
					report.HostNodeID:    serverHostNodeID,
				}).
					WithSets(report.MakeSets().
						Add(docker.ContainerIPs, report.MakeStringSet(container1IP)).
						Add(docker.ContainerIPsWithScopes, report.MakeStringSet(report.MakeAddressNodeID("", container1IP))).
						Add(docker.ContainerPorts, report.MakeStringSet(fmt.Sprintf("%s:%s->%s/tcp", serverIP, serverPort, serverPort))),
					).WithTopology(report.Container),
				container2NodeID: report.MakeNodeWith(container2NodeID, map[string]string{
					docker.ContainerID:   container2ID,
					docker.ContainerName: container2Name,
					report.HostNodeID:    serverHostNodeID,
				}).
					WithSets(report.MakeSets().
						Add(docker.ContainerIPs, report.MakeStringSet(container2IP)).
						Add(docker.ContainerIPsWithScopes, report.MakeStringSet(report.MakeAddressNodeID("", container2IP))),
					).WithTopology(report.Container),
				pauseContainerNodeID: report.MakeNodeWith(pauseContainerNodeID, map[string]string{
					docker.ContainerID:   pauseContainerID,
					docker.ContainerName: pauseContainerName,
					report.HostNodeID:    serverHostNodeID,
				}).
					WithSets(report.MakeSets().
						Add(docker.ContainerIPs, report.MakeStringSet(pauseContainerIP)).
						Add(docker.ContainerIPsWithScopes, report.MakeStringSet(report.MakeAddressNodeID("", pauseContainerIP))),
					).WithTopology(report.Container).WithLatest(report.DoesNotMakeConnections, mtime.Now(), ""),
			},
		},
		Host: report.Topology{
			Nodes: report.Nodes{
				serverHostNodeID: report.MakeNodeWith(serverHostNodeID, map[string]string{
					report.HostNodeID: serverHostNodeID,
				}).
					WithSets(report.MakeSets().
						Add(host.LocalNetworks, report.MakeStringSet("192.168.0.0/16")),
					).WithTopology(report.Host),
			},
		},
	}
)

func TestShortLivedInternetNodeConnections(t *testing.T) {
	have := utils.Prune(render.ContainerWithImageNameRenderer.Render(rpt, FilterNoop).Nodes)

	// Conntracked-only connections from the internet should be assigned to the internet pseudonode
	internet, ok := have[render.IncomingInternetID]
	if !ok {
		t.Fatal("Expected output to have an incoming internet node")
	}

	if !internet.Adjacency.Contains(container1NodeID) {
		t.Errorf("Expected internet node to have adjacency to %s, but only had %v", container1NodeID, internet.Adjacency)
	}
}

func TestPauseContainerDiscarded(t *testing.T) {
	have := utils.Prune(render.ContainerWithImageNameRenderer.Render(rpt, FilterNoop).Nodes)
	// There should only be a connection from container1 and the destination should be container2
	container1, ok := have[container1NodeID]
	if !ok {
		t.Fatal("Expected output to have container1")
	}

	if len(container1.Adjacency) != 1 || !container1.Adjacency.Contains(container2NodeID) {
		t.Errorf("Expected container1 to have a unique adjacency to %s, but instead had %v", container2NodeID, container1.Adjacency)
	}

}
