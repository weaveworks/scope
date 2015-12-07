package probe

import (
	"compress/gzip"
	"encoding/gob"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestApply(t *testing.T) {
	var (
		endpointNodeID = "c"
		addressNodeID  = "d"
		endpointNode   = report.MakeNodeWith(map[string]string{"5": "6"})
		addressNode    = report.MakeNodeWith(map[string]string{"7": "8"})
	)

	p := New(0, 0, nil)
	p.AddTagger(NewTopologyTagger())

	r := report.MakeReport()
	r.Endpoint.AddNode(endpointNodeID, endpointNode)
	r.Address.AddNode(addressNodeID, addressNode)
	r = p.tag(r)

	for _, tuple := range []struct {
		want report.Node
		from report.Topology
		via  string
	}{
		{endpointNode.Merge(report.MakeNode().WithID("c").WithTopology("endpoint")), r.Endpoint, endpointNodeID},
		{addressNode.Merge(report.MakeNode().WithID("d").WithTopology("address")), r.Address, addressNodeID},
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
	} else if err := gob.NewDecoder(reader).Decode(&r); err != nil {
		return err
	}
	m.have <- r
	return nil
}

func (m mockPublisher) Stop() {
	close(m.have)
}

func TestProbe(t *testing.T) {
	want := report.MakeReport()
	node := report.MakeNodeWith(map[string]string{"b": "c"})
	want.Endpoint.AddNode("a", node)
	pub := mockPublisher{make(chan report.Report)}

	p := New(10*time.Millisecond, 100*time.Millisecond, pub)
	p.AddReporter(mockReporter{want})
	p.Start()
	defer p.Stop()

	test.Poll(t, 300*time.Millisecond, want, func() interface{} {
		return <-pub.have
	})
}
