package probe

import (
	"github.com/weaveworks/scope/report"
)

type topologyTagger struct{}

// NewTopologyTagger tags each node with the topology that it comes from. It's
// kind of a proof-of-concept tagger, useful primarily for debugging.
func NewTopologyTagger() Tagger {
	return &topologyTagger{}
}

func (topologyTagger) Name() string { return "Topology" }

// Tag implements Tagger
func (topologyTagger) Tag(r report.Report) (report.Report, error) {
	r.WalkNamedTopologies(func(name string, t *report.Topology) {
		for _, node := range t.Nodes {
			t.ReplaceNode(node.WithTopology(name))
		}
	})
	return r, nil
}
