package endpoint

import (
	"encoding/xml"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

type mockConntracker struct {
	flows []Flow
}

func (m *mockConntracker) WalkFlows(f func(Flow)) {
	for _, flow := range m.flows {
		f(flow)
	}
}

func (m *mockConntracker) Stop() {}

// TODO(pb): dedupe later
func makeFlow(ty string) Flow {
	return Flow{
		XMLName: xml.Name{
			Local: "flow",
		},
		Type: ty,
	}
}

// TODO(pb): dedupe later
func addMeta(f *Flow, dir, srcIP, dstIP string, srcPort, dstPort int) *Meta {
	meta := Meta{
		XMLName: xml.Name{
			Local: "meta",
		},
		Direction: dir,
		Layer3: Layer3{
			XMLName: xml.Name{
				Local: "layer3",
			},
			SrcIP: srcIP,
			DstIP: dstIP,
		},
		Layer4: Layer4{
			XMLName: xml.Name{
				Local: "layer4",
			},
			SrcPort: srcPort,
			DstPort: dstPort,
			Proto:   TCP,
		},
	}
	f.Metas = append(f.Metas, meta)
	return &meta
}

// TODO(pb): dedupe later
func addIndependant(f *Flow, id int64, state string) *Meta {
	meta := Meta{
		XMLName: xml.Name{
			Local: "meta",
		},
		Direction: "independent",
		ID:        id,
		State:     state,
		Layer3: Layer3{
			XMLName: xml.Name{
				Local: "layer3",
			},
		},
		Layer4: Layer4{
			XMLName: xml.Name{
				Local: "layer4",
			},
		},
	}
	f.Metas = append(f.Metas, meta)
	return &meta
}

func TestNat(t *testing.T) {
	// test that two containers, on the docker network, get their connections mapped
	// correctly.
	// the setup is this:
	//
	// container2 (10.0.47.2:222222), host2 (2.3.4.5:22223) ->
	//     host1 (1.2.3.4:80), container1 (10.0.47.2:80)

	// from the PoV of host1
	{
		flow := makeFlow("")
		addIndependant(&flow, 1, "")
		flow.Original = addMeta(&flow, "original", "2.3.4.5", "1.2.3.4", 222222, 80)
		flow.Reply = addMeta(&flow, "reply", "10.0.47.1", "2.3.4.5", 80, 222222)
		ct := &mockConntracker{
			flows: []Flow{flow},
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
		flow := makeFlow("")
		addIndependant(&flow, 2, "")
		flow.Original = addMeta(&flow, "original", "10.0.47.2", "1.2.3.4", 22222, 80)
		flow.Reply = addMeta(&flow, "reply", "1.2.3.4", "2.3.4.5", 80, 22223)
		ct := &mockConntracker{
			flows: []Flow{flow},
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
