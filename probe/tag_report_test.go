package main

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestApply(t *testing.T) {
	var (
		endpointNodeID       = "c"
		addressNodeID        = "d"
		endpointNodeMetadata = report.NodeMetadata{"5": "6"}
		addressNodeMetadata  = report.NodeMetadata{"7": "8"}
	)

	r := report.MakeReport()
	r.Endpoint.NodeMetadatas[endpointNodeID] = endpointNodeMetadata
	r.Address.NodeMetadatas[addressNodeID] = addressNodeMetadata
	r = Apply(r, []Tagger{newTopologyTagger()})

	for _, tuple := range []struct {
		want report.NodeMetadata
		from report.Topology
		via  string
	}{
		{endpointNodeMetadata.Copy().Merge(report.NodeMetadata{"topology": "endpoint"}), r.Endpoint, endpointNodeID},
		{addressNodeMetadata.Copy().Merge(report.NodeMetadata{"topology": "address"}), r.Address, addressNodeID},
	} {
		if want, have := tuple.want, tuple.from.NodeMetadatas[tuple.via]; !reflect.DeepEqual(want, have) {
			t.Errorf("want %+v, have %+v", want, have)
		}
	}
}

func TestTagMissingID(t *testing.T) {
	const nodeID = "not-found"
	r := report.MakeReport()
	want := report.NodeMetadata{}
	rpt, _ := newTopologyTagger().Tag(r)
	have := rpt.Endpoint.NodeMetadatas[nodeID].Copy()
	if !reflect.DeepEqual(want, have) {
		t.Error("TopologyTagger erroneously tagged a missing node ID")
	}
}
