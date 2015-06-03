package tag_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/tag"
	"github.com/weaveworks/scope/report"
)

func TestApply(t *testing.T) {
	var (
		endpointNodeID       = "a"
		addressNodeID        = "b"
		processNodeID        = "c"
		hostNodeID           = "d"
		endpointNodeMetadata = report.NodeMetadata{"1": "2"}
		addressNodeMetadata  = report.NodeMetadata{"3": "4"}
		processNodeMetadata  = report.NodeMetadata{"5": "6"}
		hostNodeMetadata     = report.NodeMetadata{"7": "8"}
	)

	r := report.MakeReport()
	r.Endpoint.NodeMetadatas[endpointNodeID] = endpointNodeMetadata
	r.Address.NodeMetadatas[addressNodeID] = addressNodeMetadata
	r.Process.NodeMetadatas[processNodeID] = processNodeMetadata
	r.Host.NodeMetadatas[hostNodeID] = hostNodeMetadata
	r = tag.Apply(r, []tag.Tagger{tag.NewTopologyTagger()})

	for _, tuple := range []struct {
		want report.NodeMetadata
		from report.Topology
		via  string
	}{
		{copy(endpointNodeMetadata).Merge(report.NodeMetadata{"topology": "endpoint"}), r.Endpoint, endpointNodeID},
		{copy(addressNodeMetadata).Merge(report.NodeMetadata{"topology": "address"}), r.Address, addressNodeID},
		{copy(processNodeMetadata).Merge(report.NodeMetadata{"topology": "process"}), r.Process, processNodeID},
		{copy(hostNodeMetadata).Merge(report.NodeMetadata{"topology": "host"}), r.Host, hostNodeID},
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
