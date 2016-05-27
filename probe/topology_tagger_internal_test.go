package probe

import (
	"testing"

	"$GITHUB_URI/report"
)

func TestTagMissingID(t *testing.T) {
	const nodeID = "not-found"
	r := report.MakeReport()
	rpt, _ := NewTopologyTagger().Tag(r)
	_, ok := rpt.Endpoint.Nodes[nodeID]
	if ok {
		t.Error("TopologyTagger erroneously tagged a missing node ID")
	}
}
