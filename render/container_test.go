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

var (
	filterApplication = render.Transformers([]render.Transformer{
		render.AnyFilterFunc(render.IsPseudoTopology, render.IsApplication),
		render.FilterUnconnectedPseudo,
	})
	filterSystem = render.Transformers([]render.Transformer{
		render.AnyFilterFunc(render.IsPseudoTopology, render.IsSystem),
		render.FilterUnconnectedPseudo,
	})
)

func TestMapProcess2Container(t *testing.T) {
	for _, input := range []testcase{
		{"empty", report.MakeNode("empty"), true},
		{"basic process", report.MakeNodeWith("basic", process.PID, "201", docker.ContainerID, "a1b2c3"), true},
		{"uncontained", report.MakeNodeWith("uncontained", process.PID, "201", report.HostNodeID, report.MakeHostNodeID("foo")), true},
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
	if have := f(input.n); input.ok != (have.ID != "") {
		name := input.name
		if name == "" {
			name = fmt.Sprintf("%v", input.n)
		}
		t.Errorf("%s: want %v, have %v", name, input.ok, have)
	}
}

func TestContainerRenderer(t *testing.T) {
	have := utils.Prune(render.ContainerWithImageNameRenderer.Render(fixture.Report).Nodes)
	want := utils.Prune(expected.RenderedContainers)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerFilterRenderer(t *testing.T) {
	// tag on of the containers in the topology and ensure
	// it is filtered out correctly.
	input := fixture.Report.Copy()
	input.Container.Nodes[fixture.ClientContainerNodeID] = input.Container.Nodes[fixture.ClientContainerNodeID].WithLatests(
		docker.LabelPrefix+"works.weave.role", "system",
	)
	have := utils.Prune(render.Render(input, render.ContainerWithImageNameRenderer, filterApplication).Nodes)
	want := utils.Prune(expected.RenderedContainers.Copy())
	delete(want, fixture.ClientContainerNodeID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerHostnameRenderer(t *testing.T) {
	have := utils.Prune(render.Render(fixture.Report, render.ContainerHostnameRenderer, render.Transformers(nil)).Nodes)
	want := utils.Prune(expected.RenderedContainerHostnames)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerHostnameFilterRenderer(t *testing.T) {
	have := utils.Prune(render.Render(fixture.Report, render.ContainerHostnameRenderer, filterSystem).Nodes)
	want := utils.Prune(expected.RenderedContainerHostnames.Copy())
	delete(want, fixture.ClientContainerHostname)
	delete(want, fixture.ServerContainerHostname)
	delete(want, render.IncomingInternetID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageRenderer(t *testing.T) {
	have := utils.Prune(render.Render(fixture.Report, render.ContainerImageRenderer, render.Transformers(nil)).Nodes)
	want := utils.Prune(expected.RenderedContainerImages)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageFilterRenderer(t *testing.T) {
	have := utils.Prune(render.Render(fixture.Report, render.ContainerImageRenderer, filterSystem).Nodes)
	want := utils.Prune(expected.RenderedContainerHostnames.Copy())
	delete(want, fixture.ClientContainerHostname)
	delete(want, fixture.ServerContainerHostname)
	delete(want, render.IncomingInternetID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
