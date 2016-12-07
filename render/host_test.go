package render_test

import (
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

func TestHostRenderer(t *testing.T) {
	have := Prune(render.HostRenderer.Render(fixture.Report, FilterNoop))
	want := Prune(expected.RenderedHosts)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
