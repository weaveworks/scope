package render_test

import (
	"testing"

	"$GITHUB_URI/render"
	"$GITHUB_URI/render/expected"
	"$GITHUB_URI/test"
	"$GITHUB_URI/test/fixture"
	"$GITHUB_URI/test/reflect"
)

func TestEndpointRenderer(t *testing.T) {
	have := Prune(render.EndpointRenderer.Render(fixture.Report, render.FilterNoop))
	want := Prune(expected.RenderedEndpoints)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessRenderer(t *testing.T) {
	have := Prune(render.ProcessRenderer.Render(fixture.Report, render.FilterNoop))
	want := Prune(expected.RenderedProcesses)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	have := Prune(render.ProcessNameRenderer.Render(fixture.Report, render.FilterNoop))
	want := Prune(expected.RenderedProcessNames)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
