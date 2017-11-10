package render_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
	"github.com/weaveworks/scope/test/utils"
)

// FilterApplication is a Renderer which filters out application nodes.
func FilterApplication(r render.Renderer) render.Renderer {
	return render.MakeFilter(render.IsApplication, r)
}

// FilterSystem is a Renderer which filters out system nodes.
func FilterSystem(r render.Renderer) render.Renderer {
	return render.MakeFilter(render.IsSystem, r)
}

// FilterNoop does nothing.
func FilterNoop(r render.Renderer) render.Renderer { return r }

func TestMapProcess2Container(t *testing.T) {
	for _, input := range []testcase{
		{"empty", report.MakeNode("empty"), true},
		{"basic process", report.MakeNodeWith("basic", map[string]string{process.PID: "201", docker.ContainerID: "a1b2c3"}), true},
		{"uncontained", report.MakeNodeWith("uncontained", map[string]string{process.PID: "201", report.HostNodeID: report.MakeHostNodeID("foo")}), true},
	} {
		testMap(t, render.MapProcess2Container, input)
	}
}

type testcase struct {
	name string
	n    report.Node
	ok   bool
}

func testMap(t *testing.T, f render.MapFunc, input testcase) {
	localNetworks := report.MakeNetworks()
	if err := localNetworks.AddCIDR("1.2.3.0/16"); err != nil {
		t.Fatalf(err.Error())
	}
	if have := f(input.n, localNetworks); input.ok != (len(have) > 0) {
		name := input.name
		if name == "" {
			name = fmt.Sprintf("%v", input.n)
		}
		t.Errorf("%s: want %v, have %v", name, input.ok, have)
	}
}

func TestContainerRenderer(t *testing.T) {
	have := utils.Prune(render.ContainerWithImageNameRenderer.Render(fixture.Report, FilterNoop).Nodes)
	want := utils.Prune(expected.RenderedContainers)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerFilterRenderer(t *testing.T) {
	// tag on of the containers in the topology and ensure
	// it is filtered out correctly.
	input := fixture.Report.Copy()
	renderer := render.ApplyDecorator(render.ContainerWithImageNameRenderer)

	input.Container.Nodes[fixture.ClientContainerNodeID] = input.Container.Nodes[fixture.ClientContainerNodeID].WithLatests(map[string]string{
		docker.LabelPrefix + "works.weave.role": "system",
	})

	have := utils.Prune(renderer.Render(input, FilterApplication).Nodes)
	want := utils.Prune(expected.RenderedContainers.Copy())
	delete(want, fixture.ClientContainerNodeID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerHostnameRenderer(t *testing.T) {
	renderer := render.ApplyDecorator(render.ContainerHostnameRenderer)
	have := utils.Prune(renderer.Render(fixture.Report, FilterNoop).Nodes)
	want := utils.Prune(expected.RenderedContainerHostnames)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerHostnameFilterRenderer(t *testing.T) {
	renderer := render.ApplyDecorator(render.ContainerHostnameRenderer)
	have := utils.Prune(renderer.Render(fixture.Report, FilterSystem).Nodes)
	want := utils.Prune(expected.RenderedContainerHostnames.Copy())
	delete(want, fixture.ClientContainerHostname)
	delete(want, fixture.ServerContainerHostname)
	delete(want, render.IncomingInternetID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageRenderer(t *testing.T) {
	renderer := render.ApplyDecorator(render.ContainerImageRenderer)
	have := utils.Prune(renderer.Render(fixture.Report, FilterNoop).Nodes)
	want := utils.Prune(expected.RenderedContainerImages)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageFilterRenderer(t *testing.T) {
	renderer := render.ApplyDecorator(render.ContainerImageRenderer)
	have := utils.Prune(renderer.Render(fixture.Report, FilterSystem).Nodes)
	want := utils.Prune(expected.RenderedContainerHostnames.Copy())
	delete(want, fixture.ClientContainerHostname)
	delete(want, fixture.ServerContainerHostname)
	delete(want, render.IncomingInternetID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
