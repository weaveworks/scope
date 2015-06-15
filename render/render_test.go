package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

type mockRenderer struct {
	report.RenderableNodes
	aggregateMetadata report.AggregateMetadata
}

func (m mockRenderer) Render(rpt report.Report) report.RenderableNodes {
	return m.RenderableNodes
}
func (m mockRenderer) AggregateMetadata(rpt report.Report, localID, remoteID string) report.AggregateMetadata {
	return m.aggregateMetadata
}

func TestReduceRender(t *testing.T) {
	renderer := render.Reduce([]render.Renderer{
		mockRenderer{RenderableNodes: report.RenderableNodes{"foo": {ID: "foo"}}},
		mockRenderer{RenderableNodes: report.RenderableNodes{"bar": {ID: "bar"}}},
	})

	want := report.RenderableNodes{"foo": {ID: "foo"}, "bar": {ID: "bar"}}
	have := renderer.Render(report.MakeReport())

	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestReduceEdge(t *testing.T) {
	renderer := render.Reduce([]render.Renderer{
		mockRenderer{aggregateMetadata: report.AggregateMetadata{"foo": 1}},
		mockRenderer{aggregateMetadata: report.AggregateMetadata{"bar": 2}},
	})

	want := report.AggregateMetadata{"foo": 1, "bar": 2}
	have := renderer.AggregateMetadata(report.MakeReport(), "", "")

	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}
