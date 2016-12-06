package render_test

import (
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

func TestEndpointRenderer(t *testing.T) {
	have := Prune(render.EndpointRenderer.Render(fixture.Report, FilterNoop))
	want := Prune(expected.RenderedEndpoints)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessRenderer(t *testing.T) {
	have := Prune(render.ProcessRenderer.Render(fixture.Report, FilterNoop))
	want := Prune(expected.RenderedProcesses)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	have := Prune(render.ProcessNameRenderer.Render(fixture.Report, FilterNoop))
	want := Prune(expected.RenderedProcessNames)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
