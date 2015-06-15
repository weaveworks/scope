package tag_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/tag"
	"github.com/weaveworks/scope/report"
)

func TestOriginHostTagger(t *testing.T) {
	var (
		hostID         = "foo"
		endpointNodeID = report.MakeEndpointNodeID(hostID, "1.2.3.4", "56789") // hostID ignored
		nodeMetadata   = report.NodeMetadata{"foo": "bar"}
	)

	r := report.MakeReport()
	r.Endpoint.NodeMetadatas[endpointNodeID] = nodeMetadata
	want := nodeMetadata.Merge(report.NodeMetadata{report.HostNodeID: report.MakeHostNodeID(hostID)})
	have := tag.NewOriginHostTagger(hostID).Tag(r).Endpoint.NodeMetadatas[endpointNodeID].Copy()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("\nwant %+v\nhave %+v", want, have)
	}
}
