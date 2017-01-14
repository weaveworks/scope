package endpoint

import (
	"net"
	"syscall"
	"testing"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/probe/endpoint/conntrack"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

type mockFlowWalker struct {
	flows []conntrack.Flow
}

func (m *mockFlowWalker) walkFlows(f func(f conntrack.Flow, active bool)) {
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

	c1 := net.ParseIP("10.0.47.1")
	c2 := net.ParseIP("10.0.47.2")
	host2 := net.ParseIP("2.3.4.5")
	host1 := net.ParseIP("1.2.3.4")

	// from the PoV of host1
	{
		f := conntrack.Flow{
			MsgType: conntrack.NfctMsgUpdate,
			Original: conntrack.Meta{
				Layer3: conntrack.Layer3{
					SrcIP: host2,
					DstIP: host1,
				},
				Layer4: conntrack.Layer4{
					SrcPort: 22222,
					DstPort: 80,
					Proto:   syscall.IPPROTO_TCP,
				},
			},
			Reply: conntrack.Meta{
				Layer3: conntrack.Layer3{
					SrcIP: c1,
					DstIP: host2,
				},
				Layer4: conntrack.Layer4{
					SrcPort: 80,
					DstPort: 22222,
					Proto:   syscall.IPPROTO_TCP,
				},
			},
			ID: 1,
		}

		ct := &mockFlowWalker{
			flows: []conntrack.Flow{f},
		}

		have := report.MakeReport()
		originalID := report.MakeEndpointNodeID("host1", "", "10.0.47.1", "80")
		have.Endpoint.AddNode(report.MakeNodeWith(originalID, map[string]string{
			"foo": "bar",
		}))

		want := have.Copy()
		wantID := report.MakeEndpointNodeID("host1", "", "1.2.3.4", "80")
		want.Endpoint.AddNode(report.MakeNodeWith(wantID, map[string]string{
			CopyOf: originalID,
			"foo":  "bar",
		}))

		makeNATMapper(ct).applyNAT(have, "host1")
		if !reflect.DeepEqual(want, have) {
			t.Fatal(test.Diff(want, have))
		}
	}

	// form the PoV of host2
	{
		f := conntrack.Flow{
			MsgType: conntrack.NfctMsgUpdate,
			Original: conntrack.Meta{
				Layer3: conntrack.Layer3{
					SrcIP: c2,
					DstIP: host1,
				},
				Layer4: conntrack.Layer4{
					SrcPort: 22222,
					DstPort: 80,
					Proto:   syscall.IPPROTO_TCP,
				},
			},
			Reply: conntrack.Meta{
				Layer3: conntrack.Layer3{
					SrcIP: host1,
					DstIP: host2,
				},
				Layer4: conntrack.Layer4{
					SrcPort: 80,
					DstPort: 22223,
					Proto:   syscall.IPPROTO_TCP,
				},
			},
			ID: 2,
		}
		ct := &mockFlowWalker{
			flows: []conntrack.Flow{f},
		}

		have := report.MakeReport()
		originalID := report.MakeEndpointNodeID("host2", "", "10.0.47.2", "22222")
		have.Endpoint.AddNode(report.MakeNodeWith(originalID, map[string]string{
			"foo": "baz",
		}))

		want := have.Copy()
		want.Endpoint.AddNode(report.MakeNodeWith(report.MakeEndpointNodeID("host2", "", "2.3.4.5", "22223"), map[string]string{
			CopyOf: originalID,
			"foo":  "baz",
		}))

		makeNATMapper(ct).applyNAT(have, "host1")
		if !reflect.DeepEqual(want, have) {
			t.Fatal(test.Diff(want, have))
		}
	}
}
