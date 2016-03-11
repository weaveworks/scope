package render

import (
	"github.com/weaveworks/scope/report"
)

// CustomRenderer allow for mapping functions that received the entire topology
// in one call - useful for functions that need to consider the entire graph.
// We should minimise the use of this renderer type, as it is very inflexible.
type CustomRenderer struct {
	RenderFunc func(RenderableNodes) RenderableNodes
	Renderer
}

// Render implements Renderer
func (c CustomRenderer) Render(rpt report.Report) RenderableNodes {
	return c.RenderFunc(c.Renderer.Render(rpt))
}

// ColorConnected colors nodes with the IsConnected key if
// they have edges to or from them.  Edges to/from yourself
// are not counted here (see #656).
func ColorConnected(r Renderer) Renderer {
	return CustomRenderer{
		Renderer: r,
		RenderFunc: func(input RenderableNodes) RenderableNodes {
			connected := map[string]struct{}{}
			void := struct{}{}

			for id, node := range input {
				if len(node.Adjacency) == 0 {
					continue
				}

				for _, adj := range node.Adjacency {
					if adj != id {
						connected[id] = void
						connected[adj] = void
					}
				}
			}

			output := input.Copy()
			for id := range connected {
				output[id] = output[id].WithNode(report.MakeNodeWith(map[string]string{
					IsConnected: "true",
				}))
			}
			return output
		},
	}
}

// Filter removes nodes from a view based on a predicate.
type Filter struct {
	Renderer
	FilterFunc func(RenderableNode) bool
}

// MakeFilter makes a new Filter.
func MakeFilter(f func(RenderableNode) bool, r Renderer) Renderer {
	return Memoise(&Filter{r, f})
}

// Render implements Renderer
func (f *Filter) Render(rpt report.Report) RenderableNodes {
	nodes, _ := f.render(rpt)
	return nodes
}

func (f *Filter) render(rpt report.Report) (RenderableNodes, int) {
	output := RenderableNodes{}
	inDegrees := map[string]int{}
	filtered := 0
	for id, node := range f.Renderer.Render(rpt) {
		if f.FilterFunc(node) {
			output[id] = node
			inDegrees[id] = 0
		} else {
			filtered++
		}
	}

	// Deleted nodes also need to be cut as destinations in adjacency lists.
	for id, node := range output {
		newAdjacency := report.MakeIDList()
		for _, dstID := range node.Adjacency {
			if _, ok := output[dstID]; ok {
				newAdjacency = newAdjacency.Add(dstID)
				inDegrees[dstID]++
			}
		}
		node.Adjacency = newAdjacency
		output[id] = node
	}

	// Remove unconnected pseudo nodes, see #483.
	for id, inDegree := range inDegrees {
		if inDegree > 0 {
			continue
		}
		node := output[id]
		if !node.Pseudo || len(node.Adjacency) > 0 {
			continue
		}
		delete(output, id)
		filtered++
	}
	return output, filtered
}

// Stats implements Renderer
func (f Filter) Stats(rpt report.Report) Stats {
	_, filtered := f.render(rpt)
	var upstream = f.Renderer.Stats(rpt)
	upstream.FilteredNodes += filtered
	return upstream
}

// IsConnected is the key added to Node.Metadata by ColorConnected
// to indicate a node has an edge pointing to it or from it
const IsConnected = "is_connected"

// Complement takes a FilterFunc f and returns a FilterFunc that has the same
// effects, if any, and returns the opposite truth value.
func Complement(f func(RenderableNode) bool) func(RenderableNode) bool {
	return func(node RenderableNode) bool { return !f(node) }
}

// FilterPseudo produces a renderer that removes pseudo nodes from the given
// renderer
func FilterPseudo(r Renderer) Renderer {
	return MakeFilter(
		func(node RenderableNode) bool {
			return !node.Pseudo
		},
		r,
	)
}

// FilterUnconnected produces a renderer that filters unconnected nodes
// from the given renderer
func FilterUnconnected(r Renderer) Renderer {
	return MakeFilter(
		func(node RenderableNode) bool {
			_, ok := node.Latest.Lookup(IsConnected)
			return ok
		},
		ColorConnected(r),
	)
}

// FilterNoop does nothing.
func FilterNoop(in Renderer) Renderer {
	return in
}

// FilterStopped filters out stopped containers.
func FilterStopped(r Renderer) Renderer {
	return MakeFilter(RenderableNode.IsStopped, r)
}

// FilterRunning filters out running containers.
func FilterRunning(r Renderer) Renderer {
	return MakeFilter(Complement(RenderableNode.IsStopped), r)
}

// FilterSystem is a Renderer which filters out system nodes.
func FilterSystem(r Renderer) Renderer {
	return MakeFilter(RenderableNode.IsSystem, r)
}

// FilterApplication is a Renderer which filters out system nodes.
func FilterApplication(r Renderer) Renderer {
	return MakeFilter(Complement(RenderableNode.IsSystem), r)
}

var systemContainerNames = map[string]struct{}{
	"weavescope": {},
	"weavedns":   {},
	"weave":      {},
	"weaveproxy": {},
	"weaveexec":  {},
	"ecs-agent":  {},
}

var systemImagePrefixes = map[string]struct{}{
	"weaveworks/scope":                    {},
	"weaveworks/weavedns":                 {},
	"weaveworks/weave":                    {},
	"weaveworks/weaveproxy":               {},
	"weaveworks/weaveexec":                {},
	"amazon/amazon-ecs-agent":             {},
	"beta.gcr.io/google_containers/pause": {},
	"gcr.io/google_containers/pause":      {},
}
