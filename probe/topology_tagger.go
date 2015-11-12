package probe

import (
	"github.com/weaveworks/scope/report"
)

// Topology is the Node key for the origin topology.
const Topology = "topology"

type topologyTagger struct{}

// NewTopologyTagger tags each node with the topology that it comes from. It's
// kind of a proof-of-concept tagger, useful primarily for debugging.
func NewTopologyTagger() Tagger {
	return &topologyTagger{}
}

func (topologyTagger) Name() string { return "Topology" }

// Tag implements Tagger
func (topologyTagger) Tag(r report.Report) (report.Report, error) {
	for val, topology := range map[string]*report.Topology{
		"endpoint":        &(r.Endpoint),
		"address":         &(r.Address),
		"process":         &(r.Process),
		"container":       &(r.Container),
		"container_image": &(r.ContainerImage),
		"host":            &(r.Host),
		"overlay":         &(r.Overlay),
	} {
		metadata := map[string]string{Topology: val}
		for id, node := range topology.Nodes {
			topology.AddNode(id, node.WithMetadata(metadata))
		}
	}
	return r, nil
}
