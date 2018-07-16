package endpoint

import (
	"testing"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

type mockFlowWalker struct {
	flows []flow
}

func (m *mockFlowWalker) walkFlows(f func(f flow, active bool)) {
	for _, flow := range m.flows {
		f(flow, true)
	}
}

func (m *mockFlowWalker) stop() {}

func TestNat(t *testing.T) {
	mtime.NowForce(mtime.Now())
	defer mtime.NowReset()

	// test that two containers, on the docker network, get their connections mapped
	// correctly.
	// the setup is this:
	//
	// container2 (10.0.47.2:22222), host2 (2.3.4.5:22223) ->
	//     host1 (1.2.3.4:80), container1 (10.0.47.1:80)

	// from the PoV of host1
	{
		f := flow{
			Type: updateType,
			Original: meta{
				Layer3: layer3{
					SrcIP: "2.3.4.5",
					DstIP: "1.2.3.4",
				},
				Layer4: layer4{
					SrcPort: 22222,
					DstPort: 80,
					Proto:   "tcp",
				},
			},
			Reply: meta{
				Layer3: layer3{
					SrcIP: "10.0.47.1",
					DstIP: "2.3.4.5",
				},
				Layer4: layer4{
					SrcPort: 80,
					DstPort: 22222,
					Proto:   "tcp",
				},
			},
			Independent: meta{
				ID: 1,
			},
		}

		ct := &mockFlowWalker{
			flows: []flow{f},
		}

		have := report.MakeReport()
		originalID := report.MakeEndpointNodeID("host1", "", "10.0.47.1", "80")
		have.Endpoint.AddNode(report.MakeNodeWith(originalID,
			"foo", "bar",
		))

		want := have.Copy()
		wantID := report.MakeEndpointNodeID("host1", "", "1.2.3.4", "80")
		want.Endpoint.AddNode(report.MakeNodeWith(wantID,
			CopyOf, originalID,
			"foo", "bar",
		))

		makeNATMapper(ct).applyNAT(have, "host1")
		if !reflect.DeepEqual(want, have) {
			t.Fatal(test.Diff(want, have))
		}
	}

	// form the PoV of host2
	{
		f := flow{
			Type: updateType,
			Original: meta{
				Layer3: layer3{
					SrcIP: "10.0.47.2",
					DstIP: "1.2.3.4",
				},
				Layer4: layer4{
					SrcPort: 22222,
					DstPort: 80,
					Proto:   "tcp",
				},
			},
			Reply: meta{
				Layer3: layer3{
					SrcIP: "1.2.3.4",
					DstIP: "2.3.4.5",
				},
				Layer4: layer4{
					SrcPort: 80,
					DstPort: 22223,
					Proto:   "tcp",
				},
			},
			Independent: meta{
				ID: 2,
			},
		}
		ct := &mockFlowWalker{
			flows: []flow{f},
		}

		have := report.MakeReport()
		originalID := report.MakeEndpointNodeID("host2", "", "10.0.47.2", "22222")
		have.Endpoint.AddNode(report.MakeNodeWith(originalID,
			"foo", "baz",
		))

		want := have.Copy()
		want.Endpoint.AddNode(report.MakeNodeWith(report.MakeEndpointNodeID("host2", "", "2.3.4.5", "22223"),
			CopyOf, originalID,
			"foo", "baz",
		))

		makeNATMapper(ct).applyNAT(have, "host1")
		if !reflect.DeepEqual(want, have) {
			t.Fatal(test.Diff(want, have))
		}
	}
}
