package render

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/weaveworks/scope/report"
)

// MapFunc is anything which can take an arbitrary Node and
// return another Node.
//
// If the output ID is blank, the node shall be omitted from the rendered topology.
// (we chose not to return an extra bool because it adds clutter)
type MapFunc func(report.Node) report.Node

// Renderer is something that can render a report to a set of Nodes.
type Renderer interface {
	Render(context.Context, report.Report) Nodes
}

// Nodes is the result of Rendering
type Nodes struct {
	report.Nodes
	Filtered int
}

// Merge merges the results of Rendering
func (r Nodes) Merge(o Nodes) Nodes {
	return Nodes{
		Nodes:    r.Nodes.Merge(o.Nodes),
		Filtered: r.Filtered + o.Filtered,
	}
}

// Transformer is something that transforms one set of Nodes to
// another set of Nodes.
type Transformer interface {
	Transform(nodes Nodes) Nodes
}

// Transformers is a composition of Transformers
type Transformers []Transformer

// Transform implements Transformer
func (ts Transformers) Transform(nodes Nodes) Nodes {
	for _, t := range ts {
		nodes = t.Transform(nodes)
	}
	return nodes
}

// Render renders the report and then transforms it
func Render(ctx context.Context, rpt report.Report, renderer Renderer, transformer Transformer) Nodes {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Render:"+typeName(renderer))
	defer span.Finish()
	return transformer.Transform(renderer.Render(ctx, rpt))
}

// Reduce renderer is a Renderer which merges together the output of several
// other renderers.
type Reduce []Renderer

// MakeReduce is the only sane way to produce a Reduce Renderer.
func MakeReduce(renderers ...Renderer) Renderer {
	return Reduce(renderers)
}

// Render produces a set of Nodes given a Report.
func (r Reduce) Render(ctx context.Context, rpt report.Report) Nodes {
	if ctx.Err() != nil {
		return Nodes{}
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, "Reduce.Render")
	defer span.Finish()
	l := len(r)
	switch l {
	case 0:
		return Nodes{}
	}
	c := make(chan Nodes, l)
	for _, renderer := range r {
		renderer := renderer // Pike!!
		go func() {
			span, ctx := opentracing.StartSpanFromContext(ctx, typeName(renderer))
			c <- renderer.Render(ctx, rpt)
			span.Finish()
		}()
	}
	for ; l > 1; l-- {
		left, right := <-c, <-c
		go func() {
			c <- left.Merge(right)
		}()
	}
	return <-c
}

// Map is a Renderer which produces a set of Nodes from the set of
// Nodes produced by another Renderer.
type Map struct {
	MapFunc
	Renderer
}

// MakeMap makes a new Map
func MakeMap(f MapFunc, r Renderer) Renderer {
	return Map{f, r}
}

// Render transforms a set of Nodes produces by another Renderer.
// using a map function
func (m Map) Render(ctx context.Context, rpt report.Report) Nodes {
	if ctx.Err() != nil {
		return Nodes{}
	}
	span, ctx := opentracing.StartSpanFromContext(ctx, "Map.Render:"+functionName(m.MapFunc))
	defer span.Finish()
	var (
		input  = m.Renderer.Render(ctx, rpt)
		output = newJoinResults(nil)
	)

	// Rewrite all the nodes according to the map function
	for _, inRenderable := range input.Nodes {
		outRenderable := m.MapFunc(inRenderable)
		if outRenderable.ID != "" {
			output.add(inRenderable.ID, outRenderable)
		}
	}
	span.LogFields(otlog.Int("input.nodes", len(input.Nodes)),
		otlog.Int("ouput.nodes", len(output.nodes)))

	return output.result(input)
}

// Condition is a predecate over the entire report that can evaluate to true or false.
type Condition func(report.Report) bool

type conditionalRenderer struct {
	Condition
	Renderer
}

// ConditionalRenderer renders nothing if the condition is false, otherwise it defers
// to the wrapped Renderer.
func ConditionalRenderer(c Condition, r Renderer) Renderer {
	return conditionalRenderer{c, r}
}

