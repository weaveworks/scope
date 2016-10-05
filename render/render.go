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
	Render(report.Report, Decorator) report.Nodes
	Stats(report.Report, Decorator) Stats
}

// Stats is the type returned by Renderer.Stats
type Stats struct {
	FilteredNodes int
}

func (s Stats) merge(other Stats) Stats {
	return Stats{
		FilteredNodes: s.FilteredNodes + other.FilteredNodes,
	}
}

// Reduce renderer is a Renderer which merges together the output of several
// other renderers.
type Reduce []Renderer

// MakeReduce is the only sane way to produce a Reduce Renderer.
func MakeReduce(renderers ...Renderer) Renderer {
	r := Reduce(renderers)
	return Memoise(&r)
}

// Render produces a set of Nodes given a Report.
func (r *Reduce) Render(rpt report.Report, dct Decorator) report.Nodes {
	result := report.Nodes{}
	for _, renderer := range *r {
		result = result.Merge(renderer.Render(rpt, dct))
	}
	return result
}

// Stats implements Renderer
func (r *Reduce) Stats(rpt report.Report, dct Decorator) Stats {
	var result Stats
	for _, renderer := range *r {
		result = result.merge(renderer.Stats(rpt, dct))
	}
	return result
}

// Map is a Renderer which produces a set of Nodes from the set of
// Nodes produced by another Renderer.
type Map struct {
	MapFunc
	Renderer
}

// MakeMap makes a new Map
func MakeMap(f MapFunc, r Renderer) Renderer {
	return Memoise(&Map{f, r})
}

// Render transforms a set of Nodes produces by another Renderer.
// using a map function
func (m *Map) Render(rpt report.Report, dct Decorator) report.Nodes {
	var (
		input         = m.Renderer.Render(rpt, dct)
		output        = report.Nodes{}
		mapped        = map[string]report.IDList{} // input node ID -> output node IDs
		adjacencies   = map[string]report.IDList{} // output node ID -> input node Adjacencies
		localNetworks = LocalNetworks(rpt)
	)

	// Rewrite all the nodes according to the map function
	for _, inRenderable := range input {
		for _, outRenderable := range m.MapFunc(inRenderable, localNetworks) {
			existing, ok := output[outRenderable.ID]
			if ok {
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
			for _, outAdjacent := range mapped[inAdjacent] {
				outAdjacency = outAdjacency.Add(outAdjacent)
			}
		}
		outNode := output[outNodeID]
		outNode.Adjacency = outAdjacency
		output[outNodeID] = outNode
	}

	return output
}

// Stats implements Renderer
func (m *Map) Stats(_ report.Report, _ Decorator) Stats {
	// There doesn't seem to be an instance where we want stats to recurse
	// through Maps - for instance we don't want to see the number of filtered
	// processes in the container renderer.
	return Stats{}
}

// Decorator transforms one renderer to another. e.g. Filters.
type Decorator func(Renderer) Renderer

// ComposeDecorators composes decorators into one.
func ComposeDecorators(decorators ...Decorator) Decorator {
	return func(r Renderer) Renderer {
		for _, decorator := range decorators {
			r = decorator(r)
		}
		return r
	}
}

type applyDecorator struct {
	Renderer
}

func (ad applyDecorator) Render(rpt report.Report, dct Decorator) report.Nodes {
	if dct != nil {
		return dct(ad.Renderer).Render(rpt, nil)
	}
	return ad.Renderer.Render(rpt, nil)
}
func (ad applyDecorator) Stats(rpt report.Report, dct Decorator) Stats {
	if dct != nil {
		return dct(ad.Renderer).Stats(rpt, nil)
	}
	return Stats{}
}

// ApplyDecorators returns a renderer which will apply the given decorators
// to the child render.
func ApplyDecorators(renderer Renderer) Renderer {
	return applyDecorator{renderer}
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
	return Memoise(conditionalRenderer{c, r})
}

func (cr conditionalRenderer) Render(rpt report.Report, dct Decorator) report.Nodes {
	if cr.Condition(rpt) {
		return cr.Renderer.Render(rpt, dct)
	}
	return report.Nodes{}
}
func (cr conditionalRenderer) Stats(rpt report.Report, dct Decorator) Stats {
	if cr.Condition(rpt) {
		return cr.Renderer.Stats(rpt, dct)
	}
	return Stats{}
}

// ConstantRenderer renders a fixed set of nodes
type ConstantRenderer report.Nodes

// Render implements Renderer
func (c ConstantRenderer) Render(_ report.Report, _ Decorator) report.Nodes {
	return report.Nodes(c)
}

// Stats implements Renderer
func (c ConstantRenderer) Stats(_ report.Report, _ Decorator) Stats {
	return Stats{}
}
