package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

// Make sure we don't add a topology and miss it in the Topologies method.
func TestReportTopologies(t *testing.T) {
	var (
		reportType   = reflect.TypeOf(report.MakeReport())
		topologyType = reflect.TypeOf(report.MakeTopology())
	)

	var want int
	for i := 0; i < reportType.NumField(); i++ {
		if reportType.Field(i).Type == topologyType {
			want++
		}
	}

	if have := len(report.MakeReport().Topologies()); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestNode(t *testing.T) {
	{
		node := report.MakeNode().WithMetadata(report.Metadata{
			"foo": "bar",
		})
		if node.Metadata["foo"] != "bar" {
			t.Errorf("want foo, have %s", node.Metadata["foo"])
		}
	}
	{
		node := report.MakeNode().WithCounters(report.Counters{
			"foo": 1,
		})
		if node.Counters["foo"] != 1 {
			t.Errorf("want foo, have %d", node.Counters["foo"])
		}
	}
	{
		node := report.MakeNode().WithAdjacent("foo")
		if node.Adjacency[0] != "foo" {
			t.Errorf("want foo, have %v", node.Adjacency)
		}
	}
	{
		node := report.MakeNode().WithEdge("foo", report.EdgeMetadata{
			EgressPacketCount: newu64(13),
		})
		if node.Adjacency[0] != "foo" {
			t.Errorf("want foo, have %v", node.Adjacency)
		}
		if *node.Edges["foo"].EgressPacketCount != 13 {
			t.Errorf("want 13, have %v", node.Edges)
		}
	}
}
