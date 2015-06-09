package tag

import (
	"github.com/weaveworks/scope/report"
)

type topologyTagger struct{}

// NewTopologyTagger tags each node with the topology that it comes from.
func NewTopologyTagger() Tagger {
	return &topologyTagger{}
}

func (topologyTagger) Tag(r report.Report) report.Report {
	for val, topology := range map[string]*report.Topology{
		"endpoint": &(r.Endpoint),
		"network":  &(r.Network),
	} {
		md := report.NodeMetadata{"topology": val}
		for nodeID := range topology.NodeMetadatas {
			(*topology).NodeMetadatas[nodeID].Merge(md)
		}
	}
	return r
}
