package render

import (
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/report"
)

// WeaveRenderer is a Renderer which produces a renderable weave topology.
var WeaveRenderer = MakeMap(
	MapWeaveIdentity,
	SelectOverlay,
)

// MapWeaveIdentity maps an overlay topology node to a weave topology node.
func MapWeaveIdentity(m report.Node, _ report.Networks) report.Nodes {
	peerPrefix, _ := report.ParseOverlayNodeID(m.ID)
	if peerPrefix != report.WeaveOverlayPeerPrefix {
		return nil
	}

	var (
		node        = m
		nickname, _ = m.Latest.Lookup(overlay.WeavePeerNickName)
	)

	if _, ok := node.Latest.Lookup(report.HostNodeID); !ok {
		id := MakePseudoNodeID(UnmanagedID, nickname)
		node = NewDerivedPseudoNode(id, m)
	}

	return report.Nodes{node.ID: node}
}
