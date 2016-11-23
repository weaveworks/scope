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
		report.Endpoint:       &(r.Endpoint),
		report.Process:        &(r.Process),
		report.Container:      &(r.Container),
		report.ContainerImage: &(r.ContainerImage),
		report.Pod:            &(r.Pod),
		report.Service:        &(r.Service),
		report.ECSTask:        &(r.ECSTask),
		report.ECSService:     &(r.ECSService),
		report.Host:           &(r.Host),
		report.Overlay:        &(r.Overlay),
	} {
		for _, node := range t.Nodes {
			t.AddNode(node.WithTopology(name))
		}
	}
	return r, nil
}
