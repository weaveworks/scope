package expected

import (
	"fmt"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

// Exported for testing.
var (
	uncontainedServerID  = render.MakePseudoNodeID(render.UncontainedID, test.ServerHostName)
	unknownPseudoNode1ID = render.MakePseudoNodeID("10.10.10.10", test.ServerIP, "80")
	unknownPseudoNode2ID = render.MakePseudoNodeID("10.10.10.11", test.ServerIP, "80")
	unknownPseudoNode1   = render.RenderableNode{
		ID:           unknownPseudoNode1ID,
		LabelMajor:   "10.10.10.10",
		Pseudo:       true,
		NodeMetadata: report.MakeNodeMetadata(),
		EdgeMetadata: report.EdgeMetadata{},
	}
	unknownPseudoNode2 = render.RenderableNode{
		ID:           unknownPseudoNode2ID,
		LabelMajor:   "10.10.10.11",
		Pseudo:       true,
		NodeMetadata: report.MakeNodeMetadata(),
		EdgeMetadata: report.EdgeMetadata{},
	}
	theInternetNode = render.RenderableNode{
		ID:           render.TheInternetID,
		LabelMajor:   render.TheInternetMajor,
		Pseudo:       true,
		NodeMetadata: report.MakeNodeMetadata(),
		EdgeMetadata: report.EdgeMetadata{},
	}

	ClientProcess1ID      = render.MakeProcessID(test.ClientHostID, test.Client1PID)
	ClientProcess2ID      = render.MakeProcessID(test.ClientHostID, test.Client2PID)
	ServerProcessID       = render.MakeProcessID(test.ServerHostID, test.ServerPID)
	nonContainerProcessID = render.MakeProcessID(test.ServerHostID, test.NonContainerPID)

	RenderedProcesses = render.RenderableNodes{
		ClientProcess1ID: {
			ID:         ClientProcess1ID,
			LabelMajor: test.Client1Comm,
			LabelMinor: fmt.Sprintf("%s (%s)", test.ClientHostID, test.Client1PID),
			Rank:       test.Client1Comm,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(ServerProcessID),
			Origins: report.MakeIDList(
				test.Client54001NodeID,
				test.ClientProcess1NodeID,
				test.ClientHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				PacketCount: newu64(100),
				ByteCount:   newu64(10),
			},
		},
		ClientProcess2ID: {
			ID:         ClientProcess2ID,
			LabelMajor: test.Client2Comm,
			LabelMinor: fmt.Sprintf("%s (%s)", test.ClientHostID, test.Client2PID),
			Rank:       test.Client2Comm,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(ServerProcessID),
			Origins: report.MakeIDList(
				test.Client54002NodeID,
				test.ClientProcess2NodeID,
				test.ClientHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				PacketCount: newu64(200),
				ByteCount:   newu64(20),
			},
		},
		ServerProcessID: {
			ID:         ServerProcessID,
			LabelMajor: "apache",
			LabelMinor: fmt.Sprintf("%s (%s)", test.ServerHostID, test.ServerPID),
			Rank:       test.ServerComm,
			Pseudo:     false,
			Adjacency: report.MakeIDList(
				ClientProcess1ID,
				ClientProcess2ID,
				unknownPseudoNode1ID,
				unknownPseudoNode2ID,
				render.TheInternetID,
			),
			Origins: report.MakeIDList(
				test.Server80NodeID,
				test.ServerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				PacketCount: newu64(150),
				ByteCount:   newu64(1500),
			},
		},
		nonContainerProcessID: {
			ID:         nonContainerProcessID,
			LabelMajor: "bash",
			LabelMinor: fmt.Sprintf("%s (%s)", test.ServerHostID, test.NonContainerPID),
			Rank:       test.NonContainerComm,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(),
			Origins: report.MakeIDList(
				test.NonContainerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1ID: unknownPseudoNode1,
		unknownPseudoNode2ID: unknownPseudoNode2,
		render.TheInternetID: theInternetNode,
	}

	RenderedProcessNames = render.RenderableNodes{
		"curl": {
			ID:         "curl",
			LabelMajor: "curl",
			LabelMinor: "",
			Rank:       "curl",
			Pseudo:     false,
			Adjacency:  report.MakeIDList("apache"),
			Origins: report.MakeIDList(
				test.Client54001NodeID,
				test.Client54002NodeID,
				test.ClientProcess1NodeID,
				test.ClientProcess2NodeID,
				test.ClientHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				PacketCount: newu64(300),
				ByteCount:   newu64(30),
			},
		},
		"apache": {
			ID:         "apache",
			LabelMajor: "apache",
			LabelMinor: "",
			Rank:       "apache",
			Pseudo:     false,
			Adjacency: report.MakeIDList(
				"curl",
				unknownPseudoNode1ID,
				unknownPseudoNode2ID,
				render.TheInternetID,
			),
			Origins: report.MakeIDList(
				test.Server80NodeID,
				test.ServerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				PacketCount: newu64(150),
				ByteCount:   newu64(1500),
			},
		},
		"bash": {
			ID:         "bash",
			LabelMajor: "bash",
			LabelMinor: "",
			Rank:       "bash",
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.NonContainerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1ID: unknownPseudoNode1,
		unknownPseudoNode2ID: unknownPseudoNode2,
		render.TheInternetID: theInternetNode,
	}

	RenderedContainers = render.RenderableNodes{
		test.ClientContainerID: {
			ID:         test.ClientContainerID,
			LabelMajor: "client",
			LabelMinor: test.ClientHostName,
			Rank:       test.ClientContainerImageID,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(test.ServerContainerID),
			Origins: report.MakeIDList(
				test.ClientContainerNodeID,
				test.Client54001NodeID,
				test.Client54002NodeID,
				test.ClientProcess1NodeID,
				test.ClientProcess2NodeID,
				test.ClientHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				PacketCount: newu64(300),
				ByteCount:   newu64(30),
			},
		},
		test.ServerContainerID: {
			ID:         test.ServerContainerID,
			LabelMajor: "server",
			LabelMinor: test.ServerHostName,
			Rank:       test.ServerContainerImageID,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(test.ClientContainerID, render.TheInternetID),
			Origins: report.MakeIDList(
				test.ServerContainerNodeID,
				test.Server80NodeID,
				test.ServerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				PacketCount: newu64(150),
				ByteCount:   newu64(1500),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: test.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Origins: report.MakeIDList(
				test.NonContainerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: theInternetNode,
	}

	RenderedContainerImages = render.RenderableNodes{
		test.ClientContainerImageName: {
			ID:         test.ClientContainerImageName,
			LabelMajor: test.ClientContainerImageName,
			LabelMinor: "",
			Rank:       test.ClientContainerImageName,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(test.ServerContainerImageName),
			Origins: report.MakeIDList(
				test.ClientContainerImageNodeID,
				test.ClientContainerNodeID,
				test.Client54001NodeID,
				test.Client54002NodeID,
				test.ClientProcess1NodeID,
				test.ClientProcess2NodeID,
				test.ClientHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				PacketCount: newu64(300),
				ByteCount:   newu64(30),
			},
		},
		test.ServerContainerImageName: {
			ID:         test.ServerContainerImageName,
			LabelMajor: test.ServerContainerImageName,
			LabelMinor: "",
			Rank:       test.ServerContainerImageName,
			Pseudo:     false,
			Adjacency:  report.MakeIDList(test.ClientContainerImageName, render.TheInternetID),
			Origins: report.MakeIDList(
				test.ServerContainerImageNodeID,
				test.ServerContainerNodeID,
				test.Server80NodeID,
				test.ServerProcessNodeID,
				test.ServerHostNodeID),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				PacketCount: newu64(150),
				ByteCount:   newu64(1500),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: test.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Origins: report.MakeIDList(
				test.NonContainerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: theInternetNode,
	}

	ServerHostRenderedID = render.MakeHostID(test.ServerHostID)
	ClientHostRenderedID = render.MakeHostID(test.ClientHostID)
	pseudoHostID1        = render.MakePseudoNodeID("10.10.10.10", "192.168.1.1", "")
	pseudoHostID2        = render.MakePseudoNodeID("10.10.10.11", "192.168.1.1", "")

	RenderedHosts = render.RenderableNodes{
		ServerHostRenderedID: {
			ID:         ServerHostRenderedID,
			LabelMajor: "server",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
			Adjacency:  report.MakeIDList(ClientHostRenderedID, render.TheInternetID, pseudoHostID1, pseudoHostID2),
			Origins: report.MakeIDList(
				test.ServerHostNodeID,
				test.ServerAddressNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				MaxConnCountTCP: newu64(3),
			},
		},
		ClientHostRenderedID: {
			ID:         ClientHostRenderedID,
			LabelMajor: "client",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
			Adjacency:  report.MakeIDList(ServerHostRenderedID),
			Origins: report.MakeIDList(
				test.ClientHostNodeID,
				test.ClientAddressNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				MaxConnCountTCP: newu64(3),
			},
		},
		pseudoHostID1: {
			ID:           pseudoHostID1,
			LabelMajor:   "10.10.10.10",
			Pseudo:       true,
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{},
		},
		pseudoHostID2: {
			ID:           pseudoHostID2,
			LabelMajor:   "10.10.10.11",
			Pseudo:       true,
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: theInternetNode,
	}
)

func newu64(value uint64) *uint64 { return &value }
