package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func newu64(value uint64) *uint64 { return &value }

// Make sure we don't add a topology and miss it in the Topologies method.
func TestReportTopologies(t *testing.T) {
	var (
		reportType   = reflect.TypeOf(report.MakeReport())
		topologyType = reflect.TypeOf(report.MakeTopology(report.Container))
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
		node := report.MakeNode().WithLatests(map[string]string{
			"foo": "bar",
		})
		if v, _ := node.Latest.Lookup("foo"); v != "bar" {
			t.Errorf("want foo, have %s", v)
		}
	}
	{
		node := report.MakeNode().WithCounters(
			map[string]int{"foo": 1},
		)
		if value, _ := node.Counters.Lookup("foo"); value != 1 {
			t.Errorf("want foo, have %d", value)
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
		if v, ok := node.Edges.Lookup("foo"); ok && *v.EgressPacketCount != 13 {
			t.Errorf("want 13, have %v", node.Edges)
		}
	}
}
