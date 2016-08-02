package report_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	s_reflect "github.com/weaveworks/scope/test/reflect"
)

func newu64(value uint64) *uint64 { return &value }

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

func TestReportTopology(t *testing.T) {
	r := report.MakeReport()
	if _, ok := r.Topology(report.Container); !ok {
		t.Errorf("Expected %s topology to be found", report.Container)
	}
	if _, ok := r.Topology("foo"); ok {
		t.Errorf("Expected %s topology not to be found", "foo")
	}
}

func TestNode(t *testing.T) {
	{
		node := report.MakeNodeWith("foo", map[string]string{
			"foo": "bar",
		})

		if v, _ := node.Latest.Lookup("foo"); v != "bar" {
			t.Errorf("want foo, have %s", v)
		}
	}
	{
		node := report.MakeNode("foo").WithCounters(
			map[string]int{"foo": 1},
		)
		if value, _ := node.Counters.Lookup("foo"); value != 1 {
			t.Errorf("want foo, have %d", value)
		}
	}
	{
		node := report.MakeNode("foo").WithAdjacent("foo")
		if node.Adjacency[0] != "foo" {
			t.Errorf("want foo, have %v", node.Adjacency)
		}
	}
	{
		node := report.MakeNode("foo").WithEdge("foo", report.EdgeMetadata{
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

func TestReportBackwardCompatibility(t *testing.T) {
	mtime.NowForce(time.Now())
	defer mtime.NowReset()
	rpt := report.MakeReport()
	controls := map[string]report.NodeControlData{
		"dead": {
			Dead: true,
		},
		"alive": {
			Dead: false,
		},
	}
	node := report.MakeNode("foo").WithLatestControls(controls)
	expectedNode := node.WithControls("alive")
	rpt.Pod.AddNode(node)
	expected := report.MakeReport()
	expected.Pod.AddNode(expectedNode)
	got := rpt.BackwardCompatible()
	if !s_reflect.DeepEqual(expected, got) {
		t.Error(test.Diff(expected, got))
	}
}

func TestReportUpgrade(t *testing.T) {
	mtime.NowForce(time.Now())
	defer mtime.NowReset()
	node := report.MakeNode("foo").WithControls("alive")
	controls := map[string]report.NodeControlData{
		"alive": {
			Dead: false,
		},
	}
	expectedNode := node.WithLatestControls(controls)
	rpt := report.MakeReport()
	rpt.Pod.AddNode(node)
	expected := report.MakeReport()
	expected.Pod.AddNode(expectedNode)
	got := rpt.Upgrade()
	if !s_reflect.DeepEqual(expected, got) {
		t.Error(test.Diff(expected, got))
	}
}
