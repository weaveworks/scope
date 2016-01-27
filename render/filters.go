package render

import (
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

// CustomRenderer allow for mapping functions that recived the entire topology
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

			input.ForEach(func(node RenderableNode) {
				if len(node.Adjacency) == 0 {
					return
				}

				for _, adj := range node.Adjacency {
					if adj != node.ID {
						connected[node.ID] = void
						connected[adj] = void
					}
				}
			})

			for id := range connected {
				if existing, ok := input.Lookup(id); ok {
					input = input.Add(existing.WithNode(report.MakeNodeWith(map[string]string{
						IsConnected: "true",
					})))
				}
			}
			return input
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
	return &Filter{r, f}
}

// Render implements Renderer
func (f *Filter) Render(rpt report.Report) RenderableNodes {
	nodes, _ := f.render(rpt)
	return nodes
}

func (f *Filter) render(rpt report.Report) (RenderableNodes, int) {
	output := EmptyRenderableNodes
	inDegrees := map[string]int{}
	filtered := 0
	memoisedRender(f.Renderer, rpt).ForEach(func(node RenderableNode) {
		if f.FilterFunc(node) {
			output = output.Add(node)
			inDegrees[node.ID] = 0
		} else {
			filtered++
		}
	})

	// Deleted nodes also need to be cut as destinations in adjacency lists.
	output.ForEach(func(node RenderableNode) {
		newAdjacency := report.MakeIDList()
		for _, dstID := range node.Adjacency {
			if _, ok := output.Lookup(dstID); ok {
				newAdjacency = newAdjacency.Add(dstID)
				inDegrees[dstID]++
			}
		}
		node.Adjacency = newAdjacency
		output = output.Add(node)
	})

	// Remove unconnected pseudo nodes, see #483.
	for id, inDegree := range inDegrees {
		if inDegree > 0 {
			continue
		}
		node, _ := output.Lookup(id)
		if !node.Pseudo || len(node.Adjacency) > 0 {
			continue
		}
		output = output.Delete(id)
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
	return MakeFilter(
		func(node RenderableNode) bool {
			containerState, ok := node.Latest.Lookup(docker.ContainerState)
			return !ok || containerState != docker.StateStopped
		},
		r,
	)
}

// FilterSystem is a Renderer which filters out system nodes.
func FilterSystem(r Renderer) Renderer {
	return MakeFilter(
		func(node RenderableNode) bool {
			containerName, _ := node.Latest.Lookup(docker.ContainerName)
			if _, ok := systemContainerNames[containerName]; ok {
				return false
			}
			imageName, _ := node.Latest.Lookup(docker.ImageName)
			imagePrefix := strings.SplitN(imageName, ":", 2)[0] // :(
			if _, ok := systemImagePrefixes[imagePrefix]; ok {
				return false
			}
			roleLabel, _ := node.Latest.Lookup(docker.LabelPrefix + "works.weave.role")
			if roleLabel == "system" {
				return false
			}
			namespace, _ := node.Latest.Lookup(kubernetes.Namespace)
			if namespace == "kube-system" {
				return false
			}
			podName, _ := node.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.name")
			if strings.HasPrefix(podName, "kube-system/") {
				return false
			}
			return true
		},
		r,
	)
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
