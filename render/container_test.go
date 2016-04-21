package render_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

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
	have := Prune(render.ContainerRenderer.Render(fixture.Report))
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
	have := Prune(render.FilterSystem(render.ContainerRenderer).Render(input))
	want := Prune(expected.RenderedContainers.Copy())
	delete(want, fixture.ClientContainerNodeID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerWithHostIPsRenderer(t *testing.T) {
	input := fixture.Report.Copy()
	input.Container.Nodes[fixture.ClientContainerNodeID] = input.Container.Nodes[fixture.ClientContainerNodeID].WithLatests(map[string]string{
		docker.ContainerNetworkMode: "host",
	})
	nodes := render.ContainerWithHostIPsRenderer.Render(input)

	// Test host network nodes get the host IPs added.
	haveNode, ok := nodes[fixture.ClientContainerNodeID]
	if !ok {
		t.Fatal("Expected output to have the client container node")
	}
	have, ok := haveNode.Sets.Lookup(docker.ContainerIPs)
	if !ok {
		t.Fatal("Container had no IPs set.")
	}
	want := report.MakeStringSet("10.10.10.0")
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageRenderer(t *testing.T) {
	have := Prune(render.ContainerImageRenderer.Render(fixture.Report))
	want := Prune(expected.RenderedContainerImages)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