func (cr conditionalRenderer) Render(ctx context.Context, rpt report.Report) Nodes {
	if cr.Condition(rpt) {
		return cr.Renderer.Render(ctx, rpt)
	}
	return Nodes{}
}

// joinResults is used by Renderers that join sets of nodes
type joinResults struct {
	nodes  report.Nodes
	mapped map[string]string   // input node ID -> output node ID - common case
	multi  map[string][]string // input node ID -> output node IDs - exceptional case
}

func newJoinResults(inputNodes report.Nodes) joinResults {
	nodes := make(report.Nodes, len(inputNodes))
	for id, n := range inputNodes {
		n.Adjacency = nil              // result() assumes all nodes start with no adjacencies
		n.Children = n.Children.Copy() // so we can do unsafe adds
		nodes[id] = n
	}
	return joinResults{nodes: nodes, mapped: map[string]string{}, multi: map[string][]string{}}
}

func (ret *joinResults) mapChild(from, to string) {
	if _, ok := ret.mapped[from]; !ok {
		ret.mapped[from] = to
	} else {
		ret.multi[from] = append(ret.multi[from], to)
	}
}

// Add m into the results as a top-level node, mapped from original ID
// Note it is not safe to mix calls to add() with addChild(), addChildAndChildren() or addUnmappedChild()
func (ret *joinResults) add(from string, m report.Node) {
	if existing, ok := ret.nodes[m.ID]; ok {
		m = m.Merge(existing)
	}
	ret.nodes[m.ID] = m
	ret.mapChild(from, m.ID)
}

// Add m as a child of the node at id, creating a new result node in
// the specified topology if not already there.
func (ret *joinResults) addUnmappedChild(m report.Node, id string, topology string) {
	result, exists := ret.nodes[id]
	if !exists {
		result = report.MakeNode(id).WithTopology(topology)
	}
	result.Children.UnsafeAdd(m)
	if m.Topology != report.Endpoint { // optimisation: we never look at endpoint counts
		result = result.AddCounter(m.Topology, 1)
	}
	ret.nodes[id] = result
}

// Add m as a child of the node at id, creating a new result node in
// the specified topology if not already there, and updating the
// mapping from old ID to new ID.
func (ret *joinResults) addChild(m report.Node, id string, topology string) {
	ret.addUnmappedChild(m, id, topology)
	ret.mapChild(m.ID, id)
}

// Like addChild, but also add m's children.
func (ret *joinResults) addChildAndChildren(m report.Node, id string, topology string) {
	ret.addUnmappedChild(m, id, topology)
	result := ret.nodes[id]
	result.Children.UnsafeMerge(m.Children)
	ret.nodes[id] = result
	ret.mapChild(m.ID, id)
}

// Add a copy of n straight into the results
func (ret *joinResults) passThrough(n report.Node) {
	n.Adjacency = nil // result() assumes all nodes start with no adjacencies
	ret.nodes[n.ID] = n
	n.Children = n.Children.Copy() // so we can do unsafe adds
	ret.mapChild(n.ID, n.ID)
}

// Rewrite Adjacency of nodes in ret mapped from original nodes in
// input, and return the result.
func (ret *joinResults) result(input Nodes) Nodes {
	for _, n := range input.Nodes {
		outID, ok := ret.mapped[n.ID]
		if !ok {
			continue
		}
		ret.rewriteAdjacency(outID, n.Adjacency)
		for _, outID := range ret.multi[n.ID] {
			ret.rewriteAdjacency(outID, n.Adjacency)
		}
	}
	return Nodes{Nodes: ret.nodes}
}

func (ret *joinResults) rewriteAdjacency(outID string, adjacency report.IDList) {
	out := ret.nodes[outID]
	// for each adjacency in the original node, find out what it maps
	// to (if any), and add that to the new node
	for _, a := range adjacency {
		if mappedDest, found := ret.mapped[a]; found {
			out.Adjacency = out.Adjacency.Add(mappedDest)
			out.Adjacency = out.Adjacency.Add(ret.multi[a]...)
		}
	}
	ret.nodes[outID] = out
}

// ResetCache blows away the rendered node cache, and known service
// cache.
func ResetCache() {
	renderCache.Purge()
	purgeKnownServiceCache()
}
