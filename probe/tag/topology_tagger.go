package tag

import (
	"github.com/weaveworks/scope/report"
)

type topologyTagger struct{}

// NewTopologyTagger tags each node with the topology that it comes from.
func NewTopologyTagger() Tagger {
	return &topologyTagger{}
}

func (topologyTagger) Tag(r report.Report, _ report.TopologySelector, id string) report.NodeMetadata {
	for val, ts := range map[string]report.TopologySelector{
		"endpoint": report.SelectEndpoint,
		"address":  report.SelectAddress,
		"process":  report.SelectProcess,
		"host":     report.SelectHost,
	} {
		if _, ok := ts(r).NodeMetadatas[id]; ok {
			return report.NodeMetadata{"topology": val}
		}
	}
	return report.NodeMetadata{}
}

func (topologyTagger) Stop() {}
