package tag_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/tag"
	"github.com/weaveworks/scope/report"
)

func TestApply(t *testing.T) {
	var (
		processNodeID       = "c"
		networkNodeID       = "d"
		processNodeMetadata = report.NodeMetadata{"5": "6"}
		networkNodeMetadata = report.NodeMetadata{"7": "8"}
	)

	r := report.MakeReport()
	r.Process.NodeMetadatas[processNodeID] = processNodeMetadata
	r.Network.NodeMetadatas[networkNodeID] = networkNodeMetadata
	r = tag.Apply(r, []tag.Tagger{tag.NewTopologyTagger()})

	for _, tuple := range []struct {
		want report.NodeMetadata
		from report.Topology
		via  string
	}{
		{copy(processNodeMetadata).Merge(report.NodeMetadata{"topology": "process"}), r.Process, processNodeID},
		{copy(networkNodeMetadata).Merge(report.NodeMetadata{"topology": "network"}), r.Network, networkNodeID},
	} {
		if want, have := tuple.want, tuple.from.NodeMetadatas[tuple.via]; !reflect.DeepEqual(want, have) {
			t.Errorf("want %+v, have %+v", want, have)
		}
	}
}

func copy(input report.NodeMetadata) report.NodeMetadata {
	output := make(report.NodeMetadata, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}
