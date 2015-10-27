package endpoint

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

type mockFlowWalker struct {
	flows []flow
}

func (m *mockFlowWalker) walkFlows(f func(flow)) {
	for _, flow := range m.flows {
		f(flow)
	}
}

func (m *mockFlowWalker) stop() {}

func TestNat(t *testing.T) {
	// test that two containers, on the docker network, get their connections mapped
	// correctly.
	// the setup is this:
	//
	// container2 (10.0.47.2:222222), host2 (2.3.4.5:22223) ->
	//     host1 (1.2.3.4:80), container1 (10.0.47.2:80)

	// from the PoV of host1
	{
		f := makeFlow("")
		addIndependant(&f, 1, "")
		f.Original = addMeta(&f, "original", "2.3.4.5", "1.2.3.4", 222222, 80)
		f.Reply = addMeta(&f, "reply", "10.0.47.1", "2.3.4.5", 80, 222222)
		ct := &mockFlowWalker{
			flows: []flow{f},
		}

		have := report.MakeReport()
		originalID := report.MakeEndpointNodeID("host1", "10.0.47.1", "80")
		have.Endpoint.AddNode(originalID, report.MakeNodeWith(report.Metadata{
			Addr:  "10.0.47.1",
			Port:  "80",
			"foo": "bar",
		}))

		want := have.Copy()
		want.Endpoint.AddNode(report.MakeEndpointNodeID("host1", "1.2.3.4", "80"), report.MakeNodeWith(report.Metadata{
			Addr:      "1.2.3.4",
			Port:      "80",
			"copy_of": originalID,
			"foo":     "bar",
		}))

		makeNATMapper(ct).applyNAT(have, "host1")
		if !reflect.DeepEqual(want, have) {
			t.Fatal(test.Diff(want, have))
		}
	}

	// form the PoV of host2
	{
		f := makeFlow("")
		addIndependant(&f, 2, "")
		f.Original = addMeta(&f, "original", "10.0.47.2", "1.2.3.4", 22222, 80)
		f.Reply = addMeta(&f, "reply", "1.2.3.4", "2.3.4.5", 80, 22223)
		ct := &mockFlowWalker{
			flows: []flow{f},
		}

		have := report.MakeReport()
		originalID := report.MakeEndpointNodeID("host2", "10.0.47.2", "22222")
		have.Endpoint.AddNode(originalID, report.MakeNodeWith(report.Metadata{
			Addr:  "10.0.47.2",
			Port:  "22222",
			"foo": "baz",
		}))

		want := have.Copy()
		want.Endpoint.AddNode(report.MakeEndpointNodeID("host2", "2.3.4.5", "22223"), report.MakeNodeWith(report.Metadata{
			Addr:      "2.3.4.5",
			Port:      "22223",
			"copy_of": originalID,
			"foo":     "baz",
		}))

		makeNATMapper(ct).applyNAT(have, "host1")
		if !reflect.DeepEqual(want, have) {
			t.Fatal(test.Diff(want, have))
		}
	}
}
