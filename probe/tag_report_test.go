package main

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestApply(t *testing.T) {
	var (
		endpointNodeID = "c"
		addressNodeID  = "d"
		endpointNode   = report.MakeNodeWith(map[string]string{"5": "6"})
		addressNode    = report.MakeNodeWith(map[string]string{"7": "8"})
	)

	r := report.MakeReport()
	r.Endpoint.AddNode(endpointNodeID, endpointNode)
	r.Address.AddNode(addressNodeID, addressNode)
	r = Apply(r, []Tagger{newTopologyTagger()})

	for _, tuple := range []struct {
		want report.Node
		from report.Topology
		via  string
	}{
		{endpointNode.Merge(report.MakeNodeWith(map[string]string{"topology": "endpoint"})), r.Endpoint, endpointNodeID},
		{addressNode.Merge(report.MakeNodeWith(map[string]string{"topology": "address"})), r.Address, addressNodeID},
	} {
		if want, have := tuple.want, tuple.from.Nodes[tuple.via]; !reflect.DeepEqual(want, have) {
			t.Errorf("want %+v, have %+v", want, have)
		}
	}
}

func TestTagMissingID(t *testing.T) {
	const nodeID = "not-found"
	r := report.MakeReport()
	want := report.MakeNode()
	rpt, _ := newTopologyTagger().Tag(r)
	have := rpt.Endpoint.Nodes[nodeID].Copy()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
		t.Error("TopologyTagger erroneously tagged a missing node ID")
	}
}
