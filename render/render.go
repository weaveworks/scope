package render

import (
	"github.com/weaveworks/scope/report"
)

// MapFunc is anything which can take an arbitrary Node and
// return a set of other Nodes.
//
// If the output is empty, the node shall be omitted from the rendered topology.
type MapFunc func(report.Node, report.Networks) report.Nodes

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

// Render renders the report and then applies the filter
func Render(rpt report.Report, renderer Renderer, filter FilterFunc) Nodes {
	nodes := renderer.Render(rpt)
	if filter != nil {
		nodes = filter.Apply(nodes)
	}
	return nodes
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
		input         = m.Renderer.Render(rpt)
		output        = report.Nodes{}
		mapped        = map[string]report.IDList{} // input node ID -> output node IDs
		adjacencies   = map[string]report.IDList{} // output node ID -> input node Adjacencies
		localNetworks = LocalNetworks(rpt)
	)

	// Rewrite all the nodes according to the map function
	for _, inRenderable := range input.Nodes {
		for _, outRenderable := range m.MapFunc(inRenderable, localNetworks) {
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

func newJoinResults() joinResults {
	return joinResults{nodes: make(report.Nodes), mapped: map[string]string{}}
}

// Add Node M under id, creating a new result node if not already there
// and updating the mapping from old ID to new ID
// Note we do not update any counters for child topologies here, because addToResults
// is only ever called when m is an endpoint and we never look at endpoint counts
func (ret *joinResults) addToResults(m report.Node, id string, create func(string) report.Node) {
	result, exists := ret.nodes[id]
	if !exists {
		result = create(id)
	}
	result.Children = result.Children.Add(m)
	result.Children = result.Children.Merge(m.Children)
	ret.nodes[id] = result
	ret.mapped[m.ID] = id
}

// Rewrite Adjacency for new nodes in ret for original nodes in input
func (ret *joinResults) fixupAdjacencies(input Nodes) {
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
}

func (ret *joinResults) copyUnmatched(input Nodes) {
	for _, n := range input.Nodes {
		if _, found := ret.nodes[n.ID]; !found {
			ret.nodes[n.ID] = n
		}
	}
}

func (ret *joinResults) result() Nodes {
	return Nodes{Nodes: ret.nodes}
}
