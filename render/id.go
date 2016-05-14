package render

import (
	"strings"

	"$GITHUB_URI/report"
)

// MakePseudoNodeID joins the parts of an id into the id of a pseudonode
func MakePseudoNodeID(parts ...string) string {
	return strings.Join(append([]string{"pseudo"}, parts...), ":")
}

// MakeGroupNodeTopology joins the parts of a group topology into the topology of a group node
func MakeGroupNodeTopology(originalTopology, key string) string {
	return strings.Join([]string{"group", originalTopology, key}, ":")
}

// NewDerivedNode makes a node based on node, but with a new ID
func NewDerivedNode(id string, node report.Node) report.Node {
	return report.MakeNode(id).WithChildren(node.Children.Add(node))
}

// NewDerivedPseudoNode makes a new pseudo node with the node as a child
func NewDerivedPseudoNode(id string, node report.Node) report.Node {
	output := NewDerivedNode(id, node).WithTopology(Pseudo)
	return output
}
