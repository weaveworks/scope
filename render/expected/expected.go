package expected

import (
	"fmt"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

// Exported for testing.
var (
	circle   = "circle"
	square   = "square"
	heptagon = "heptagon"
	hexagon  = "hexagon"
	cloud    = "cloud"

	Client54001EndpointID    = render.MakeEndpointID(fixture.ClientHostID, fixture.ClientIP, fixture.ClientPort54001)
	Client54002EndpointID    = render.MakeEndpointID(fixture.ClientHostID, fixture.ClientIP, fixture.ClientPort54002)
	ServerEndpointID         = render.MakeEndpointID(fixture.ServerHostID, fixture.ServerIP, fixture.ServerPort)
	UnknownClient1EndpointID = render.MakeEndpointID("", fixture.UnknownClient1IP, fixture.UnknownClient1Port)
	UnknownClient2EndpointID = render.MakeEndpointID("", fixture.UnknownClient2IP, fixture.UnknownClient2Port)
	UnknownClient3EndpointID = render.MakeEndpointID("", fixture.UnknownClient3IP, fixture.UnknownClient3Port)
	RandomClientEndpointID   = render.MakeEndpointID("", fixture.RandomClientIP, fixture.RandomClientPort)
	NonContainerEndpointID   = render.MakeEndpointID(fixture.ServerHostID, fixture.ServerIP, fixture.NonContainerClientPort)
	GoogleEndpointID         = render.MakeEndpointID("", fixture.GoogleIP, fixture.GooglePort)

	RenderedEndpoints = (render.RenderableNodes{
		Client54001EndpointID: {
			ID:    Client54001EndpointID,
			Shape: circle,
			Node:  report.MakeNode().WithAdjacent(ServerEndpointID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(10),
				EgressByteCount:   newu64(100),
			},
		},
		Client54002EndpointID: {
			ID:    Client54002EndpointID,
			Shape: circle,
			Node:  report.MakeNode().WithAdjacent(ServerEndpointID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(20),
				EgressByteCount:   newu64(200),
			},
		},
		ServerEndpointID: {
			ID:    ServerEndpointID,
			Shape: circle,
			Node:  report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		UnknownClient1EndpointID: {
			ID:    UnknownClient1EndpointID,
			Shape: circle,
			Node:  report.MakeNode().WithAdjacent(ServerEndpointID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		UnknownClient2EndpointID: {
			ID:    UnknownClient2EndpointID,
			Shape: circle,
			Node:  report.MakeNode().WithAdjacent(ServerEndpointID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(40),
				EgressByteCount:   newu64(400),
			},
		},
		UnknownClient3EndpointID: {
			ID:    UnknownClient3EndpointID,
			Shape: circle,
			Node:  report.MakeNode().WithAdjacent(ServerEndpointID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(50),
				EgressByteCount:   newu64(500),
			},
		},
		RandomClientEndpointID: {
			ID:    RandomClientEndpointID,
			Shape: circle,
			Node:  report.MakeNode().WithAdjacent(ServerEndpointID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
		},
		NonContainerEndpointID: {
			ID:           NonContainerEndpointID,
			Shape:        circle,
			Node:         report.MakeNode().WithAdjacent(GoogleEndpointID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		GoogleEndpointID: {
			ID:    GoogleEndpointID,
			Shape: circle,
			Node:  report.MakeNode(),
		},
	}).Prune()

	ClientProcess1ID      = render.MakeProcessID(fixture.ClientHostID, fixture.Client1PID)
	ClientProcess2ID      = render.MakeProcessID(fixture.ClientHostID, fixture.Client2PID)
	ServerProcessID       = render.MakeProcessID(fixture.ServerHostID, fixture.ServerPID)
	nonContainerProcessID = render.MakeProcessID(fixture.ServerHostID, fixture.NonContainerPID)
	unknownPseudoNode1ID  = render.MakePseudoNodeID(fixture.UnknownClient1IP)
	unknownPseudoNode2ID  = render.MakePseudoNodeID(fixture.UnknownClient3IP)

	unknownPseudoNode1 = func(adjacent string) render.RenderableNode {
		return render.RenderableNode{
			ID:     unknownPseudoNode1ID,
			Label:  fixture.UnknownClient1IP,
			Pseudo: true,
			Shape:  circle,
			Node:   report.MakeNode().WithAdjacent(adjacent),
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[UnknownClient1EndpointID],
				RenderedEndpoints[UnknownClient2EndpointID],
			),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(70),
				EgressByteCount:   newu64(700),
			},
		}
	}
	unknownPseudoNode2 = func(adjacent string) render.RenderableNode {
		return render.RenderableNode{
			ID:     unknownPseudoNode2ID,
			Label:  fixture.UnknownClient3IP,
			Pseudo: true,
			Shape:  circle,
			Node:   report.MakeNode().WithAdjacent(adjacent),
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[UnknownClient3EndpointID],
			),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(50),
				EgressByteCount:   newu64(500),
			},
		}
	}
	theIncomingInternetNode = func(adjacent string) render.RenderableNode {
		return render.RenderableNode{
			ID:         render.IncomingInternetID,
			Label:      render.InboundMajor,
			LabelMinor: render.InboundMinor,
			Pseudo:     true,
			Shape:      cloud,
			Node:       report.MakeNode().WithAdjacent(adjacent),
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[RandomClientEndpointID],
			),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
		}
	}
	theOutgoingInternetNode = render.RenderableNode{
		ID:           render.OutgoingInternetID,
		Label:        render.OutboundMajor,
		LabelMinor:   render.OutboundMinor,
		Pseudo:       true,
		Shape:        cloud,
		Node:         report.MakeNode(),
		EdgeMetadata: report.EdgeMetadata{},
		Children: render.MakeRenderableNodeSet(
			RenderedEndpoints[GoogleEndpointID],
		),
	}

	RenderedProcesses = (render.RenderableNodes{
		ClientProcess1ID: {
			ID:         ClientProcess1ID,
			Label:      fixture.Client1Name,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ClientHostID, fixture.Client1PID),
			Rank:       fixture.Client1Name,
			Shape:      square,
			Node:       report.MakeNode().WithAdjacent(ServerProcessID),
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[Client54001EndpointID],
			),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(10),
				EgressByteCount:   newu64(100),
			},
		},
		ClientProcess2ID: {
			ID:         ClientProcess2ID,
			Label:      fixture.Client2Name,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ClientHostID, fixture.Client2PID),
			Rank:       fixture.Client2Name,
			Shape:      square,
			Node:       report.MakeNode().WithAdjacent(ServerProcessID),
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[Client54002EndpointID],
			),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(20),
				EgressByteCount:   newu64(200),
			},
		},
		ServerProcessID: {
			ID:         ServerProcessID,
			Label:      fixture.ServerName,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ServerHostID, fixture.ServerPID),
			Rank:       fixture.ServerName,
			Shape:      square,
			Node:       report.MakeNode(),
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[ServerEndpointID],
			),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		nonContainerProcessID: {
			ID:         nonContainerProcessID,
			Label:      fixture.NonContainerName,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ServerHostID, fixture.NonContainerPID),
			Rank:       fixture.NonContainerName,
			Shape:      square,
			Node:       report.MakeNode().WithAdjacent(render.OutgoingInternetID),
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[NonContainerEndpointID],
			),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1ID:      unknownPseudoNode1(ServerProcessID),
		unknownPseudoNode2ID:      unknownPseudoNode2(ServerProcessID),
		render.IncomingInternetID: theIncomingInternetNode(ServerProcessID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}).Prune()

	RenderedProcessNames = (render.RenderableNodes{
		fixture.Client1Name: {
			ID:         fixture.Client1Name,
			Label:      fixture.Client1Name,
			LabelMinor: "2 processes",
			Rank:       fixture.Client1Name,
			Shape:      square,
			Stack:      true,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[Client54001EndpointID],
				RenderedEndpoints[Client54002EndpointID],
				RenderedProcesses[ClientProcess1ID],
				RenderedProcesses[ClientProcess2ID],
			),
			Node: report.MakeNode().WithAdjacent(fixture.ServerName),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		fixture.ServerName: {
			ID:         fixture.ServerName,
			Label:      fixture.ServerName,
			LabelMinor: "1 process",
			Rank:       fixture.ServerName,
			Shape:      square,
			Stack:      true,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[ServerEndpointID],
				RenderedProcesses[ServerProcessID],
			),
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		fixture.NonContainerName: {
			ID:         fixture.NonContainerName,
			Label:      fixture.NonContainerName,
			LabelMinor: "1 process",
			Rank:       fixture.NonContainerName,
			Shape:      square,
			Stack:      true,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[NonContainerEndpointID],
				RenderedProcesses[nonContainerProcessID],
			),
			Node:         report.MakeNode().WithAdjacent(render.OutgoingInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1ID:      unknownPseudoNode1(fixture.ServerName),
		unknownPseudoNode2ID:      unknownPseudoNode2(fixture.ServerName),
		render.IncomingInternetID: theIncomingInternetNode(fixture.ServerName),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}).Prune()

	ClientContainerID   = render.MakeContainerID(fixture.ClientContainerID)
	ServerContainerID   = render.MakeContainerID(fixture.ServerContainerID)
	uncontainedServerID = render.MakePseudoNodeID(render.UncontainedID, fixture.ServerHostID)

	RenderedContainers = (render.RenderableNodes{
		ClientContainerID: {
			ID:         ClientContainerID,
			Label:      "client",
			LabelMinor: fixture.ClientHostName,
			Shape:      hexagon,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[Client54001EndpointID],
				RenderedEndpoints[Client54002EndpointID],
				RenderedProcesses[ClientProcess1ID],
				RenderedProcesses[ClientProcess2ID],
			),
			Node: report.MakeNode().WithAdjacent(ServerContainerID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
			ControlNode: fixture.ClientContainerNodeID,
		},
		ServerContainerID: {
			ID:         ServerContainerID,
			Label:      "server",
			LabelMinor: fixture.ServerHostName,
			Shape:      hexagon,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[ServerEndpointID],
				RenderedProcesses[ServerProcessID],
			),
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
			ControlNode: fixture.ServerContainerNodeID,
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			Label:      render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Shape:      square,
			Stack:      true,
			Pseudo:     true,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[NonContainerEndpointID],
				RenderedProcesses[nonContainerProcessID],
			),
			Node:         report.MakeNode().WithAdjacent(render.OutgoingInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		// unknownPseudoNode1ID:      unknownPseudoNode1(ServerContainerID),
		// unknownPseudoNode2ID:      unknownPseudoNode2(ServerContainerID),
		render.IncomingInternetID: theIncomingInternetNode(ServerContainerID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}).Prune()

	ClientContainerImageID = render.MakeContainerImageID(fixture.ClientContainerImageName)
	ServerContainerImageID = render.MakeContainerImageID(fixture.ServerContainerImageName)

	RenderedContainerImages = (render.RenderableNodes{
		ClientContainerImageID: {
			ID:         ClientContainerImageID,
			Label:      fixture.ClientContainerImageName,
			LabelMinor: "1 container",
			Rank:       fixture.ClientContainerImageName,
			Shape:      hexagon,
			Stack:      true,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[Client54001EndpointID],
				RenderedEndpoints[Client54002EndpointID],
				RenderedProcesses[ClientProcess1ID],
				RenderedProcesses[ClientProcess2ID],
				RenderedContainers[ClientContainerID],
			),
			Node: report.MakeNode().WithAdjacent(ServerContainerImageID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		ServerContainerImageID: {
			ID:         ServerContainerImageID,
			Label:      fixture.ServerContainerImageName,
			LabelMinor: "1 container",
			Rank:       fixture.ServerContainerImageName,
			Shape:      hexagon,
			Stack:      true,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[ServerEndpointID],
				RenderedProcesses[ServerProcessID],
				RenderedContainers[ServerContainerID],
			),
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			Label:      render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Shape:      square,
			Stack:      true,
			Pseudo:     true,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[NonContainerEndpointID],
				RenderedProcesses[nonContainerProcessID],
			),
			Node:         report.MakeNode().WithAdjacent(render.OutgoingInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		// unknownPseudoNode1ID:      unknownPseudoNode1(ServerContainerImageID),
		// unknownPseudoNode2ID:      unknownPseudoNode2(ServerContainerImageID),
		render.IncomingInternetID: theIncomingInternetNode(ServerContainerImageID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}).Prune()

	ClientPodRenderedID = render.MakePodID("ping/pong-a")
	ServerPodRenderedID = render.MakePodID("ping/pong-b")

	RenderedPods = (render.RenderableNodes{
		ClientPodRenderedID: {
			ID:         ClientPodRenderedID,
			Label:      "pong-a",
			LabelMinor: "1 container",
			Rank:       "ping/pong-a",
			Shape:      heptagon,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[Client54001EndpointID],
				RenderedEndpoints[Client54002EndpointID],
				RenderedProcesses[ClientProcess1ID],
				RenderedProcesses[ClientProcess2ID],
				RenderedContainers[ClientContainerID],
			),
			Node: report.MakeNode().WithAdjacent(ServerPodRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		ServerPodRenderedID: {
			ID:         ServerPodRenderedID,
			Label:      "pong-b",
			LabelMinor: "1 container",
			Rank:       "ping/pong-b",
			Shape:      heptagon,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[ServerEndpointID],
				RenderedProcesses[ServerProcessID],
				RenderedContainers[ServerContainerID],
			),
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			Label:      render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Pseudo:     true,
			Shape:      square,
			Stack:      true,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[NonContainerEndpointID],
				RenderedProcesses[nonContainerProcessID],
			),
			Node:         report.MakeNode().WithAdjacent(render.OutgoingInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		// unknownPseudoNode1ID:      unknownPseudoNode1(ServerPodRenderedID),
		// unknownPseudoNode2ID:      unknownPseudoNode2(ServerPodRenderedID),
		render.IncomingInternetID: theIncomingInternetNode(ServerPodRenderedID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}).Prune()

	ServerHostID = render.MakeHostID(fixture.ServerHostID)
	ClientHostID = render.MakeHostID(fixture.ClientHostID)

	RenderedHosts = (render.RenderableNodes{
		ClientHostID: {
			ID:         ClientHostID,
			Label:      "client",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Shape:      circle,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[Client54001EndpointID],
				RenderedEndpoints[Client54002EndpointID],
				RenderedProcesses[ClientProcess1ID],
				RenderedProcesses[ClientProcess2ID],
				RenderedContainers[ClientContainerID],
				RenderedContainerImages[ClientContainerImageID],
				//RenderedPods[ClientPodRenderedID], #1142
			),
			Node: report.MakeNode().WithAdjacent(ServerHostID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		ServerHostID: {
			ID:         ServerHostID,
			Label:      "server",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Shape:      circle,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[ServerEndpointID],
				RenderedEndpoints[NonContainerEndpointID],
				RenderedProcesses[ServerProcessID],
				RenderedProcesses[nonContainerProcessID],
				RenderedContainers[ServerContainerID],
				RenderedContainerImages[ServerContainerImageID],
				//RenderedPods[ServerPodRenderedID], #1142
			),
			Node: report.MakeNode().WithAdjacent(render.OutgoingInternetID),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		unknownPseudoNode1ID:      unknownPseudoNode1(ServerHostID),
		unknownPseudoNode2ID:      unknownPseudoNode2(ServerHostID),
		render.IncomingInternetID: theIncomingInternetNode(ServerHostID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}).Prune()

	ServiceRenderedID = render.MakeServiceID("ping/pongservice")

	RenderedPodServices = (render.RenderableNodes{
		ServiceRenderedID: {
			ID:         ServiceRenderedID,
			Label:      "pongservice",
			LabelMinor: "2 pods",
			Rank:       fixture.ServiceID,
			Shape:      heptagon,
			Stack:      true,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[Client54001EndpointID],
				RenderedEndpoints[Client54002EndpointID],
				RenderedEndpoints[ServerEndpointID],
				RenderedProcesses[ClientProcess1ID],
				RenderedProcesses[ClientProcess2ID],
				RenderedProcesses[ServerProcessID],
				RenderedContainers[ClientContainerID],
				RenderedContainers[ServerContainerID],
				RenderedPods[ClientPodRenderedID],
				RenderedPods[ServerPodRenderedID],
			),
			Node: report.MakeNode().WithAdjacent(ServiceRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount:  newu64(30),
				EgressByteCount:    newu64(300),
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			Label:      render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Pseudo:     true,
			Shape:      square,
			Stack:      true,
			Children: render.MakeRenderableNodeSet(
				RenderedEndpoints[NonContainerEndpointID],
				RenderedProcesses[nonContainerProcessID],
			),
			Node:         report.MakeNode().WithAdjacent(render.OutgoingInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		// unknownPseudoNode1ID:      unknownPseudoNode1(ServiceRenderedID),
		// unknownPseudoNode2ID:      unknownPseudoNode2(ServiceRenderedID),
		render.IncomingInternetID: theIncomingInternetNode(ServiceRenderedID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}).Prune()
)

func newu64(value uint64) *uint64 { return &value }
