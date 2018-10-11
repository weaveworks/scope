package render

import (
	"context"
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

// CustomRenderer allow for mapping functions that received the entire topology
// in one call - useful for functions that need to consider the entire graph.
// We should minimise the use of this renderer type, as it is very inflexible.
type CustomRenderer struct {
	RenderFunc func(Nodes) Nodes
	Renderer
}

// Render implements Renderer
func (c CustomRenderer) Render(ctx context.Context, rpt report.Report) Nodes {
	return c.RenderFunc(c.Renderer.Render(ctx, rpt))
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

// Complement takes a FilterFunc f and returns a FilterFunc that has the same
// effects, if any, and returns the opposite truth value.
func Complement(f FilterFunc) FilterFunc {
	return func(node report.Node) bool { return !f(node) }
}

// Transform applies the filter to all nodes
func (f FilterFunc) Transform(nodes Nodes) Nodes {
	output := report.Nodes{}
	filtered := nodes.Filtered
	for id, node := range nodes.Nodes {
		if f(node) {
			output[id] = node
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
			}
		}
		node.Adjacency = newAdjacency
		output[id] = node
	}

	return Nodes{Nodes: output, Filtered: filtered}
}

// Filter removes nodes from a view based on a predicate.
type Filter struct {
	Renderer
	FilterFunc FilterFunc
}

// MakeFilter makes a new Filter (that ignores pseudo nodes).
func MakeFilter(f FilterFunc, r Renderer) Renderer {
	return Filter{
		Renderer: r,
		FilterFunc: func(n report.Node) bool {
			return n.Topology == Pseudo || f(n)
		},
	}
}

// MakeFilterPseudo makes a new Filter that will not ignore pseudo nodes.
func MakeFilterPseudo(f FilterFunc, r Renderer) Renderer {
	return Filter{
		Renderer:   r,
		FilterFunc: f,
	}
}

// Render implements Renderer
func (f Filter) Render(ctx context.Context, rpt report.Report) Nodes {
	return f.FilterFunc.Transform(f.Renderer.Render(ctx, rpt))
}

// IsConnectedMark is the key added to Node.Metadata by
// ColorConnected to indicate a node has an edge pointing to it or
// from it
const IsConnectedMark = "is_connected"

// IsConnected checks whether the node has been marked with the
// IsConnectedMark.
func IsConnected(node report.Node) bool {
	_, ok := node.Latest.Lookup(IsConnectedMark)
	return ok
}

// IsPodComponent check whether given node is everything but PV, PVC, SC
func IsPodComponent(node report.Node) bool {
	var ok bool
	ok = true
	if node.Topology == "persistent_volume" || node.Topology == "persistent_volume_claim" || node.Topology == "storage_class" {
		ok = false
	}
	return ok
}

// connected returns the node ids of nodes which have edges to/from
// them, excluding edges to/from themselves.
func connected(nodes report.Nodes) map[string]struct{} {
	res := map[string]struct{}{}
	void := struct{}{}
	for id, node := range nodes {
		for _, adj := range node.Adjacency {
			if adj != id {
				res[id] = void
				res[adj] = void
			}
		}
	}
	return res
}

// filterInternetAdjacencies filters out edges between the incoming
// and outgoing internet node. These are typically artifacts of
// imperfect connection tracking, e.g. when VIPs and NAT traversal are
// in use.
func filterInternetAdjacencies(nodes report.Nodes) {
	incomingInternet, ok := nodes[IncomingInternetID]
	if !ok {
		return
	}
	newAdjacency := report.MakeIDList()
	for _, dstID := range incomingInternet.Adjacency {
		if dstID != OutgoingInternetID {
			newAdjacency = newAdjacency.Add(dstID)
		}
	}
	incomingInternet.Adjacency = newAdjacency
	nodes[IncomingInternetID] = incomingInternet
}

// ColorConnected colors nodes with the IsConnectedMark key if they
// have edges to or from them.  Edges to/from yourself are not counted
// here (see #656).
func ColorConnected(r Renderer) Renderer {
	return CustomRenderer{
		Renderer: r,
		RenderFunc: func(input Nodes) Nodes {
			output := input.Copy()
			for id := range connected(input.Nodes) {
				output[id] = output[id].WithLatest(IsConnectedMark, mtime.Now(), "true")
			}
			return Nodes{Nodes: output, Filtered: input.Filtered}
		},
	}
}

type filterUnconnected struct {
	onlyPseudo bool
}

// Transform implements Transformer
func (f filterUnconnected) Transform(input Nodes) Nodes {
	output := input.Nodes.Copy()
	filterInternetAdjacencies(output)
	connected := connected(output)
	filtered := input.Filtered
	for id, node := range output {
		if _, ok := connected[id]; !ok && (!f.onlyPseudo || IsPseudoTopology(node)) {
			delete(output, id)
			filtered++
		}
	}
	return Nodes{Nodes: output, Filtered: filtered}
}

// FilterUnconnected is a transformer that filters unconnected nodes
var FilterUnconnected = filterUnconnected{onlyPseudo: false}

// FilterUnconnectedPseudo is a transformer that filters unconnected
// pseudo nodes
var FilterUnconnectedPseudo = filterUnconnected{onlyPseudo: true}

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
	return n.Topology != Pseudo || IsInternetNode(n) || strings.HasPrefix(n.ID, ServiceNodeIDPrefix)
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
