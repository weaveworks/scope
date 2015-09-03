package host_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestTagger(t *testing.T) {
	var (
		hostID         = "foo"
		endpointNodeID = report.MakeEndpointNodeID(hostID, "1.2.3.4", "56789") // hostID ignored
		nodeMetadata   = report.MakeNodeWith(map[string]string{"foo": "bar"})
	)

	r := report.MakeReport()
	r.Process.Nodes[endpointNodeID] = nodeMetadata
	want := nodeMetadata.Merge(report.MakeNodeWith(map[string]string{
		report.HostNodeID: report.MakeHostNodeID(hostID),
	}))
	rpt, _ := host.NewTagger(hostID).Tag(r)
	have := rpt.Process.Nodes[endpointNodeID].Copy()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
