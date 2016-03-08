package render_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

func nrn(nmd report.Node) render.RenderableNode {
	return render.NewRenderableNode("").WithNode(nmd)
}

func TestMapEndpointIdentity(t *testing.T) {
	for _, input := range []testcase{
		{"empty", nrn(report.MakeNode()), false},
		{"", nrn(report.MakeNodeWith(map[string]string{endpoint.Addr: "1.2.3.4", endpoint.Procspied: "true"})), false},
		{"", nrn(report.MakeNodeWith(map[string]string{endpoint.Port: "1234", endpoint.Procspied: "true"})), false},
		{"", nrn(report.MakeNodeWith(map[string]string{endpoint.Addr: "1.2.3.4", endpoint.Port: "1234", endpoint.Procspied: "true"})), true},
		{"", nrn(report.MakeNodeWith(map[string]string{endpoint.Addr: "1.2.3.4", endpoint.Port: "40000", endpoint.Procspied: "true"})), true},
		{"", nrn(report.MakeNodeWith(map[string]string{report.HostNodeID: report.MakeHostNodeID("foo"), endpoint.Addr: "10.0.0.1", endpoint.Port: "20001", endpoint.Procspied: "true"})), true},
	} {
		testMap(t, render.MapEndpointIdentity, input)
	}
}

func TestMapProcessIdentity(t *testing.T) {
	for _, input := range []testcase{
		{"empty", nrn(report.MakeNode()), false},
		{"basic process", nrn(report.MakeNodeWith(map[string]string{process.PID: "201"})), true},
	} {
		testMap(t, render.MapProcessIdentity, input)
	}
}

func TestMapProcess2Container(t *testing.T) {
	for _, input := range []testcase{
		{"empty", nrn(report.MakeNode()), true},
		{"basic process", nrn(report.MakeNodeWith(map[string]string{process.PID: "201", docker.ContainerID: "a1b2c3"})), true},
		{"uncontained", nrn(report.MakeNodeWith(map[string]string{process.PID: "201", report.HostNodeID: report.MakeHostNodeID("foo")})), true},
	} {
		testMap(t, render.MapProcess2Container, input)
	}
}

func TestMapContainerIdentity(t *testing.T) {
	for _, input := range []testcase{
		{"empty", nrn(report.MakeNode()), false},
		{"basic container", nrn(report.MakeNodeWith(map[string]string{docker.ContainerID: "a1b2c3"})), true},
	} {
		testMap(t, render.MapContainerIdentity, input)
	}
}

func TestMapContainerImageIdentity(t *testing.T) {
	for _, input := range []testcase{
		{"empty", nrn(report.MakeNode()), false},
		{"basic image", nrn(report.MakeNodeWith(map[string]string{docker.ImageID: "a1b2c3"})), true},
	} {
		testMap(t, render.MapContainerImageIdentity, input)
	}
}

func TestMapHostIdentity(t *testing.T) {
	for _, input := range []testcase{
		{"empty", nrn(report.MakeNode()), true}, // TODO it's questionable if this is actually correct
	} {
		testMap(t, render.MapHostIdentity, input)
	}
}

func TestMapPodIdentity(t *testing.T) {
	for _, input := range []testcase{
		{"empty", nrn(report.MakeNode()), false},
		{"basic pod", nrn(report.MakeNodeWith(map[string]string{kubernetes.PodID: "ping/pong", kubernetes.PodName: "pong"})), true},
	} {
		testMap(t, render.MapPodIdentity, input)
	}
}

func TestMapServiceIdentity(t *testing.T) {
	for _, input := range []testcase{
		{"empty", nrn(report.MakeNode()), false},
		{"basic service", nrn(report.MakeNodeWith(map[string]string{kubernetes.ServiceID: "ping/pong", kubernetes.ServiceName: "pong"})), true},
	} {
		testMap(t, render.MapServiceIdentity, input)
	}
}

type testcase struct {
	name string
	n    render.RenderableNode
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
