package render_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
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
