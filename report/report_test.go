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
		node := report.MakeNodeWith("foo", map[string]string{
			"foo": "bar",
		})

		if v, _ := node.Latest.Lookup("foo"); v != "bar" {
			t.Errorf("want foo, have %s", v)
		}
	}
	{
		node := report.MakeNode("foo").WithCounter("foo", 1)
		if value, _ := node.LookupCounter("foo"); value != 1 {
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
		WithLatests(map[string]string{report.KubernetesNamespace: namespaceName}).
		WithParents(report.MakeSets().Add(report.ReplicaSet, report.MakeStringSet("bar")))
	expectedPodNode := podNode.PruneParents().WithParents(parentsWithDeployment)
	rpt := report.MakeReport()
	rpt.ReplicaSet.AddNode(rsNode)
	rpt.Pod.AddNode(podNode)
	namespaceNode := report.MakeNode(namespaceID).
		WithLatests(map[string]string{report.KubernetesName: namespaceName})
	expected := report.MakeReport()
	expected.ReplicaSet.AddNode(rsNode)
	expected.Pod.AddNode(expectedPodNode)
	expected.Namespace.AddNode(namespaceNode)
	got := rpt.Upgrade()
	if !s_reflect.DeepEqual(expected, got) {
		t.Error(test.Diff(expected, got))
	}
}

func TestReportUnMerge(t *testing.T) {
	n1 := report.MakeNodeWith("foo", map[string]string{"foo": "bar"})
	r1 := makeTestReport()
	r2 := r1.Copy()
	r2.Container.AddNode(n1)
	// r2 should be the same as r1 with just the foo-bar node added
	r2.UnsafeUnMerge(r1)
	// Now r2 should have everything removed except that one node, and its ID
	expected := report.Report{
		ID: r2.ID,
		Container: report.Topology{
			Nodes: report.Nodes{
				"foo": n1,
			},
		},
	}

	// Now test report with two nodes unmerged on report with one
	r1.Container.AddNode(n1)
	r2 = r1.Copy()
	n2 := report.MakeNodeWith("foo2", map[string]string{"ping": "pong"})
	r2.Container.AddNode(n2)
	// r2 should be the same as r1 with one extra node
	r2.UnsafeUnMerge(r1)
	expected = report.Report{
		ID: r2.ID,
		Container: report.Topology{
			Nodes: report.Nodes{
				"foo2": n2,
			},
		},
	}

	if !s_reflect.DeepEqual(expected, r2) {
		t.Error(test.Diff(expected, r2))
	}
}
