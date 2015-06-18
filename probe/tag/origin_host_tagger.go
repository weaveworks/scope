package tag

import (
	"github.com/weaveworks/scope/report"
)

type originHostTagger struct{ hostNodeID string }

// NewOriginHostTagger tags each node with a foreign key linking it to its
// origin host in the host topology.
func NewOriginHostTagger(hostID string) Tagger {
	return &originHostTagger{hostNodeID: report.MakeHostNodeID(hostID)}
}

func (t originHostTagger) Tag(r report.Report) (report.Report, error) {
	for _, topology := range r.Topologies() {
		md := report.NodeMetadata{report.HostNodeID: t.hostNodeID}
		for nodeID := range topology.NodeMetadatas {
			topology.NodeMetadatas[nodeID].Merge(md)
		}
	}
	return r, nil
}
