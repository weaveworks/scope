package expected

import (
	"fmt"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

// Sterilize cleans up RenderableNodes test fixtures
func Sterilize(r render.RenderableNodes) render.RenderableNodes {
	// RenderableNodes contain NodeMetadatas, but generally we
	// only care about the Adjacency field - the rest is internal
	// state not sent to the client.  So in the tests we ignore
	// this state.
	for id, n := range r {
		n.NodeMetadata.Metadata = report.Metadata{}
		n.NodeMetadata.Counters = report.Counters{}
		n.NodeMetadata.Edges = report.EdgeMetadatas{}
		r[id] = n
	}
	return r
}

// Exported for testing.
var (
	uncontainedServerID  = render.MakePseudoNodeID(render.UncontainedID, test.ServerHostName)
	unknownPseudoNode1ID = render.MakePseudoNodeID("10.10.10.10", test.ServerIP, "80")
	unknownPseudoNode2ID = render.MakePseudoNodeID("10.10.10.11", test.ServerIP, "80")
	unknownPseudoNode1   = func(adjacency report.IDList) render.RenderableNode {
		return render.RenderableNode{
			ID:           unknownPseudoNode1ID,
			LabelMajor:   "10.10.10.10",
			Pseudo:       true,
			NodeMetadata: report.MakeNodeMetadata().WithAdjacency(adjacency),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(70),
				EgressByteCount:   newu64(700),
			},
			Origins: report.MakeIDList(
				test.UnknownClient1NodeID,
				test.UnknownClient2NodeID,
			),
		}
	}
	unknownPseudoNode2 = func(adjacency report.IDList) render.RenderableNode {
		return render.RenderableNode{
			ID:           unknownPseudoNode2ID,
			LabelMajor:   "10.10.10.11",
			Pseudo:       true,
			NodeMetadata: report.MakeNodeMetadata().WithAdjacency(adjacency),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(50),
				EgressByteCount:   newu64(500),
			},
			Origins: report.MakeIDList(
				test.UnknownClient3NodeID,
			),
		}
	}
	theInternetNode = func(adjacency report.IDList) render.RenderableNode {
		return render.RenderableNode{
			ID:           render.TheInternetID,
			LabelMajor:   render.TheInternetMajor,
			Pseudo:       true,
			NodeMetadata: report.MakeNodeMetadata().WithAdjacency(adjacency),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(60),
				EgressByteCount:   newu64(600),
			},
			Origins: report.MakeIDList(
				test.RandomClientNodeID,
				test.GoogleEndpointNodeID,
			),
		}
	}
	ClientProcess1ID      = render.MakeProcessID(test.ClientHostID, test.Client1PID)
	ClientProcess2ID      = render.MakeProcessID(test.ClientHostID, test.Client2PID)
	ServerProcessID       = render.MakeProcessID(test.ServerHostID, test.ServerPID)
	nonContainerProcessID = render.MakeProcessID(test.ServerHostID, test.NonContainerPID)

	RenderedProcesses = Sterilize(render.RenderableNodes{
		ClientProcess1ID: {
			ID:         ClientProcess1ID,
			LabelMajor: test.Client1Comm,
			LabelMinor: fmt.Sprintf("%s (%s)", test.ClientHostID, test.Client1PID),
			Rank:       test.Client1Comm,
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.Client54001NodeID,
				test.ClientProcess1NodeID,
				test.ClientHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(ServerProcessID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(10),
				EgressByteCount:   newu64(100),
			},
		},
		ClientProcess2ID: {
			ID:         ClientProcess2ID,
			LabelMajor: test.Client2Comm,
			LabelMinor: fmt.Sprintf("%s (%s)", test.ClientHostID, test.Client2PID),
			Rank:       test.Client2Comm,
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.Client54002NodeID,
				test.ClientProcess2NodeID,
				test.ClientHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(ServerProcessID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(20),
				EgressByteCount:   newu64(200),
			},
		},
		ServerProcessID: {
			ID:         ServerProcessID,
			LabelMajor: "apache",
			LabelMinor: fmt.Sprintf("%s (%s)", test.ServerHostID, test.ServerPID),
			Rank:       test.ServerComm,
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.Server80NodeID,
				test.ServerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(210),
				EgressByteCount:   newu64(2100),
			},
		},
		nonContainerProcessID: {
			ID:         nonContainerProcessID,
			LabelMajor: test.NonContainerComm,
			LabelMinor: fmt.Sprintf("%s (%s)", test.ServerHostID, test.NonContainerPID),
			Rank:       test.NonContainerComm,
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.NonContainerProcessNodeID,
				test.ServerHostNodeID,
				test.NonContainerNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1ID: unknownPseudoNode1(report.MakeIDList(ServerProcessID)),
		unknownPseudoNode2ID: unknownPseudoNode2(report.MakeIDList(ServerProcessID)),
		render.TheInternetID: theInternetNode(report.MakeIDList(ServerProcessID)),
	})

	RenderedProcessNames = Sterilize(render.RenderableNodes{
		"curl": {
			ID:         "curl",
			LabelMajor: "curl",
			LabelMinor: "2 processes",
			Rank:       "curl",
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.Client54001NodeID,
				test.Client54002NodeID,
				test.ClientProcess1NodeID,
				test.ClientProcess2NodeID,
				test.ClientHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent("apache"),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		"apache": {
			ID:         "apache",
			LabelMajor: "apache",
			LabelMinor: "1 process",
			Rank:       "apache",
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.Server80NodeID,
				test.ServerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(210),
				EgressByteCount:   newu64(2100),
			},
		},
		test.NonContainerComm: {
			ID:         test.NonContainerComm,
			LabelMajor: test.NonContainerComm,
			LabelMinor: "1 process",
			Rank:       test.NonContainerComm,
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.NonContainerProcessNodeID,
				test.ServerHostNodeID,
				test.NonContainerNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		unknownPseudoNode1ID: unknownPseudoNode1(report.MakeIDList("apache")),
		unknownPseudoNode2ID: unknownPseudoNode2(report.MakeIDList("apache")),
		render.TheInternetID: theInternetNode(report.MakeIDList("apache")),
	})

	RenderedContainers = Sterilize(render.RenderableNodes{
		test.ClientContainerID: {
			ID:         test.ClientContainerID,
			LabelMajor: "client",
			LabelMinor: test.ClientHostName,
			Rank:       test.ClientContainerImageID,
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.ClientContainerNodeID,
				test.Client54001NodeID,
				test.Client54002NodeID,
				test.ClientProcess1NodeID,
				test.ClientProcess2NodeID,
				test.ClientHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(test.ServerContainerID),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		test.ServerContainerID: {
			ID:         test.ServerContainerID,
			LabelMajor: "server",
			LabelMinor: test.ServerHostName,
			Rank:       test.ServerContainerImageID,
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.ServerContainerNodeID,
				test.Server80NodeID,
				test.ServerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(210),
				EgressByteCount:   newu64(2100),
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
				test.NonContainerNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: theInternetNode(report.MakeIDList(test.ServerContainerID)),
	})

	RenderedContainerImages = Sterilize(render.RenderableNodes{
		test.ClientContainerImageName: {
			ID:         test.ClientContainerImageName,
			LabelMajor: test.ClientContainerImageName,
			LabelMinor: "1 container",
			Rank:       test.ClientContainerImageName,
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.ClientContainerImageNodeID,
				test.ClientContainerNodeID,
				test.Client54001NodeID,
				test.Client54002NodeID,
				test.ClientProcess1NodeID,
				test.ClientProcess2NodeID,
				test.ClientHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(test.ServerContainerImageName),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(30),
				EgressByteCount:   newu64(300),
			},
		},
		test.ServerContainerImageName: {
			ID:         test.ServerContainerImageName,
			LabelMajor: test.ServerContainerImageName,
			LabelMinor: "1 container",
			Rank:       test.ServerContainerImageName,
			Pseudo:     false,
			Origins: report.MakeIDList(
				test.ServerContainerImageNodeID,
				test.ServerContainerNodeID,
				test.Server80NodeID,
				test.ServerProcessNodeID,
				test.ServerHostNodeID),
			NodeMetadata: report.MakeNodeMetadata(),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount: newu64(210),
				EgressByteCount:   newu64(2100),
			},
		},
		uncontainedServerID: {
			ID:         uncontainedServerID,
			LabelMajor: render.UncontainedMajor,
			LabelMinor: test.ServerHostName,
			Rank:       "",
			Pseudo:     true,
			Origins: report.MakeIDList(
				test.NonContainerNodeID,
				test.NonContainerProcessNodeID,
				test.ServerHostNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(render.TheInternetID),
			EdgeMetadata: report.EdgeMetadata{},
		},
		render.TheInternetID: theInternetNode(report.MakeIDList(test.ServerContainerImageName)),
	})

	ServerHostRenderedID = render.MakeHostID(test.ServerHostID)
	ClientHostRenderedID = render.MakeHostID(test.ClientHostID)
	pseudoHostID1        = render.MakePseudoNodeID(test.UnknownClient1IP, test.ServerIP)
	pseudoHostID2        = render.MakePseudoNodeID(test.UnknownClient3IP, test.ServerIP)

	RenderedHosts = Sterilize(render.RenderableNodes{
		ServerHostRenderedID: {
			ID:         ServerHostRenderedID,
			LabelMajor: "server",       // before first .
			LabelMinor: "hostname.com", // after first .
			Rank:       "hostname.com",
			Pseudo:     false,
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
			Origins: report.MakeIDList(
				test.ClientHostNodeID,
				test.ClientAddressNodeID,
			),
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{
				MaxConnCountTCP: newu64(3),
			},
		},
		pseudoHostID1: {
			ID:           pseudoHostID1,
			LabelMajor:   test.UnknownClient1IP,
			Pseudo:       true,
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
			Origins:      report.MakeIDList(test.UnknownAddress1NodeID, test.UnknownAddress2NodeID),
		},
		pseudoHostID2: {
			ID:           pseudoHostID2,
			LabelMajor:   test.UnknownClient3IP,
			Pseudo:       true,
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
			Origins:      report.MakeIDList(test.UnknownAddress3NodeID),
		},
		render.TheInternetID: {
			ID:           render.TheInternetID,
			LabelMajor:   render.TheInternetMajor,
			Pseudo:       true,
			NodeMetadata: report.MakeNodeMetadata().WithAdjacent(ServerHostRenderedID),
			EdgeMetadata: report.EdgeMetadata{},
			Origins:      report.MakeIDList(test.RandomAddressNodeID),
		},
	})
)

func newu64(value uint64) *uint64 { return &value }
