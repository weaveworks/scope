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
	pentagon = "pentagon"
	hexagon  = "hexagon"
	cloud    = "cloud"

	uncontainedServerID  = render.MakePseudoNodeID(render.UncontainedID, fixture.ServerHostName)
	unknownPseudoNode1ID = render.MakePseudoNodeID("10.10.10.10", fixture.ServerIP, "80")
	unknownPseudoNode2ID = render.MakePseudoNodeID("10.10.10.11", fixture.ServerIP, "80")
	unknownPseudoNode1   = func(adjacent string) render.RenderableNode {
		return render.RenderableNode{
			ID:         unknownPseudoNode1ID,
			LabelMajor: "10.10.10.10",
			Pseudo:     true,
			Shape:      circle,
			Node:       report.MakeNode().WithAdjacent(adjacent),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(70),
				EgressByteCount:   newu64(700),
			},
		}
	}
	unknownPseudoNode2 = func(adjacent string) render.RenderableNode {
		return render.RenderableNode{
			ID:         unknownPseudoNode2ID,
			LabelMajor: "10.10.10.11",
			Pseudo:     true,
			Shape:      circle,
			Node:       report.MakeNode().WithAdjacent(adjacent),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(50),
				EgressByteCount:   newu64(500),
			},
		}
	}
	theInternetNode = func(adjacent string) render.RenderableNode {
		return render.RenderableNode{
			ID:         render.TheInternetID,
			LabelMajor: render.TheInternetMajor,
			Pseudo:     true,
			Shape:      cloud,
			Node:       report.MakeNode().WithAdjacent(adjacent),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
		}
	}
	ClientProcess1ID      = render.MakeProcessID(fixture.ClientHostID, fixture.Client1PID)
	ClientProcess2ID      = render.MakeProcessID(fixture.ClientHostID, fixture.Client2PID)
	ServerProcessID       = render.MakeProcessID(fixture.ServerHostID, fixture.ServerPID)
	nonContainerProcessID = render.MakeProcessID(fixture.ServerHostID, fixture.NonContainerPID)

	RenderedProcesses = (render.RenderableNodes{
		ClientProcess1ID: {
			ID:         ClientProcess1ID,
			LabelMajor: fixture.Client1Name,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ClientHostID, fixture.Client1PID),
			Rank:       fixture.Client1Name,
			Pseudo:     false,
			Shape:      square,
			Node:       report.MakeNode().WithAdjacent(ServerProcessID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(10),
				EgressByteCount:   newu64(100),
			},
		},
		ClientProcess2ID: {
			ID:         ClientProcess2ID,
			LabelMajor: fixture.Client2Name,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ClientHostID, fixture.Client2PID),
			Rank:       fixture.Client2Name,
			Pseudo:     false,
			Shape:      square,
			Node:       report.MakeNode().WithAdjacent(ServerProcessID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(20),
				EgressByteCount:   newu64(200),
			},
		},
		ServerProcessID: {
			ID:         ServerProcessID,
			LabelMajor: fixture.ServerName,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ServerHostID, fixture.ServerPID),
			Rank:       fixture.ServerName,
			Pseudo:     false,
			Shape:      square,
			Node:       report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		nonContainerProcessID: {
			ID:           nonContainerProcessID,
			LabelMajor:   fixture.NonContainerName,
			LabelMinor:   fmt.Sprintf("%s (%s)", fixture.ServerHostID, fixture.NonContainerPID),
			Rank:         fixture.NonContainerName,
			Pseudo:       false,
			Shape:        square,
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1ID: unknownPseudoNode1(ServerProcessID),
		unknownPseudoNode2ID: unknownPseudoNode2(ServerProcessID),
		render.TheInternetID: theInternetNode(ServerProcessID),
	}).Prune()

	ServerProcessRenderedID  = render.MakeProcessID(fixture.ServerHostID, fixture.ServerPID)
	ClientProcess1RenderedID = render.MakeProcessID(fixture.ClientHostID, fixture.Client1PID)
	ClientProcess2RenderedID = render.MakeProcessID(fixture.ClientHostID, fixture.Client2PID)

	RenderedProcessNames = (render.RenderableNodes{
		fixture.Client1Name: {
			ID:         fixture.Client1Name,
			LabelMajor: fixture.Client1Name,
			LabelMinor: "2 processes",
			Rank:       fixture.Client1Name,
			Pseudo:     false,
			Shape:      square,
			Stack:      true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
			),
			Node: report.MakeNode().WithAdjacent(fixture.ServerName),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		fixture.ServerName: {
			ID:         fixture.ServerName,
			LabelMajor: fixture.ServerName,
			LabelMinor: "1 process",
			Rank:       fixture.ServerName,
			Pseudo:     false,
			Shape:      square,
			Stack:      true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
			),
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		fixture.NonContainerName: {
			ID:         fixture.NonContainerName,
			LabelMajor: fixture.NonContainerName,
			LabelMinor: "1 process",
			Rank:       fixture.NonContainerName,
			Pseudo:     false,
			Shape:      square,
			Stack:      true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			),
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1ID: unknownPseudoNode1(fixture.ServerName),
		unknownPseudoNode2ID: unknownPseudoNode2(fixture.ServerName),
		render.TheInternetID: theInternetNode(fixture.ServerName),
	}).Prune()

	ServerContainerRenderedID = render.MakeContainerID(fixture.ServerContainerID)
	ClientContainerRenderedID = render.MakeContainerID(fixture.ClientContainerID)

	RenderedContainers = (render.RenderableNodes{
		ClientContainerRenderedID: {
			ID:         ClientContainerRenderedID,
			LabelMajor: "client",
			LabelMinor: fixture.ClientHostName,
			Rank:       fixture.ClientContainerImageName,
			Pseudo:     false,
			Shape:      hexagon,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
			),
			Node: report.MakeNode().WithAdjacent(ServerContainerRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
			ControlNode: fixture.ClientContainerNodeID,
		},
		ServerContainerRenderedID: {
			ID:         ServerContainerRenderedID,
			LabelMajor: "server",
			LabelMinor: fixture.ServerHostName,
			Rank:       fixture.ServerContainerImageName,
			Pseudo:     false,
			Shape:      hexagon,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
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
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Shape:      square,
			Stack:      true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			),
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: theInternetNode(ServerContainerRenderedID),
	}).Prune()

	ClientContainerImageRenderedName = render.MakeContainerImageID(fixture.ClientContainerImageName)
	ServerContainerImageRenderedName = render.MakeContainerImageID(fixture.ServerContainerImageName)

	RenderedContainerImages = (render.RenderableNodes{
		ClientContainerImageRenderedName: {
			ID:         ClientContainerImageRenderedName,
			LabelMajor: fixture.ClientContainerImageName,
			LabelMinor: "1 container",
			Rank:       fixture.ClientContainerImageName,
			Pseudo:     false,
			Shape:      hexagon,
			Stack:      true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
				fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
			),
			Node: report.MakeNode().WithAdjacent(ServerContainerImageRenderedName),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		ServerContainerImageRenderedName: {
			ID:         ServerContainerImageRenderedName,
			LabelMajor: fixture.ServerContainerImageName,
			LabelMinor: "1 container",
			Rank:       fixture.ServerContainerImageName,
			Pseudo:     false,
			Shape:      hexagon,
			Stack:      true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
				fixture.Report.Container.Nodes[fixture.ServerContainerNodeID],
			),
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Shape:      square,
			Stack:      true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			),
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: theInternetNode(ServerContainerImageRenderedName),
	}).Prune()

	ServerHostRenderedID = render.MakeHostID(fixture.ServerHostID)
	ClientHostRenderedID = render.MakeHostID(fixture.ClientHostID)
	pseudoHostID1        = render.MakePseudoNodeID(fixture.UnknownClient1IP, fixture.ServerIP)
	pseudoHostID2        = render.MakePseudoNodeID(fixture.UnknownClient3IP, fixture.ServerIP)

	RenderedHosts = (render.RenderableNodes{
		ServerHostRenderedID: {
			ID:         ServerHostRenderedID,
			LabelMajor: "server",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
			Shape:      circle,
			Children: report.MakeNodeSet(
				fixture.Report.Container.Nodes[fixture.ServerContainerNodeID],
				fixture.Report.Container.Nodes[fixture.ServerProcessNodeID],
			),
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		ClientHostRenderedID: {
			ID:         ClientHostRenderedID,
			LabelMajor: "client",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
			Shape:      circle,
			Children: report.MakeNodeSet(
				fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
				fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
			),
			Node: report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		pseudoHostID1: {
			ID:           pseudoHostID1,
			LabelMajor:   fixture.UnknownClient1IP,
			Pseudo:       true,
			Shape:        circle,
			Node:         report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
			Children: report.MakeNodeSet(
				fixture.Report.Container.Nodes[fixture.ServerContainerNodeID],
				fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
			),
		},
		pseudoHostID2: {
			ID:           pseudoHostID2,
			LabelMajor:   fixture.UnknownClient3IP,
			Pseudo:       true,
			Shape:        circle,
			Node:         report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: {
			ID:           render.TheInternetID,
			LabelMajor:   render.TheInternetMajor,
			Pseudo:       true,
			Shape:        cloud,
			Node:         report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
		},
	}).Prune()

	ClientPodRenderedID = render.MakePodID("ping/pong-a")
	ServerPodRenderedID = render.MakePodID("ping/pong-b")

	RenderedPods = (render.RenderableNodes{
		ClientPodRenderedID: {
			ID:         ClientPodRenderedID,
			LabelMajor: "pong-a",
			LabelMinor: "1 container",
			Rank:       "ping/pong-a",
			Pseudo:     false,
			Shape:      pentagon,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
				fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
				fixture.Report.ContainerImage.Nodes[fixture.ClientContainerImageNodeID],
				fixture.Report.Pod.Nodes[fixture.ClientPodNodeID],
			),
			Node: report.MakeNode().WithAdjacent(ServerPodRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		ServerPodRenderedID: {
			ID:         ServerPodRenderedID,
			LabelMajor: "pong-b",
			LabelMinor: "1 container",
			Rank:       "ping/pong-b",
			Pseudo:     false,
			Shape:      pentagon,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
				fixture.Report.Container.Nodes[fixture.ServerContainerNodeID],
				fixture.Report.ContainerImage.Nodes[fixture.ServerContainerImageNodeID],
				fixture.Report.Pod.Nodes[fixture.ServerPodNodeID],
			),
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Shape:      square,
			Stack:      true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			),
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: {
			ID:         render.TheInternetID,
			LabelMajor: render.TheInternetMajor,
			Pseudo:     true,
			Shape:      cloud,
			Node:       report.MakeNode().WithAdjacent(ServerPodRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
		},
	}).Prune()

	ServiceRenderedID = render.MakeServiceID("ping/pongservice")

	RenderedPodServices = (render.RenderableNodes{
		ServiceRenderedID: {
			ID:         ServiceRenderedID,
			LabelMajor: "pongservice",
			LabelMinor: "2 pods",
			Rank:       fixture.ServiceID,
			Pseudo:     false,
			Shape:      pentagon,
			Stack:      true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
				fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
				fixture.Report.ContainerImage.Nodes[fixture.ClientContainerImageNodeID],
				fixture.Report.Pod.Nodes[fixture.ClientPodNodeID],
				fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
				fixture.Report.Container.Nodes[fixture.ServerContainerNodeID],
				fixture.Report.ContainerImage.Nodes[fixture.ServerContainerImageNodeID],
				fixture.Report.Pod.Nodes[fixture.ServerPodNodeID],
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
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Shape:      square,
			Stack:      true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			),
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: {
			ID:         render.TheInternetID,
			LabelMajor: render.TheInternetMajor,
			Pseudo:     true,
			Shape:      cloud,
			Node:       report.MakeNode().WithAdjacent(ServiceRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
		},
	}).Prune()
)

func newu64(value uint64) *uint64 { return &value }
