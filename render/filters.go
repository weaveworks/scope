package render

import (
	"log"
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
func (c CustomRenderer) Render(rpt report.Report, dct Decorator) report.Nodes {
	return c.RenderFunc(c.Renderer.Render(rpt, dct))
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

// FilterFunc is the function type used by Filters
type FilterFunc func(report.Node) bool

// ComposeFilterFuncs composes filterfuncs into a single FilterFunc checking all.
func ComposeFilterFuncs(fs ...FilterFunc) FilterFunc {
	return func(n report.Node) bool {
		for _, f := range fs {
			if !f(n) {
				return false
			}
		}
		return true
	}
}

// Filter removes nodes from a view based on a predicate.
type Filter struct {
	Renderer
	FilterFunc FilterFunc
}

// MakeFilter makes a new Filter (that ignores pseudo nodes).
func MakeFilter(f FilterFunc, r Renderer) Renderer {
	return Memoise(&Filter{
		Renderer: r,
		FilterFunc: func(n report.Node) bool {
			return n.Topology == Pseudo || f(n)
		},
	})
}

// MakeFilterPseudo makes a new Filter that will not ignore pseudo nodes.
func MakeFilterPseudo(f FilterFunc, r Renderer) Renderer {
	return Memoise(&Filter{
		Renderer:   r,
		FilterFunc: f,
	})
}

// MakeFilterDecorator makes a decorator that filters out non-pseudo nodes
// which match the predicate.
func MakeFilterDecorator(f FilterFunc) Decorator {
	return func(renderer Renderer) Renderer {
		return MakeFilter(f, renderer)
	}
}

// MakeFilterPseudoDecorator makes a decorator that filters out all nodes
// (including pseudo nodes) which match the predicate.
func MakeFilterPseudoDecorator(f FilterFunc) Decorator {
	return func(renderer Renderer) Renderer {
		return MakeFilterPseudo(f, renderer)
	}
}

// Render implements Renderer
func (f *Filter) Render(rpt report.Report, dct Decorator) report.Nodes {
	nodes, _ := f.render(rpt, dct)
	return nodes
}

func (f *Filter) render(rpt report.Report, dct Decorator) (report.Nodes, int) {
	output := report.Nodes{}
	inDegrees := map[string]int{}
	filtered := 0
	for id, node := range f.Renderer.Render(rpt, dct) {
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

// Stats implements Renderer. General logic is to take the first (i.e.
// highest-level) stats we find, so upstream stats are ignored. This means that
// if we want to count the stats from multiple filters we need to compose their
// FilterFuncs, into a single Filter.
func (f Filter) Stats(rpt report.Report, dct Decorator) Stats {
	_, filtered := f.render(rpt, dct)
	return Stats{FilteredNodes: filtered}
}

// IsConnected is the key added to Node.Metadata by ColorConnected
// to indicate a node has an edge pointing to it or from it
const IsConnected = "is_connected"

// Complement takes a FilterFunc f and returns a FilterFunc that has the same
// effects, if any, and returns the opposite truth value.
func Complement(f FilterFunc) FilterFunc {
	return func(node report.Node) bool { return !f(node) }
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

// Noop allows all nodes through
func Noop(_ report.Node) bool { return true }

// IsRunning checks if the node is a running docker container
func IsRunning(n report.Node) bool {
	state, ok := n.Latest.Lookup(docker.ContainerState)
	return !ok || (state == docker.StateRunning || state == docker.StateRestarting || state == docker.StatePaused)
}

// IsStopped checks if the node is *not* a running docker container
var IsStopped = Complement(IsRunning)

// FilterNonProcspied removes endpoints which were not found in procspy.
func FilterNonProcspied(r Renderer) Renderer {
	return MakeFilter(
		func(node report.Node) bool {
			_, ok := node.Latest.Lookup(endpoint.Procspied)
			return ok
		},
		r,
	)
}

// IsApplication checks if the node is an "application" node
func IsApplication(n report.Node) bool {
	containerName, _ := n.Latest.Lookup(docker.ContainerName)
	if _, ok := systemContainerNames[containerName]; ok {
		return false
	}
	imageName, _ := n.Latest.Lookup(docker.ImageName)
	imagePrefix := strings.SplitN(imageName, ":", 2)[0] // :(
	if _, ok := systemImagePrefixes[imagePrefix]; ok || kubernetes.IsPauseImageName(imagePrefix) {
		return false
	}
	roleLabel, _ := n.Latest.Lookup(docker.LabelPrefix + "works.weave.role")
	if roleLabel == "system" {
		return false
	}
	roleLabel, _ = n.Latest.Lookup(docker.ImageLabelPrefix + "works.weave.role")
	if roleLabel == "system" {
		return false
	}
	namespace, _ := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.namespace")
	if namespace == "kube-system" {
		return false
	}
	podName, _ := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.name")
	if strings.HasPrefix(podName, "kube-system/") {
		return false
	}
	return true
}

// IsSystem checks if the node is a "system" node
var IsSystem = Complement(IsApplication)

// IsDesired checks if the node has the desired label
func IsDesired(label string) FilterFunc {
	return func(n report.Node) bool {
		desiredKeyValue := strings.Split(label, "=")
		value, _ := n.Latest.Lookup(docker.LabelPrefix + desiredKeyValue[0])

		if len(desiredKeyValue) == 2 {
			if value == desiredKeyValue[1] {
				log.Println(value, "=", desiredKeyValue[1])
				return true
			}
		} else {
			log.Printf("label isn't in the correct key=value format")
		}
		return false
	}
}

// IsNotPseudo returns true if the node is not a pseudo node
// or internet/service nodes.
func IsNotPseudo(n report.Node) bool {
	return n.Topology != Pseudo || strings.HasSuffix(n.ID, TheInternetID) || strings.HasPrefix(n.ID, ServiceNodeIDPrefix)
}

// IsNamespace checks if the node is a pod/service in the specified namespace
func IsNamespace(namespace string) FilterFunc {
	return func(n report.Node) bool {
		gotNamespace, _ := n.Latest.Lookup(kubernetes.Namespace)
		return namespace == gotNamespace
	}
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
	"swarm":                          {},
	"weaveworks/scope":               {},
	"weaveworks/weavedns":            {},
	"weaveworks/weave":               {},
	"weaveworks/weaveproxy":          {},
	"weaveworks/weaveexec":           {},
	"amazon/amazon-ecs-agent":        {},
	"openshift/origin-pod":           {},
	"docker.io/openshift/origin-pod": {},
}
