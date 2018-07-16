package report_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/report"
	s_reflect "github.com/weaveworks/scope/test/reflect"
)

func newu64(value uint64) *uint64 { return &value }

// Make sure we don't add a topology and miss it in the Topologies method.
func TestReportTopologies(t *testing.T) {
	var (
		reportType   = reflect.TypeOf(report.MakeReport())
		topologyType = reflect.TypeOf(report.MakeTopology())
	)

	var want, have int
	for i := 0; i < reportType.NumField(); i++ {
		if reportType.Field(i).Type == topologyType {
			want++
		}
	}

	r := report.MakeReport()
	r.WalkTopologies(func(_ *report.Topology) {
		have++
	})
	if want != have {
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
		node := report.MakeNodeWith("foo",
			"foo", "bar",
		)

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
}

func TestReportUpgrade(t *testing.T) {
	mtime.NowForce(time.Now())
	defer mtime.NowReset()
	parentsWithDeployment := report.MakeSets().Add(report.Deployment, report.MakeStringSet("id"))
	rsNode := report.MakeNode("bar").
		WithParents(parentsWithDeployment)
	namespaceName := "ns"
	namespaceID := report.MakeNamespaceNodeID(namespaceName)
	podNode := report.MakeNode("foo").
		WithLatests(report.KubernetesNamespace, namespaceName).
		WithControls("alive").
		WithParents(report.MakeSets().Add(report.ReplicaSet, report.MakeStringSet("bar")))
	controls := map[string]report.NodeControlData{
		"alive": {
			Dead: false,
		},
	}
	expectedPodNode := podNode.PruneParents().WithParents(parentsWithDeployment).WithLatestControls(controls)
	rpt := report.MakeReport()
	rpt.ReplicaSet.AddNode(rsNode)
	rpt.Pod.AddNode(podNode)
	namespaceNode := report.MakeNode(namespaceID).
		WithLatests(report.KubernetesName, namespaceName)
	expected := report.MakeReport()
	expected.ReplicaSet.AddNode(rsNode)
	expected.Pod.AddNode(expectedPodNode)
	expected.Namespace.AddNode(namespaceNode)
	got := rpt.Upgrade()
	if !s_reflect.DeepEqual(expected, got) {
		t.Error(test.Diff(expected, got))
	}
}
