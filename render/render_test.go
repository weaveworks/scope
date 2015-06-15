package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

type mockRenderer struct {
	nodes report.RenderableNodes
	amd   report.AggregateMetadata
}

func (m mockRenderer) Render(rpt report.Report) report.RenderableNodes { return m.nodes }
func (m mockRenderer) EdgeMetadata(rpt report.Report, localID, remoteID string) report.AggregateMetadata {
	return m.amd
}

func TestReduceRender(t *testing.T) {
	renderer := render.Reduce{
		mockRenderer{nodes: report.RenderableNodes{"foo": {ID: "foo"}}},
		mockRenderer{nodes: report.RenderableNodes{"bar": {ID: "bar"}}},
	}

	want := report.RenderableNodes{"foo": {ID: "foo"}, "bar": {ID: "bar"}}
	have := renderer.Render(report.MakeReport())

	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestReduceEdge(t *testing.T) {
	renderer := render.Reduce{
		mockRenderer{amd: report.AggregateMetadata{"foo": 1}},
		mockRenderer{amd: report.AggregateMetadata{"bar": 2}},
	}

	want := report.AggregateMetadata{"foo": 1, "bar": 2}
	have := renderer.EdgeMetadata(report.MakeReport(), "", "")

	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}
