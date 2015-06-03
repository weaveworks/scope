package tag_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/tag"
	"github.com/weaveworks/scope/report"
)

func TestTagMissingID(t *testing.T) {
	const nodeID = "not-found"
	r := report.MakeReport()
	want := report.NodeMetadata{}
	have := tag.NewTopologyTagger().Tag(r, report.SelectAddress, nodeID)
	if !reflect.DeepEqual(want, have) {
		t.Error("TopologyTagger erroneously tagged a missing node ID")
	}
}
