package dsl

import (
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"

	"github.com/weaveworks/scope/probe/endpoint"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

func TestSelectAll(t *testing.T) {
	rns := render.RenderableNodes{
		"a": render.NewRenderableNode("a"),
		"c": render.NewRenderableNode("c"),
		"b": render.NewRenderableNode("b"),
	}
	testStringSlice(t, []string{"a", "b", "c"}, selectAll(rns))
}

func TestSelectConnected(t *testing.T) {
	rns := render.RenderableNodes{
		"a": render.NewRenderableNode("a").WithNode(report.MakeNode().WithAdjacency(report.MakeIDList("m", "b", "c"))),
		"b": render.NewRenderableNode("b").WithNode(report.MakeNode().WithAdjacency(report.MakeIDList("m", "c"))),
		"c": render.NewRenderableNode("c").WithNode(report.MakeNode().WithAdjacency(report.MakeIDList("m"))),
		"m": render.NewRenderableNode("m").WithNode(report.MakeNode().WithAdjacency(report.MakeIDList("n"))),
		"n": render.NewRenderableNode("n"),
		"x": render.NewRenderableNode("x"),
		"y": render.NewRenderableNode("y"),
		"z": render.NewRenderableNode("z"),
	}
	testStringSlice(t, []string{"a", "b", "c", "m", "n"}, selectConnected(rns))
}

func TestSelectNonlocal(t *testing.T) {
	rns := render.RenderableNodes{
		"a": render.NewRenderableNode("a").WithNode(report.MakeNodeWith(map[string]string{host.LocalNetworks: "10.10.1.0/24"})),
		"b": render.NewRenderableNode("b").WithNode(report.MakeNodeWith(map[string]string{host.LocalNetworks: "10.10.2.0/24"})),
		"c": render.NewRenderableNode("c").WithNode(report.MakeNodeWith(map[string]string{host.LocalNetworks: "10.20.0.0/16"})),

		// Test selecting nonlocal via endpoint.Addr metadata key
		"x1": render.NewRenderableNode("x1").WithNode(report.MakeNodeWith(map[string]string{endpoint.Addr: "10.10.1.34"})),  // local
		"x2": render.NewRenderableNode("x2").WithNode(report.MakeNodeWith(map[string]string{endpoint.Addr: "10.20.99.1"})),  // local
		"x3": render.NewRenderableNode("x3").WithNode(report.MakeNodeWith(map[string]string{endpoint.Addr: "10.10.3.33"})),  // nonlocal
		"x4": render.NewRenderableNode("x4").WithNode(report.MakeNodeWith(map[string]string{endpoint.Addr: "192.168.1.1"})), // nonlocal
	}

	// Test selecting nonlocal via parsing the node ID
	var (
		local1    = report.MakeAddressNodeID("some-host", "10.10.2.2")
		nonlocal1 = report.MakeAddressNodeID("some-host", "10.11.12.13")
		local2    = report.MakeEndpointNodeID("some-host", "10.20.0.1", "4040")
		nonlocal2 = report.MakeEndpointNodeID("some-host", "10.21.32.43", "8080")
	)
	rns[local1] = render.NewRenderableNode(local1).WithNode(report.MakeNode())
	rns[nonlocal1] = render.NewRenderableNode(nonlocal1).WithNode(report.MakeNode())
	rns[local2] = render.NewRenderableNode(local2).WithNode(report.MakeNode())
	rns[nonlocal2] = render.NewRenderableNode(nonlocal2).WithNode(report.MakeNode())

	testStringSlice(t, []string{"x3", "x4", nonlocal1, nonlocal2}, selectNonlocal(rns))
}

func TestSelectLike(t *testing.T) {
	rns := render.RenderableNodes{
		"abcfooxyz": render.NewRenderableNode("abcfooxyz"),
		"abcfoxyz":  render.NewRenderableNode("abcfoxyz"),
		"fo":        render.NewRenderableNode("fo"),
		"foo":       render.NewRenderableNode("foo"),
		"fooo":      render.NewRenderableNode("fooo"),
		"x_foo_y":   render.NewRenderableNode("x_foo_y"),
	}
	testStringSlice(t, []string{"abcfooxyz", "foo", "fooo", "x_foo_y"}, selectLike(`.*foo.*`)(rns))
}

func TestSelectWith(t *testing.T) {
	rns := render.RenderableNodes{
		"a": render.NewRenderableNode("a").WithNode(report.MakeNodeWith(map[string]string{"foo": "bar"})),
		"b": render.NewRenderableNode("b").WithNode(report.MakeNodeWith(map[string]string{"foo": "bar"})),
		"c": render.NewRenderableNode("c").WithNode(report.MakeNodeWith(map[string]string{"foo": "baz"})),
		"x": render.NewRenderableNode("x").WithNode(report.MakeNodeWith(map[string]string{"qux": "qix"})),
		"y": render.NewRenderableNode("y").WithNode(report.MakeNode()),
	}
	testStringSlice(t, []string{"a", "b", "c"}, selectWith("foo")(rns))
	testStringSlice(t, []string{"a", "b"}, selectWith("foo=bar")(rns))
	testStringSlice(t, []string{"x"}, selectWith("qux")(rns))
	testStringSlice(t, []string{"x"}, selectWith("qux=qix")(rns))
	testStringSlice(t, []string{}, selectWith("qux=XXX")(rns))
}

func TestSelectNot(t *testing.T) {
	rns := render.RenderableNodes{
		"a": render.NewRenderableNode("a").WithNode(report.MakeNodeWith(map[string]string{"foo": "1"})),
		"b": render.NewRenderableNode("b").WithNode(report.MakeNodeWith(map[string]string{"foo": "2"})),
		"c": render.NewRenderableNode("c").WithNode(report.MakeNodeWith(map[string]string{"foo": "3"})),
		"x": render.NewRenderableNode("x").WithNode(report.MakeNodeWith(map[string]string{"bar": "1"})),
		"y": render.NewRenderableNode("y").WithNode(report.MakeNodeWith(map[string]string{"bar": "2"})),
		"z": render.NewRenderableNode("z").WithNode(report.MakeNodeWith(map[string]string{"bar": "3"})),
	}
	testStringSlice(t, []string{"x", "y", "z"}, selectNot(selectWith("foo"))(rns))
	testStringSlice(t, []string{"a", "b", "c"}, selectNot(selectWith("bar"))(rns))
	testStringSlice(t, []string{}, selectNot(selectAll)(rns))
}

func TestTransformHighlight(t *testing.T) {
	rns := render.RenderableNodes{"a": render.NewRenderableNode("a")}
	get := func(rns render.RenderableNodes) string { return rns["a"].Metadata[highlightKey] }
	out := transformHighlight(rns, selectAll(rns))
	if want, have := "true", get(out); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestTransformRemove(t *testing.T) {
	rns := render.RenderableNodes{
		"a": render.NewRenderableNode("a").WithNode(report.MakeNodeWith(map[string]string{"foo": "1"}).WithAdjacency(report.MakeIDList("b", "c"))),
		"b": render.NewRenderableNode("b").WithNode(report.MakeNodeWith(map[string]string{"bar": "1"}).WithAdjacency(report.MakeIDList("c"))),
		"c": render.NewRenderableNode("c").WithNode(report.MakeNodeWith(map[string]string{"baz": "1"})),
	}

	// Remove all should totally kill the RenderableNodes
	if want, have := 0, len(transformRemove(rns, selectAll(rns))); want != have {
		t.Errorf("remove all: want %d, have %d", want, have)
	}

	// Removing c should kill the adjacency links to it
	out := transformRemove(rns, selectLike("c")(rns))
	if want, have := 2, len(out); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := 1, len(out["a"].Node.Adjacency); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := 0, len(out["b"].Node.Adjacency); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestTransformShowOnly(t *testing.T) {
	rns := render.RenderableNodes{
		"a": render.NewRenderableNode("a").WithNode(report.MakeNodeWith(map[string]string{"foo": "1"}).WithAdjacency(report.MakeIDList("b", "c"))),
		"b": render.NewRenderableNode("b").WithNode(report.MakeNodeWith(map[string]string{"bar": "1"}).WithAdjacency(report.MakeIDList("c"))),
		"c": render.NewRenderableNode("c").WithNode(report.MakeNodeWith(map[string]string{"baz": "1"})),
	}

	// Show only b should eliminate a, c, and the link to c
	out := transformShowOnly(rns, selectLike("b")(rns))
	if want, have := 1, len(out); want != have {
		t.Errorf("want %d, have %d (%#+v)", want, have, out)
	}
	if want, have := 0, len(out["b"].Node.Adjacency); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestTransformMerge(t *testing.T) {
	rns := render.RenderableNodes{
		"a": render.NewRenderableNode("a").WithNode(report.MakeNodeWith(map[string]string{"foo": "1"}).WithAdjacency(report.MakeIDList("b", "c"))),
		"b": render.NewRenderableNode("b").WithNode(report.MakeNodeWith(map[string]string{"bar": "1"}).WithAdjacency(report.MakeIDList("c"))),
		"c": render.NewRenderableNode("c").WithNode(report.MakeNodeWith(map[string]string{"baz": "1"})),
	}

	name := "supernode"
	out := transformMerge(name)(rns, selectAll(rns))
	if want, have := 1, len(out); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := 3, len(out[name].Node.Metadata); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := 1, len(out[name].Node.Adjacency); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
	if want, have := name, out[name].Node.Adjacency[0]; want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestTransformGroupBy(t *testing.T) {
	rns := render.RenderableNodes{
		"a": render.NewRenderableNode("a").WithNode(report.MakeNodeWith(map[string]string{"foo": "1"}).WithAdjacency(report.MakeIDList("b")).WithCounters(map[string]int{"c": 1})),
		"b": render.NewRenderableNode("b").WithNode(report.MakeNodeWith(map[string]string{"foo": "1"}).WithAdjacency(report.MakeIDList("c")).WithCounters(map[string]int{"c": 2})),
		"c": render.NewRenderableNode("c").WithNode(report.MakeNodeWith(map[string]string{"foo": "1"}).WithAdjacency(report.MakeIDList("d")).WithCounters(map[string]int{"c": 4})),
		"d": render.NewRenderableNode("d").WithNode(report.MakeNodeWith(map[string]string{"foo": "2"}).WithCounters(map[string]int{"c": 8})),
	}

	out := transformGroupBy("foo")(rns, selectAll(rns))
	if want, have := 2, len(out); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if _, ok := out["foo-1"]; !ok {
		t.Fatalf("missing merged node")
	}
	if _, ok := out["foo-2"]; !ok {
		t.Fatalf("missing unmerged node")
	}
	if want, have := 2, len(out["foo-1"].Node.Adjacency); want != have { // self-edge, and link to the other
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := 1+2+4, out["foo-1"].Node.Counters["c"]; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	if want, have := 8, out["foo-2"].Node.Counters["c"]; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestTransformJoin(t *testing.T) {
	const key, value = "😇", "😈"
	rns := render.RenderableNodes{
		"a": render.NewRenderableNode("a").WithNode(report.MakeNodeWith(map[string]string{"foo_node": "d", "1": "1"}).WithCounters(map[string]int{"c": 1})),
		"b": render.NewRenderableNode("b").WithNode(report.MakeNodeWith(map[string]string{"foo_node": "d", "2": "2"}).WithCounters(map[string]int{"c": 2})),
		"c": render.NewRenderableNode("c").WithNode(report.MakeNodeWith(map[string]string{"foo_node": "x", "4": "4"}).WithCounters(map[string]int{"c": 4})),
		"d": render.NewRenderableNode("d").WithNode(report.MakeNodeWith(map[string]string{key: value}).WithCounters(map[string]int{"c": 8})),
	}

	out := transformJoin("foo_node")(rns, selectAll(rns))
	if want, have := 3, len(out); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
	for _, id := range []string{"a", "b"} {
		if want, have := value, out[id].Node.Metadata[key]; want != have {
			t.Errorf("%s[%q]: want %q, have %q", id, key, want, have)
		}
	}
	if want, have := 1+8, out["a"].Node.Counters["c"]; want != have {
		t.Errorf("a: want %d, have %d", want, have)
	}
	if want, have := 2+8, out["b"].Node.Counters["c"]; want != have {
		t.Errorf("b: want %d, have %d", want, have)
	}
}

func testStringSlice(t *testing.T, want, have []string) {
	sort.Strings(want)
	sort.Strings(have)
	if !reflect.DeepEqual(want, have) {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf("%s:%d: want %v, have %v", filepath.Base(file), line, want, have)
	}
}
