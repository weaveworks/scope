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
	for name, t := range map[string]*report.Topology{
		"endpoint":        &(r.Endpoint),
		"address":         &(r.Address),
		"process":         &(r.Process),
		"container":       &(r.Container),
		"container_image": &(r.ContainerImage),
		"pod":             &(r.Pod),
		"service":         &(r.Service),
		"host":            &(r.Host),
		"overlay":         &(r.Overlay),
	} {
		for id, node := range t.Nodes {
			t.AddNode(id, node.WithID(id).WithTopology(name))
		}
	}
	return r, nil
}
