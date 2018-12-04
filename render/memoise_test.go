package render_test

import (
	"context"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

type renderFunc func(ctx context.Context, r report.Report) render.Nodes

func (f renderFunc) Render(ctx context.Context, r report.Report) render.Nodes { return f(ctx, r) }

func TestMemoise(t *testing.T) {
	calls := 0
	r := renderFunc(func(ctx context.Context, rpt report.Report) render.Nodes {
		calls++
		return render.Nodes{Nodes: report.Nodes{rpt.ID: report.MakeNode(rpt.ID)}}
	})
	m := render.Memoise(r)

	if render.Memoise(m) != m {
		t.Errorf("Memoised renderers should be fixpoints.")
	}

	rpt1 := report.MakeReport()

	ctx := context.Background()
	result1 := m.Render(ctx, rpt1)
	// it should have rendered it.
	if _, ok := result1.Nodes[rpt1.ID]; !ok {
		t.Errorf("Expected rendered report to contain a node, but got: %v", result1)
	}
	if calls != 1 {
		t.Errorf("Expected renderer to have been called the first time")
	}

	result2 := m.Render(ctx, rpt1)
	if !reflect.DeepEqual(result1, result2) {
		t.Errorf("Expected memoised result to be returned: %s", test.Diff(result1, result2))
	}
	if calls != 1 {
		t.Errorf("Expected renderer to not have been called the second time")
	}

	rpt2 := report.MakeReport()
	result3 := m.Render(ctx, rpt2)
	if reflect.DeepEqual(result1, result3) {
		t.Errorf("Expected different result for different report, but were the same")
	}
	if calls != 2 {
		t.Errorf("Expected renderer to have been called again for a different report")
	}

	render.ResetCache()
	result4 := m.Render(ctx, rpt1)
	if !reflect.DeepEqual(result1, result4) {
		t.Errorf("Expected original result to be returned: %s", test.Diff(result1, result4))
	}
	if calls != 3 {
		t.Errorf("Expected renderer to have been called again after cache reset")
	}
}
