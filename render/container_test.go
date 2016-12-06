package render_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

// FilterApplication is a Renderer which filters out application nodes.
func FilterApplication(r render.Renderer) render.Renderer {
	return render.MakeFilter(render.IsApplication, r)
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
	_, ipNet, err := net.ParseCIDR("1.2.3.0/16")
	if err != nil {
		t.Fatalf(err.Error())
	}
	localNetworks := report.Networks([]*net.IPNet{ipNet})
	if have := f(input.n, localNetworks); input.ok != (len(have) > 0) {
		name := input.name
		if name == "" {
			name = fmt.Sprintf("%v", input.n)
		}
		t.Errorf("%s: want %v, have %v", name, input.ok, have)
	}
}

func TestContainerRenderer(t *testing.T) {
	have := Prune(render.ContainerWithImageNameRenderer.Render(fixture.Report, FilterNoop))
	want := Prune(expected.RenderedContainers)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerFilterRenderer(t *testing.T) {
	// tag on of the containers in the topology and ensure
	// it is filtered out correctly.
	input := fixture.Report.Copy()
	input.Container.Nodes[fixture.ClientContainerNodeID] = input.Container.Nodes[fixture.ClientContainerNodeID].WithLatests(map[string]string{
		docker.LabelPrefix + "works.weave.role": "system",
	})
	have := Prune(render.ContainerWithImageNameRenderer.Render(input, FilterApplication))
	want := Prune(expected.RenderedContainers.Copy())
	delete(want, fixture.ClientContainerNodeID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerHostnameRenderer(t *testing.T) {
	have := Prune(render.ContainerHostnameRenderer.Render(fixture.Report, FilterNoop))
	want := Prune(expected.RenderedContainerHostnames)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerHostnameFilterRenderer(t *testing.T) {
	// add a system container into the topology and ensure
	// it is filtered out correctly.
	input := fixture.Report.Copy()

	clientContainer2ID := "f6g7h8i9j1"
	clientContainer2NodeID := report.MakeContainerNodeID(clientContainer2ID)
	input.Container.AddNode(report.MakeNodeWith(clientContainer2NodeID, map[string]string{
		docker.LabelPrefix + "works.weave.role": "system",
		docker.ContainerHostname:                fixture.ClientContainerHostname,
		report.HostNodeID:                       fixture.ClientHostNodeID,
	}).
		WithParents(report.EmptySets.
			Add("host", report.MakeStringSet(fixture.ClientHostNodeID)),
		).WithTopology(report.Container))

	have := Prune(render.ContainerHostnameRenderer.Render(input, FilterApplication))
	want := Prune(expected.RenderedContainerHostnames)
	// Test works by virtue of the RenderedContainerHostname only having a container
	// counter == 1

	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageRenderer(t *testing.T) {
	have := Prune(render.ContainerImageRenderer.Render(fixture.Report, FilterNoop))
	want := Prune(expected.RenderedContainerImages)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageFilterRenderer(t *testing.T) {
	// add a system container into the topology and ensure
	// it is filtered out correctly.
	input := fixture.Report.Copy()

	clientContainer2ID := "f6g7h8i9j1"
	clientContainer2NodeID := report.MakeContainerNodeID(clientContainer2ID)
	input.Container.AddNode(report.MakeNodeWith(clientContainer2NodeID, map[string]string{
		docker.LabelPrefix + "works.weave.role": "system",

		docker.ImageID:    fixture.ClientContainerImageID,
		docker.ImageName:  fixture.ClientContainerImageName,
		report.HostNodeID: fixture.ClientHostNodeID,
	}).
		WithParents(report.EmptySets.
			Add("host", report.MakeStringSet(fixture.ClientHostNodeID)),
		).WithTopology(report.ContainerImage))

	have := Prune(render.ContainerImageRenderer.Render(input, FilterApplication))
	want := Prune(expected.RenderedContainerImages.Copy())
	// Test works by virtue of the RenderedContainerImage only having a container
	// counter == 1

	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
