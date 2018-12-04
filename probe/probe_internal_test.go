package probe

import (
	"testing"
	"time"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestApply(t *testing.T) {
	var (
		endpointNodeID = "c"
		endpointNode   = report.MakeNodeWith(endpointNodeID, map[string]string{"5": "6"})
	)

	p := New(0, 0, nil, false)
	p.AddTagger(NewTopologyTagger())

	r := report.MakeReport()
	r.Endpoint.AddNode(endpointNode)
	r = p.tag(r)

	for _, tuple := range []struct {
		want report.Node
		from report.Topology
		via  string
	}{
		{endpointNode.Merge(report.MakeNode("c").WithTopology(report.Endpoint)), r.Endpoint, endpointNodeID},
	} {
		if want, have := tuple.want, tuple.from.Nodes[tuple.via]; !reflect.DeepEqual(want, have) {
			t.Errorf("want %+v, have %+v", want, have)
		}
	}
}

type mockReporter struct {
	r report.Report
}

func (m mockReporter) Report() (report.Report, error) {
	return m.r.Copy(), nil
}

func (mockReporter) Name() string { return "Mock" }

type mockPublisher struct {
	have chan report.Report
}

func (m mockPublisher) Publish(r report.Report) error {
	m.have <- r
	return nil
}

func (m mockPublisher) Stop() {
	close(m.have)
}

func TestProbe(t *testing.T) {
	const probeID = "probeid"
	now := time.Now()
	mtime.NowForce(now)
	defer mtime.NowReset()

	want := report.MakeReport()
	node := report.MakeNodeWith("a", map[string]string{"b": "c"})

	// marshalling->unmarshaling is not idempotent due to `json:"omitempty"`
	// tags, transforming empty slices into nils. So, we make DeepEqual
	// happy by setting empty `json:"omitempty"` entries to nil
	node.Metrics = nil
	want.WalkTopologies(func(t *report.Topology) {
		t.Controls = nil
	})
	want.Endpoint.AddNode(node)

	pub := mockPublisher{make(chan report.Report, 10)}

	p := New(10*time.Millisecond, 100*time.Millisecond, pub, false)
	p.AddReporter(mockReporter{want})
	p.Start()
	defer p.Stop()

	test.Poll(t, 300*time.Millisecond, want, func() interface{} {
		return <-pub.have
	})
}
