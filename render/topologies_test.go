package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test"
)

func TestProcessRenderer(t *testing.T) {
	have := sterilize(render.ProcessRenderer.Render(test.Report), true)
	want := expected.RenderedProcesses
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	have := sterilize(render.ProcessNameRenderer.Render(test.Report), true)
	want := expected.RenderedProcessNames
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerRenderer(t *testing.T) {
	have := sterilize(render.ContainerRenderer.Render(test.Report), true)
	want := expected.RenderedContainers
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageRenderer(t *testing.T) {
	have := sterilize(render.ContainerImageRenderer.Render(test.Report), true)
	want := expected.RenderedContainerImages
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestHostRenderer(t *testing.T) {
	have := sterilize(render.HostRenderer.Render(test.Report), true)
	want := expected.RenderedHosts
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
