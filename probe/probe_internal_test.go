package probe

import (
	"compress/gzip"
	"io"
	"testing"
	"time"

	"github.com/ugorji/go/codec"
	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestApply(t *testing.T) {
	var (
		endpointNodeID = "c"
		endpointNode   = report.MakeNodeWith(endpointNodeID, map[string]string{"5": "6"})
	)

	p := New("", 0, 0, nil)
	p.AddTagger(NewTopologyTagger())

	r := report.MakeReport()
	r.Endpoint.AddNode(endpointNodeID, endpointNode)
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

func (m mockPublisher) Publish(in io.Reader) error {
	var r report.Report
	if reader, err := gzip.NewReader(in); err != nil {
		return err
	} else if err := codec.NewDecoder(reader, &codec.MsgpackHandle{}).Decode(&r); err != nil {
		return err
	}
	m.have <- r
	return nil
}

func (m mockPublisher) Stop() {
	close(m.have)
}

func TestProbe(t *testing.T) {
	// marshalling->unmarshaling is not idempotent due to `json:"omitempty"`
	// tags, transforming empty slices into nils. So, we make DeepEqual
	// happy by setting empty `json:"omitempty"` entries to nil
	const probeID = "probeid"
	now := time.Now()
	mtime.NowForce(now)
	defer mtime.NowReset()

	want := report.MakeReport()
	node := report.MakeNodeWith("a", map[string]string{"b": "c"})
	node.Metrics = nil // omitempty
	want.Endpoint.AddNode("a", node)
	want.Probes[probeID] = report.Probe{
		ID:       probeID,
		LastSeen: now,
	}

	pub := mockPublisher{make(chan report.Report)}

	// omitempty
	want.Endpoint.Controls = nil
	want.Process.Controls = nil
	want.Container.Controls = nil
	want.ContainerImage.Controls = nil
	want.Pod.Controls = nil
	want.Service.Controls = nil
	want.Host.Controls = nil
	want.Overlay.Controls = nil

	p := New(probeID, 10*time.Millisecond, 100*time.Millisecond, pub)
	p.AddReporter(mockReporter{want})
	p.Start()
	defer p.Stop()

	test.Poll(t, 300*time.Millisecond, want, func() interface{} {
		return <-pub.have
	})
}
