package render_test

import (
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
		{nrn(report.MakeNode()), false},
		{nrn(report.MakeNodeWith(map[string]string{endpoint.Addr: "1.2.3.4", endpoint.Procspied: "true"})), false},
		{nrn(report.MakeNodeWith(map[string]string{endpoint.Port: "1234", endpoint.Procspied: "true"})), false},
		{nrn(report.MakeNodeWith(map[string]string{endpoint.Addr: "1.2.3.4", endpoint.Port: "1234", endpoint.Procspied: "true"})), true},
		{nrn(report.MakeNodeWith(map[string]string{endpoint.Addr: "1.2.3.4", endpoint.Port: "40000", endpoint.Procspied: "true"})), true},
		{nrn(report.MakeNodeWith(map[string]string{report.HostNodeID: report.MakeHostNodeID("foo"), endpoint.Addr: "10.0.0.1", endpoint.Port: "20001", endpoint.Procspied: "true"})), true},
	} {
		testMap(t, render.MapEndpointIdentity, input)
	}
}

func TestMapProcessIdentity(t *testing.T) {
	for _, input := range []testcase{
		{nrn(report.MakeNode()), false},
		{nrn(report.MakeNodeWith(map[string]string{process.PID: "201"})), true},
	} {
		testMap(t, render.MapProcessIdentity, input)
	}
}

func TestMapContainerIdentity(t *testing.T) {
	for _, input := range []testcase{
		{nrn(report.MakeNode()), false},
		{nrn(report.MakeNodeWith(map[string]string{docker.ContainerID: "a1b2c3"})), true},
	} {
		testMap(t, render.MapContainerIdentity, input)
	}
}

func TestMapContainerImageIdentity(t *testing.T) {
	for _, input := range []testcase{
		{nrn(report.MakeNode()), false},
		{nrn(report.MakeNodeWith(map[string]string{docker.ImageID: "a1b2c3"})), true},
	} {
		testMap(t, render.MapContainerImageIdentity, input)
	}
}

func TestMapAddressIdentity(t *testing.T) {
	for _, input := range []testcase{
		{nrn(report.MakeNode()), false},
		{nrn(report.MakeNodeWith(map[string]string{endpoint.Addr: "192.168.1.1"})), true},
	} {
		testMap(t, render.MapAddressIdentity, input)
	}
}

func TestMapHostIdentity(t *testing.T) {
	for _, input := range []testcase{
		{nrn(report.MakeNode()), true}, // TODO it's questionable if this is actually correct
	} {
		testMap(t, render.MapHostIdentity, input)
	}
}

func TestMapPodIdentity(t *testing.T) {
	for _, input := range []testcase{
		{nrn(report.MakeNode()), false},
		{nrn(report.MakeNodeWith(map[string]string{kubernetes.PodID: "ping/pong", kubernetes.PodName: "pong"})), true},
	} {
		testMap(t, render.MapPodIdentity, input)
	}
}

func TestMapServiceIdentity(t *testing.T) {
	for _, input := range []testcase{
		{nrn(report.MakeNode()), false},
		{nrn(report.MakeNodeWith(map[string]string{kubernetes.ServiceID: "ping/pong", kubernetes.ServiceName: "pong"})), true},
	} {
		testMap(t, render.MapServiceIdentity, input)
	}
}

type testcase struct {
	md render.RenderableNode
	ok bool
}

func testMap(t *testing.T, f render.MapFunc, input testcase) {
	_, ipNet, err := net.ParseCIDR("1.2.3.0/16")
	if err != nil {
		t.Fatalf(err.Error())
	}
	localNetworks := report.Networks([]*net.IPNet{ipNet})
	if have := f(input.md, localNetworks); input.ok != (len(have) > 0) {
		t.Errorf("%v: want %v, have %v", input.md, input.ok, have)
	}
}
