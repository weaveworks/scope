package render

import (
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/report"
)

// WeaveRenderer is a Renderer which produces a renderable weave topology.
//
// not memoised
var WeaveRenderer = MakeMap(
	MapWeaveIdentity,
	SelectOverlay,
)

// MapWeaveIdentity maps an overlay topology node to a weave topology node.
func MapWeaveIdentity(m report.Node) report.Node {
	peerPrefix, _ := report.ParseOverlayNodeID(m.ID)
	if peerPrefix != report.WeaveOverlayPeerPrefix {
		return report.Node{}
	}

	var (
		node        = m
		nickname, _ = m.Latest.Lookup(overlay.WeavePeerNickName)
	)

	// Nodes without a host id indicate they are not monitored by Scope
	// (their info doesn't come from a probe monitoring that peer directly)
	// , display them as pseudo nodes.
	if _, ok := node.Latest.Lookup(report.HostNodeID); !ok {
		id := MakePseudoNodeID(UnmanagedID, nickname)
		node = NewDerivedPseudoNode(id, m)
	}

	return node
}
