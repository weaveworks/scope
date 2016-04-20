package render_test

import (
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

func TestEndpointRenderer(t *testing.T) {
	have := render.EndpointRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedEndpoints.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessRenderer(t *testing.T) {
	have := render.ProcessRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedProcesses.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	have := render.ProcessNameRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedProcessNames.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
