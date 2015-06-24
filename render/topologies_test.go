package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test"
)

func trimNodeMetadata(rns render.RenderableNodes) render.RenderableNodes {
	result := render.RenderableNodes{}
	for id, rn := range rns {
		rn.NodeMetadata = nil
		result[id] = rn
	}
	return result
}

func TestProcessRenderer(t *testing.T) {
	have := render.ProcessRenderer.Render(test.Report)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(expected.RenderedProcesses, have) {
		t.Error("\n" + test.Diff(expected.RenderedProcesses, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	have := render.ProcessNameRenderer.Render(test.Report)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(expected.RenderedProcessNames, have) {
		t.Error("\n" + test.Diff(expected.RenderedProcessNames, have))
	}
}

func TestContainerRenderer(t *testing.T) {
	have := render.ContainerRenderer.Render(test.Report)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(expected.RenderedContainers, have) {
		t.Error("\n" + test.Diff(expected.RenderedContainers, have))
	}
}

func TestContainerImageRenderer(t *testing.T) {
	have := render.ContainerImageRenderer.Render(test.Report)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(expected.RenderedContainerImages, have) {
		t.Error("\n" + test.Diff(expected.RenderedContainerImages, have))
	}
}

func TestHostRenderer(t *testing.T) {
	have := render.HostRenderer.Render(test.Report)
	have = trimNodeMetadata(have)
	if !reflect.DeepEqual(expected.RenderedHosts, have) {
		t.Error("\n" + test.Diff(expected.RenderedHosts, have))
	}
}
