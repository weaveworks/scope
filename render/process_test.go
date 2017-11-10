package render_test

import (
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
	"github.com/weaveworks/scope/test/utils"
)

func TestEndpointRenderer(t *testing.T) {
	have := utils.Prune(render.EndpointRenderer.Render(fixture.Report, FilterNoop).Nodes)
	want := utils.Prune(expected.RenderedEndpoints)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessRenderer(t *testing.T) {
	have := utils.Prune(render.ProcessRenderer.Render(fixture.Report, FilterNoop).Nodes)
	want := utils.Prune(expected.RenderedProcesses)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	have := utils.Prune(render.ProcessNameRenderer.Render(fixture.Report, FilterNoop).Nodes)
	want := utils.Prune(expected.RenderedProcessNames)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
