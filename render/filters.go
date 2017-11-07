package render

import (
	"strings"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

const (
	k8sNamespaceLabel   = "io.kubernetes.pod.namespace"
	swarmNamespaceLabel = "com.docker.stack.namespace"
)

// PreciousNodeRenderer ensures a node is never filtered out by decorators
type PreciousNodeRenderer struct {
	PreciousNodeID string
	Renderer
}

// Render implements Renderer
func (p PreciousNodeRenderer) Render(rpt report.Report, dct Decorator) Nodes {
	undecoratedNodes := p.Renderer.Render(rpt, nil)
	preciousNode, foundBeforeDecoration := undecoratedNodes.Nodes[p.PreciousNodeID]
	finalNodes := applyDecorator{ConstantRenderer{undecoratedNodes}}.Render(rpt, dct)
	if _, ok := finalNodes.Nodes[p.PreciousNodeID]; !ok && foundBeforeDecoration {
		finalNodes.Nodes[p.PreciousNodeID] = preciousNode
		finalNodes.Filtered--
	}
	return finalNodes
}

// Stats implements Renderer
func (p PreciousNodeRenderer) Stats(rpt report.Report, dct Decorator) Stats {
	// default to the underlying renderer
	// TODO: we don't take into account the precious node, so we may be off by one
	return p.Renderer.Stats(rpt, dct)
}

// CustomRenderer allow for mapping functions that received the entire topology
// in one call - useful for functions that need to consider the entire graph.
// We should minimise the use of this renderer type, as it is very inflexible.
type CustomRenderer struct {
	RenderFunc func(Nodes) Nodes
	Renderer
}

// Render implements Renderer
func (c CustomRenderer) Render(rpt report.Report, dct Decorator) Nodes {
	return c.RenderFunc(c.Renderer.Render(rpt, dct))
}

// ColorConnected colors nodes with the IsConnected key if
// they have edges to or from them.  Edges to/from yourself
// are not counted here (see #656).
func ColorConnected(r Renderer) Renderer {
	return CustomRenderer{
		Renderer: r,
		RenderFunc: func(input Nodes) Nodes {
			connected := map[string]struct{}{}
			void := struct{}{}

			for id, node := range input.Nodes {
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
			return Nodes{Nodes: output, Filtered: input.Filtered}
		},
	}
}

// FilterFunc is the function type used by Filters
type FilterFunc func(report.Node) bool

// AnyFilterFunc checks if any of the filterfuncs matches.
func AnyFilterFunc(fs ...FilterFunc) FilterFunc {
	return func(n report.Node) bool {
		for _, f := range fs {
			if f(n) {
				return true
			}
		}
		return false
	}
}

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
	return &Filter{
		Renderer: r,
		FilterFunc: func(n report.Node) bool {
			return n.Topology == Pseudo || f(n)
		},
	}
}

// MakeFilterPseudo makes a new Filter that will not ignore pseudo nodes.
func MakeFilterPseudo(f FilterFunc, r Renderer) Renderer {
	return &Filter{
		Renderer:   r,
		FilterFunc: f,
	}
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
func (f *Filter) Render(rpt report.Report, dct Decorator) Nodes {
	return f.render(rpt, dct)
}

func (f *Filter) render(rpt report.Report, dct Decorator) Nodes {
	output := report.Nodes{}
	inDegrees := map[string]int{}
	filtered := 0
	for id, node := range f.Renderer.Render(rpt, dct).Nodes {
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
	return Nodes{Nodes: output, Filtered: filtered}
}

// Stats implements Renderer. General logic is to take the first (i.e.
// highest-level) stats we find, so upstream stats are ignored. This means that
// if we want to count the stats from multiple filters we need to compose their
// FilterFuncs, into a single Filter.
func (f Filter) Stats(rpt report.Report, dct Decorator) Stats {
	nodes := f.render(rpt, dct)
	return Stats{FilteredNodes: nodes.Filtered}
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
	return MakeFilterPseudo(
		func(node report.Node) bool {
			_, ok := node.Latest.Lookup(IsConnected)
			return ok
		},
		ColorConnected(r),
	)
}

// FilterUnconnectedPseudo produces a renderer that filters
// unconnected pseudo nodes from the given renderer
func FilterUnconnectedPseudo(r Renderer) Renderer {
	return MakeFilterPseudo(
		func(node report.Node) bool {
			if !IsPseudoTopology(node) {
				return true
			}
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
	namespace, _ := n.Latest.Lookup(docker.LabelPrefix + k8sNamespaceLabel)
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

// HasLabel checks if the node has the desired docker label
func HasLabel(labelKey string, labelValue string) FilterFunc {
	return func(n report.Node) bool {
		value, _ := n.Latest.Lookup(docker.LabelPrefix + labelKey)
		if value == labelValue {
			return true
		}
		return false
	}
}

// DoesNotHaveLabel checks if the node does NOT have the specified docker label
func DoesNotHaveLabel(labelKey string, labelValue string) FilterFunc {
	return Complement(HasLabel(labelKey, labelValue))
}

// IsNotPseudo returns true if the node is not a pseudo node
// or internet/service nodes.
func IsNotPseudo(n report.Node) bool {
	return n.Topology != Pseudo || strings.HasSuffix(n.ID, TheInternetID) || strings.HasPrefix(n.ID, ServiceNodeIDPrefix)
}

// IsNamespace checks if the node is a pod/service in the specified namespace
func IsNamespace(namespace string) FilterFunc {
	return func(n report.Node) bool {
		tryKeys := []string{kubernetes.Namespace, docker.LabelPrefix + k8sNamespaceLabel, docker.StackNamespace, docker.LabelPrefix + swarmNamespaceLabel}
		gotNamespace := ""
		for _, key := range tryKeys {
			if value, ok := n.Latest.Lookup(key); ok {
				gotNamespace = value
				break
			}
		}
		// Special case for docker
		if namespace == docker.DefaultNamespace && gotNamespace == "" {
			return true
		}
		return namespace == gotNamespace
	}
}

// IsTopology checks if the node is from a particular report topology
func IsTopology(topology string) FilterFunc {
	return func(n report.Node) bool {
		return n.Topology == topology
	}
}

// IsPseudoTopology returns true if the node is in a pseudo topology,
// mimicing the check performed by MakeFilter() instead of the more complex check in IsNotPseudo()
var IsPseudoTopology = IsTopology(Pseudo)

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
