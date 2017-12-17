package render

import (
	"github.com/weaveworks/scope/report"
)

// MapFunc is anything which can take an arbitrary Node and
// return a set of other Nodes.
//
// If the output is empty, the node shall be omitted from the rendered topology.
type MapFunc func(report.Node) report.Nodes

// Renderer is something that can render a report to a set of Nodes.
type Renderer interface {
	Render(report.Report) Nodes
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
func Render(rpt report.Report, renderer Renderer, transformer Transformer) Nodes {
	return transformer.Transform(renderer.Render(rpt))
}

// Reduce renderer is a Renderer which merges together the output of several
// other renderers.
type Reduce []Renderer

// MakeReduce is the only sane way to produce a Reduce Renderer.
func MakeReduce(renderers ...Renderer) Renderer {
	return Reduce(renderers)
}

// Render produces a set of Nodes given a Report.
func (r Reduce) Render(rpt report.Report) Nodes {
	l := len(r)
	switch l {
	case 0:
		return Nodes{}
	}
	c := make(chan Nodes, l)
	for _, renderer := range r {
		renderer := renderer // Pike!!
		go func() {
			c <- renderer.Render(rpt)
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
func (m Map) Render(rpt report.Report) Nodes {
	var (
		input       = m.Renderer.Render(rpt)
		output      = report.Nodes{}
		mapped      = map[string]report.IDList{} // input node ID -> output node IDs
		adjacencies = map[string]report.IDList{} // output node ID -> input node Adjacencies
	)

	// Rewrite all the nodes according to the map function
	for _, inRenderable := range input.Nodes {
		for _, outRenderable := range m.MapFunc(inRenderable) {
			if existing, ok := output[outRenderable.ID]; ok {
				outRenderable = outRenderable.Merge(existing)
			}

			output[outRenderable.ID] = outRenderable
			mapped[inRenderable.ID] = mapped[inRenderable.ID].Add(outRenderable.ID)
			adjacencies[outRenderable.ID] = adjacencies[outRenderable.ID].Merge(inRenderable.Adjacency)
		}
	}

	// Rewrite Adjacency for new node IDs.
	for outNodeID, inAdjacency := range adjacencies {
		outAdjacency := report.MakeIDList()
		for _, inAdjacent := range inAdjacency {
			outAdjacency = outAdjacency.Merge(mapped[inAdjacent])
		}
		outNode := output[outNodeID]
		outNode.Adjacency = outAdjacency
		output[outNodeID] = outNode
	}

	return Nodes{Nodes: output}
}

func propagateLatest(key string, from, to report.Node) report.Node {
	if value, timestamp, ok := from.Latest.LookupEntry(key); ok {
		to.Latest = to.Latest.Set(key, timestamp, value)
	}
	return to
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

func (cr conditionalRenderer) Render(rpt report.Report) Nodes {
	if cr.Condition(rpt) {
		return cr.Renderer.Render(rpt)
	}
	return Nodes{}
}

// joinResults is used by Renderers that join sets of nodes
type joinResults struct {
	nodes  report.Nodes
	mapped map[string]string // input node ID -> output node ID
}

func newJoinResults(inputNodes report.Nodes) joinResults {
	nodes := make(report.Nodes, len(inputNodes))
	for id, n := range inputNodes {
		nodes[id] = n
	}
	return joinResults{nodes: nodes, mapped: map[string]string{}}
}

// Add m as a child of the node at id, creating a new result node if
// not already there, and updating the mapping from old ID to new ID.
func (ret *joinResults) addChild(m report.Node, id string, create func(string) report.Node) {
	result, exists := ret.nodes[id]
	if !exists {
		result = create(id)
	}
	result.Children = result.Children.Add(m)
	if m.Topology != report.Endpoint { // optimisation: we never look at endpoint counts
		result.Counters = result.Counters.Add(m.Topology, 1)
	}
	ret.nodes[id] = result
	ret.mapped[m.ID] = id
}

// Like addChild, but also add m's children.
func (ret *joinResults) addChildAndChildren(m report.Node, id string, create func(string) report.Node) {
	result, exists := ret.nodes[id]
	if !exists {
		result = create(id)
	}
	result.Children = result.Children.Add(m)
	result.Children = result.Children.Merge(m.Children)
	if m.Topology != report.Endpoint { // optimisation: we never look at endpoint counts
		result.Counters = result.Counters.Add(m.Topology, 1)
	}
	ret.nodes[id] = result
	ret.mapped[m.ID] = id
}

// Add a copy of n straight into the results
func (ret *joinResults) passThrough(n report.Node) {
	n.Adjacency = nil // result() assumes all nodes start with blank lists
	ret.nodes[n.ID] = n
	ret.mapped[n.ID] = n.ID
}

// Rewrite Adjacency of nodes in ret mapped from original nodes in
// input, and return the result.
func (ret *joinResults) result(input Nodes) Nodes {
	for _, n := range input.Nodes {
		outID, ok := ret.mapped[n.ID]
		if !ok {
			continue
		}
		out := ret.nodes[outID]
		// for each adjacency in the original node, find out what it maps to (if any),
		// and add that to the new node
		for _, a := range n.Adjacency {
			if mappedDest, found := ret.mapped[a]; found {
				out.Adjacency = out.Adjacency.Add(mappedDest)
			}
		}
		ret.nodes[outID] = out
	}
	return Nodes{Nodes: ret.nodes}
}

// ResetCache blows away the rendered node cache, and known service
// cache.
func ResetCache() {
	renderCache.Purge()
	purgeKnownServiceCache()
}
