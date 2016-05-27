package render_test

import (
	"testing"

	"$GITHUB_URI/render"
	"$GITHUB_URI/render/expected"
	"$GITHUB_URI/test"
	"$GITHUB_URI/test/fixture"
	"$GITHUB_URI/test/reflect"
)

func TestHostRenderer(t *testing.T) {
	have := Prune(render.HostRenderer.Render(fixture.Report, render.FilterNoop))
	want := Prune(expected.RenderedHosts)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
