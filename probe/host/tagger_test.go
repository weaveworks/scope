package host_test

import (
	"testing"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

func TestTagger(t *testing.T) {
	var (
		hostID         = "foo"
		endpointNodeID = report.MakeEndpointNodeID(hostID, "", "1.2.3.4", "56789") // hostID ignored
		node           = report.MakeNodeWith(endpointNodeID, map[string]string{"foo": "bar"})
	)

	r := report.MakeReport()
	r.Process.AddNode(node)
	rpt, _ := host.NewTagger(hostID).Tag(r)
	have := rpt.Process.Nodes[endpointNodeID]

	// It should now have the host ID
	wantHostID := report.MakeHostNodeID(hostID)
	if hostID, ok := have.Latest.Lookup(report.HostNodeID); !ok || hostID != wantHostID {
		t.Errorf("Expected %q got %q", wantHostID, report.MakeHostNodeID(hostID))
	}

	// It should still have the other keys
	want := "bar"
	if have, ok := have.Latest.Lookup("foo"); !ok || have != want {
		t.Errorf("Expected %q got %q", want, have)
	}

	// It should have the host as a parent
	wantParent := report.MakeHostNodeID(hostID)
	if have, ok := have.Parents.Lookup(report.Host); !ok || len(have) != 1 || have[0] != wantParent {
		t.Errorf("Expected %q got %q", report.MakeStringSet(wantParent), have)
	}
}
