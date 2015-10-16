package render_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
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
				randomEndpointNodeID: report.MakeNode().WithLatests(map[string]string{
					endpoint.Addr:        randomIP,
					endpoint.Port:        randomPort,
					endpoint.Conntracked: "true",
				}).WithAdjacent(serverEndpointNodeID).WithID(randomEndpointNodeID).WithTopology(report.Endpoint),

				serverEndpointNodeID: report.MakeNode().WithLatests(map[string]string{
					endpoint.Addr:        serverIP,
					endpoint.Port:        serverPort,
					endpoint.Conntracked: "true",
				}).WithID(serverEndpointNodeID).WithTopology(report.Endpoint),
			},
		},
		Container: report.Topology{
			Nodes: report.Nodes{
				containerNodeID: report.MakeNode().WithLatests(map[string]string{
					docker.ContainerID:   containerID,
					docker.ContainerName: containerName,
					report.HostNodeID:    serverHostNodeID,
				}).WithSets(report.EmptySets.
					Add(docker.ContainerIPs, report.MakeStringSet(containerIP)).
					Add(docker.ContainerPorts, report.MakeStringSet(fmt.Sprintf("%s:%s->%s/tcp", serverIP, serverPort, serverPort))),
				).WithID(containerNodeID).WithTopology(report.Container),
			},
		},
		Host: report.Topology{
			Nodes: report.Nodes{
				serverHostNodeID: report.MakeNodeWith(map[string]string{
					report.HostNodeID: serverHostNodeID,
				}).WithSets(report.EmptySets.
					Add(host.LocalNetworks, report.MakeStringSet("192.168.0.0/16")),
				).WithID(serverHostNodeID).WithTopology(report.Host),
			},
		},
	}

	want = (render.RenderableNodes{
		render.IncomingInternetID: {
			ID:         render.IncomingInternetID,
			LabelMajor: render.InboundMajor,
			LabelMinor: render.RequestsMinor,
			Pseudo:     true,
			Shape:      "cloud",
			Node:       report.MakeNode().WithAdjacent(render.MakeContainerID(containerID)),
		},
		render.MakeContainerID(containerID): {
			ID:          render.MakeContainerID(containerID),
			LabelMajor:  containerName,
			LabelMinor:  serverHostID,
			Rank:        "",
			Pseudo:      false,
			Shape:       "hexagon",
			Node:        report.MakeNode(),
			ControlNode: containerNodeID,
		},
	}).Prune()
)

func TestShortLivedInternetNodeConnections(t *testing.T) {
	have := (render.ContainerWithImageNameRenderer.Render(rpt)).Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
