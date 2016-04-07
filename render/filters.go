package render

import (
	"strings"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

// CustomRenderer allow for mapping functions that received the entire topology
// in one call - useful for functions that need to consider the entire graph.
// We should minimise the use of this renderer type, as it is very inflexible.
type CustomRenderer struct {
	RenderFunc func(report.Nodes) report.Nodes
	Renderer
}

// Render implements Renderer
func (c CustomRenderer) Render(rpt report.Report) report.Nodes {
	return c.RenderFunc(c.Renderer.Render(rpt))
}

// ColorConnected colors nodes with the IsConnected key if
// they have edges to or from them.  Edges to/from yourself
// are not counted here (see #656).
func ColorConnected(r Renderer) Renderer {
	return CustomRenderer{
		Renderer: r,
		RenderFunc: func(input report.Nodes) report.Nodes {
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
				output[id] = output[id].WithLatest(IsConnected, mtime.Now(), "true")
			}
			return output
		},
	}
}

// Filter removes nodes from a view based on a predicate.
type Filter struct {
	Renderer
	FilterFunc func(report.Node) bool
	Silent     bool // true means we don't report stats for how many are filtered
}

// MakeFilter makes a new Filter.
func MakeFilter(f func(report.Node) bool, r Renderer) Renderer {
	return Memoise(&Filter{
		Renderer:   r,
		FilterFunc: f,
	})
}

// MakeSilentFilter makes a new Filter which does not report how many nodes it filters in Stats.
func MakeSilentFilter(f func(report.Node) bool, r Renderer) Renderer {
	return Memoise(&Filter{
		Renderer:   r,
		FilterFunc: f,
		Silent:     true,
	})
}

// Render implements Renderer
func (f *Filter) Render(rpt report.Report) report.Nodes {
	nodes, _ := f.render(rpt)
	return nodes
}

func (f *Filter) render(rpt report.Report) (report.Nodes, int) {
	output := report.Nodes{}
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
		if node.Topology != Pseudo || len(node.Adjacency) > 0 {
			continue
		}
		delete(output, id)
		filtered++
	}
	return output, filtered
}

// Stats implements Renderer
func (f Filter) Stats(rpt report.Report) Stats {
	var upstream = f.Renderer.Stats(rpt)
	if !f.Silent {
		_, filtered := f.render(rpt)
		upstream.FilteredNodes += filtered
	}
	return upstream
}

// IsConnected is the key added to Node.Metadata by ColorConnected
// to indicate a node has an edge pointing to it or from it
const IsConnected = "is_connected"

// Complement takes a FilterFunc f and returns a FilterFunc that has the same
// effects, if any, and returns the opposite truth value.
func Complement(f func(report.Node) bool) func(report.Node) bool {
	return func(node report.Node) bool { return !f(node) }
}

// FilterPseudo produces a renderer that removes pseudo nodes from the given
// renderer
func FilterPseudo(r Renderer) Renderer {
	return MakeFilter(
		func(node report.Node) bool {
			return node.Topology != Pseudo
		},
		r,
	)
}

// FilterUnconnected produces a renderer that filters unconnected nodes
// from the given renderer
func FilterUnconnected(r Renderer) Renderer {
	return MakeFilter(
		func(node report.Node) bool {
			_, ok := node.Latest.Lookup(IsConnected)
			return ok
		},
		ColorConnected(r),
	)
}

// SilentFilterUnconnected produces a renderer that filters unconnected nodes
// from the given renderer; nodes filtered by this are not reported in stats.
func SilentFilterUnconnected(r Renderer) Renderer {
	return MakeSilentFilter(
		func(node report.Node) bool {
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
	return MakeFilter(IsRunning, r)
}

// IsRunning checks if the node is a running docker container
func IsRunning(n report.Node) bool {
	state, ok := n.Latest.Lookup(docker.ContainerState)
	return !ok || (state == docker.StateRunning || state == docker.StateRestarting || state == docker.StatePaused)
}

// FilterRunning filters out running containers.
func FilterRunning(r Renderer) Renderer {
	return MakeFilter(Complement(IsRunning), r)
}

// FilterNonProcspied removes endpoints which were not found in procspy.
func FilterNonProcspied(r Renderer) Renderer {
	return MakeSilentFilter(
		func(node report.Node) bool {
			_, ok := node.Latest.Lookup(endpoint.Procspied)
			return ok
		},
		r,
	)
}

// IsSystem checks if the node is a "system" node
func IsSystem(n report.Node) bool {
	containerName, _ := n.Latest.Lookup(docker.ContainerName)
	if _, ok := systemContainerNames[containerName]; ok {
		return false
	}
	imageName, _ := n.Latest.Lookup(docker.ImageName)
	imagePrefix := strings.SplitN(imageName, ":", 2)[0] // :(
	if _, ok := systemImagePrefixes[imagePrefix]; ok {
		return false
	}
	roleLabel, _ := n.Latest.Lookup(docker.LabelPrefix + "works.weave.role")
	if roleLabel == "system" {
		return false
	}
	namespace, _ := n.Latest.Lookup(kubernetes.Namespace)
	if namespace == "kube-system" {
		return false
	}
	podName, _ := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.name")
	if strings.HasPrefix(podName, "kube-system/") {
		return false
	}
	return true
}

// FilterSystem is a Renderer which filters out system nodes.
func FilterSystem(r Renderer) Renderer {
	return MakeFilter(IsSystem, r)
}

// FilterApplication is a Renderer which filters out system nodes.
func FilterApplication(r Renderer) Renderer {
	return MakeFilter(Complement(IsSystem), r)
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
