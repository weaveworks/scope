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
	var node = m
	if _, ok := m.Latest.Lookup(report.HostNodeID); !ok {
		nickname, _ := m.Latest.Lookup(overlay.WeavePeerNickName)
		id := MakePseudoNodeID(UnmanagedID, nickname)
		node = NewDerivedPseudoNode(id, m)
	}
	return report.Nodes{node.ID: node}
}
