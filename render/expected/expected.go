package expected

import (
	"fmt"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

// Exported for testing.
var (
	uncontainedServerID  = render.MakePseudoNodeID(render.UncontainedID, fixture.ServerHostName)
	unknownPseudoNode1ID = render.MakePseudoNodeID("10.10.10.10", fixture.ServerIP, "80")
	unknownPseudoNode2ID = render.MakePseudoNodeID("10.10.10.11", fixture.ServerIP, "80")
	unknownPseudoNode1   = func(adjacent string) render.RenderableNode {
		return render.RenderableNode{
			ID:         unknownPseudoNode1ID,
			LabelMajor: "10.10.10.10",
			Pseudo:     true,
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

	RenderedProcesses = (render.MakeRenderableNodes(
		render.RenderableNode{
			ID:         ClientProcess1ID,
			LabelMajor: fixture.Client1Name,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ClientHostID, fixture.Client1PID),
			Rank:       fixture.Client1Name,
			Pseudo:     false,
			Node:       report.MakeNode().WithAdjacent(ServerProcessID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(10),
				EgressByteCount:   newu64(100),
			},
		},
		render.RenderableNode{
			ID:         ClientProcess2ID,
			LabelMajor: fixture.Client2Name,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ClientHostID, fixture.Client2PID),
			Rank:       fixture.Client2Name,
			Pseudo:     false,
			Node:       report.MakeNode().WithAdjacent(ServerProcessID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(20),
				EgressByteCount:   newu64(200),
			},
		},
		render.RenderableNode{
			ID:         ServerProcessID,
			LabelMajor: fixture.ServerName,
			LabelMinor: fmt.Sprintf("%s (%s)", fixture.ServerHostID, fixture.ServerPID),
			Rank:       fixture.ServerName,
			Pseudo:     false,
			Node:       report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		render.RenderableNode{
			ID:           nonContainerProcessID,
			LabelMajor:   fixture.NonContainerName,
			LabelMinor:   fmt.Sprintf("%s (%s)", fixture.ServerHostID, fixture.NonContainerPID),
			Rank:         fixture.NonContainerName,
			Pseudo:       false,
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1(ServerProcessID),
		unknownPseudoNode2(ServerProcessID),
		theInternetNode(ServerProcessID),
	)).Prune()

	ServerProcessRenderedID  = render.MakeProcessID(fixture.ServerHostID, fixture.ServerPID)
	ClientProcess1RenderedID = render.MakeProcessID(fixture.ClientHostID, fixture.Client1PID)
	ClientProcess2RenderedID = render.MakeProcessID(fixture.ClientHostID, fixture.Client2PID)

	RenderedProcessNames = (render.MakeRenderableNodes(
		render.RenderableNode{
			ID:         fixture.Client1Name,
			LabelMajor: fixture.Client1Name,
			LabelMinor: "2 processes",
			Rank:       fixture.Client1Name,
			Pseudo:     false,
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
		render.RenderableNode{
			ID:         fixture.ServerName,
			LabelMajor: fixture.ServerName,
			LabelMinor: "1 process",
			Rank:       fixture.ServerName,
			Pseudo:     false,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
			),
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
			},
		},
		render.RenderableNode{
			ID:         fixture.NonContainerName,
			LabelMajor: fixture.NonContainerName,
			LabelMinor: "1 process",
			Rank:       fixture.NonContainerName,
			Pseudo:     false,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			),
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1(fixture.ServerName),
		unknownPseudoNode2(fixture.ServerName),
		theInternetNode(fixture.ServerName),
	)).Prune()

	ServerContainerRenderedID = render.MakeContainerID(fixture.ServerContainerID)
	ClientContainerRenderedID = render.MakeContainerID(fixture.ClientContainerID)

	RenderedContainers = (render.MakeRenderableNodes(
		render.RenderableNode{
			ID:         ClientContainerRenderedID,
			LabelMajor: "client",
			LabelMinor: fixture.ClientHostName,
			Rank:       fixture.ClientContainerImageName,
			Pseudo:     false,
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
		render.RenderableNode{
			ID:         ServerContainerRenderedID,
			LabelMajor: "server",
			LabelMinor: fixture.ServerHostName,
			Rank:       fixture.ServerContainerImageName,
			Pseudo:     false,
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
		render.RenderableNode{
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			),
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		theInternetNode(ServerContainerRenderedID),
	)).Prune()

	ClientContainerImageRenderedName = render.MakeContainerImageID(fixture.ClientContainerImageName)
	ServerContainerImageRenderedName = render.MakeContainerImageID(fixture.ServerContainerImageName)

	RenderedContainerImages = (render.MakeRenderableNodes(
		render.RenderableNode{
			ID:         ClientContainerImageRenderedName,
			LabelMajor: fixture.ClientContainerImageName,
			LabelMinor: "1 container",
			Rank:       fixture.ClientContainerImageName,
			Pseudo:     false,
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
		render.RenderableNode{
			ID:         ServerContainerImageRenderedName,
			LabelMajor: fixture.ServerContainerImageName,
			LabelMinor: "1 container",
			Rank:       fixture.ServerContainerImageName,
			Pseudo:     false,
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
		render.RenderableNode{
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			),
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		theInternetNode(ServerContainerImageRenderedName),
	)).Prune()

	ServerHostRenderedID = render.MakeHostID(fixture.ServerHostID)
	ClientHostRenderedID = render.MakeHostID(fixture.ClientHostID)
	pseudoHostID1        = render.MakePseudoNodeID(fixture.UnknownClient1IP, fixture.ServerIP)
	pseudoHostID2        = render.MakePseudoNodeID(fixture.UnknownClient3IP, fixture.ServerIP)

	RenderedHosts = (render.MakeRenderableNodes(
		render.RenderableNode{
			ID:         ServerHostRenderedID,
			LabelMajor: "server",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
			Children: report.MakeNodeSet(
				fixture.Report.Container.Nodes[fixture.ServerContainerNodeID],
				fixture.Report.Container.Nodes[fixture.ServerProcessNodeID],
			),
			Node: report.MakeNode(),
			EdgeMetadata: report.EdgeMetadata{
				IngressPacketCount: newu64(210),
				IngressByteCount:   newu64(2100),
				MaxConnCountTCP:    newu64(3),
			},
		},
		render.RenderableNode{
			ID:         ClientHostRenderedID,
			LabelMajor: "client",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
			Children: report.MakeNodeSet(
				fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
				fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
				fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID],
			),
			Node: report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
				MaxConnCountTCP:   newu64(3),
			},
		},
		render.RenderableNode{
			ID:           pseudoHostID1,
			LabelMajor:   fixture.UnknownClient1IP,
			Pseudo:       true,
			Node:         report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
			Children: report.MakeNodeSet(
				fixture.Report.Container.Nodes[fixture.ServerContainerNodeID],
				fixture.Report.Process.Nodes[fixture.ServerProcessNodeID],
			),
		},
		render.RenderableNode{
			ID:           pseudoHostID2,
			LabelMajor:   fixture.UnknownClient3IP,
			Pseudo:       true,
			Node:         report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.RenderableNode{
			ID:           render.TheInternetID,
			LabelMajor:   render.TheInternetMajor,
			Pseudo:       true,
			Node:         report.MakeNode().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
		},
	)).Prune()

	ClientPodRenderedID = render.MakePodID("ping/pong-a")
	ServerPodRenderedID = render.MakePodID("ping/pong-b")

	RenderedPods = (render.MakeRenderableNodes(
		render.RenderableNode{
			ID:         ClientPodRenderedID,
			LabelMajor: "pong-a",
			LabelMinor: "1 container",
			Rank:       "ping/pong-a",
			Pseudo:     false,
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
		render.RenderableNode{
			ID:         ServerPodRenderedID,
			LabelMajor: "pong-b",
			LabelMinor: "1 container",
			Rank:       "ping/pong-b",
			Pseudo:     false,
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
		render.RenderableNode{
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			),
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.RenderableNode{
			ID:         render.TheInternetID,
			LabelMajor: render.TheInternetMajor,
			Pseudo:     true,
			Node:       report.MakeNode().WithAdjacent(ServerPodRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
		},
	)).Prune()

	ServiceRenderedID = render.MakeServiceID("ping/pongservice")

	RenderedPodServices = (render.MakeRenderableNodes(
		render.RenderableNode{
			ID:         ServiceRenderedID,
			LabelMajor: "pongservice",
			LabelMinor: "2 pods",
			Rank:       fixture.ServiceID,
			Pseudo:     false,
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
		render.RenderableNode{
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: fixture.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Children: report.MakeNodeSet(
				fixture.Report.Process.Nodes[fixture.NonContainerProcessNodeID],
			),
			Node:         report.MakeNode().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.RenderableNode{
			ID:         render.TheInternetID,
			LabelMajor: render.TheInternetMajor,
			Pseudo:     true,
			Node:       report.MakeNode().WithAdjacent(ServiceRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
		},
	)).Prune()
)

func newu64(value uint64) *uint64 { return &value }
