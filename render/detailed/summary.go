package detailed

import (
	"fmt"
	"strings"

	"github.com/weaveworks/scope/probe/awsecs"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

// Shapes that are allowed
const (
	ImageNameNone = "<none>"

	// Keys we use to render container names
	AmazonECSContainerNameLabel  = "com.amazonaws.ecs.container-name"
	KubernetesContainerNameLabel = "io.kubernetes.container.name"
	MarathonAppIDEnv             = "MARATHON_APP_ID"
)

// NodeSummaryGroup is a topology-typed group of children for a Node.
type NodeSummaryGroup struct {
	ID         string        `json:"id"`
	Label      string        `json:"label"`
	Nodes      []NodeSummary `json:"nodes"`
	TopologyID string        `json:"topologyId"`
	Columns    []Column      `json:"columns"`
}

// Column provides special json serialization for column ids, so they include
// their label for the frontend.
type Column struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	DefaultSort bool   `json:"defaultSort"`
	Datatype    string `json:"dataType"`
}

// NodeSummary is summary information about a child for a Node.
type NodeSummary struct {
	ID         string               `json:"id"`
	Label      string               `json:"label"`
	LabelMinor string               `json:"labelMinor"`
	Rank       string               `json:"rank"`
	Shape      string               `json:"shape,omitempty"`
	Stack      bool                 `json:"stack,omitempty"`
	Linkable   bool                 `json:"linkable,omitempty"` // Whether this node can be linked-to
	Pseudo     bool                 `json:"pseudo,omitempty"`
	Metadata   []report.MetadataRow `json:"metadata,omitempty"`
	Parents    []Parent             `json:"parents,omitempty"`
	Metrics    []report.MetricRow   `json:"metrics,omitempty"`
	Tables     []report.Table       `json:"tables,omitempty"`
	Adjacency  report.IDList        `json:"adjacency,omitempty"`
}

var renderers = map[string]func(NodeSummary, report.Node) (NodeSummary, bool){
	render.Pseudo:         pseudoNodeSummary,
	report.Process:        processNodeSummary,
	report.Container:      containerNodeSummary,
	report.ContainerImage: containerImageNodeSummary,
	report.Pod:            podNodeSummary,
	report.Service:        podGroupNodeSummary,
	report.Deployment:     podGroupNodeSummary,
	report.DaemonSet:      podGroupNodeSummary,
	report.ECSTask:        ecsTaskNodeSummary,
	report.ECSService:     ecsServiceNodeSummary,
	report.SwarmService:   swarmServiceNodeSummary,
	report.Host:           hostNodeSummary,
	report.Overlay:        weaveNodeSummary,
	report.Endpoint:       nil, // Do not render
}

var templates = map[string]struct{ Label, LabelMinor string }{
	render.TheInternetID:      {render.InboundMajor, ""},
	render.IncomingInternetID: {render.InboundMajor, render.InboundMinor},
	render.OutgoingInternetID: {render.OutboundMajor, render.OutboundMinor},
}

// For each report.Topology, map to a 'primary' API topology. This can then be used in a variety of places.
var primaryAPITopology = map[string]string{
	report.Process:        "processes",
	report.Container:      "containers",
	report.ContainerImage: "containers-by-image",
	report.Pod:            "pods",
	report.Deployment:     "kube-controllers",
	report.DaemonSet:      "kube-controllers",
	report.Service:        "services",
	report.ECSTask:        "ecs-tasks",
	report.ECSService:     "ecs-services",
	report.SwarmService:   "swarm-services",
	report.Host:           "hosts",
}

// MakeNodeSummary summarizes a node, if possible.
func MakeNodeSummary(r report.Report, n report.Node) (NodeSummary, bool) {
	if renderer, ok := renderers[n.Topology]; ok {
		// Skip (and don't fall through to fallback) if renderer maps to nil
		if renderer != nil {
			return renderer(baseNodeSummary(r, n), n)
		}
	} else if _, ok := r.Topology(n.Topology); ok {
		summary := baseNodeSummary(r, n)
		summary.Label = n.ID // This is unlikely to look very good, but is a reasonable fallback
		return summary, true
	}
	if strings.HasPrefix(n.Topology, "group:") {
		return groupNodeSummary(baseNodeSummary(r, n), r, n)
	}
	return NodeSummary{}, false
}

// SummarizeMetrics returns a copy of the NodeSummary where the metrics are
// replaced with their summaries
func (n NodeSummary) SummarizeMetrics() NodeSummary {
	summarizedMetrics := make([]report.MetricRow, len(n.Metrics))
	for i, m := range n.Metrics {
		summarizedMetrics[i] = m.Summary()
	}
	n.Metrics = summarizedMetrics
	return n
}

func baseNodeSummary(r report.Report, n report.Node) NodeSummary {
	t, _ := r.Topology(n.Topology)
	return NodeSummary{
		ID:        n.ID,
		Shape:     t.GetShape(),
		Linkable:  true,
		Metadata:  NodeMetadata(r, n),
		Metrics:   NodeMetrics(r, n),
		Parents:   Parents(r, n),
		Tables:    NodeTables(r, n),
		Adjacency: n.Adjacency,
	}
}

func pseudoNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Pseudo = true
	base.Rank = n.ID

	// try rendering as an internet node
	if template, ok := templates[n.ID]; ok {
		base.Label = template.Label
		base.LabelMinor = template.LabelMinor
		base.Shape = report.Cloud
		return base, true
	}

	// try rendering as a known service node
	if strings.HasPrefix(n.ID, render.ServiceNodeIDPrefix) {
		base.Label = n.ID[len(render.ServiceNodeIDPrefix):]
		base.LabelMinor = ""
		base.Shape = report.Cloud
		return base, true
	}

	// try rendering it as an uncontained node
	if strings.HasPrefix(n.ID, render.UncontainedIDPrefix) {
		base.Label = render.UncontainedMajor
		base.LabelMinor = report.ExtractHostID(n)
		base.Shape = report.Square
		base.Stack = true
		return base, true
	}

	// try rendering it as an unmanaged node
	if strings.HasPrefix(n.ID, render.UnmanagedIDPrefix) {
		base.Label = render.UnmanagedMajor
		base.Shape = report.Square
		base.Stack = true
		base.LabelMinor = report.ExtractHostID(n)
		return base, true
	}

	// try rendering it as an endpoint
	if _, addr, _, ok := report.ParseEndpointNodeID(n.ID); ok {
		base.Label = addr
		base.Shape = report.Circle
		return base, true
	}

	return NodeSummary{}, false
}

func processNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(process.Name)
	base.Rank, _ = n.Latest.Lookup(process.Name)

	pid, ok := n.Latest.Lookup(process.PID)
	if !ok {
		return NodeSummary{}, false
	}
	if containerName, ok := n.Latest.Lookup(docker.ContainerName); ok {
		base.LabelMinor = fmt.Sprintf("%s (%s:%s)", report.ExtractHostID(n), containerName, pid)
	} else {
		base.LabelMinor = fmt.Sprintf("%s (%s)", report.ExtractHostID(n), pid)
	}

	_, isConnected := n.Latest.Lookup(render.IsConnected)
	base.Linkable = isConnected
	return base, true
}

func containerNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label = getRenderableContainerName(n)
	base.LabelMinor = report.ExtractHostID(n)

	if imageName, ok := n.Latest.Lookup(docker.ImageName); ok {
		base.Rank = docker.ImageNameWithoutVersion(imageName)
	}

	return base, true
}

func containerImageNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	imageName, ok := n.Latest.Lookup(docker.ImageName)
	if !ok {
		return NodeSummary{}, false
	}

	imageNameWithoutVersion := docker.ImageNameWithoutVersion(imageName)
	base.Label = imageNameWithoutVersion
	base.Rank = imageNameWithoutVersion
	base.Stack = true

	if base.Label == ImageNameNone {
		base.Label, _ = n.Latest.Lookup(docker.ImageID)
		if len(base.Label) > 12 {
			base.Label = base.Label[:12]
		}
	}

	base.LabelMinor = pluralize(n.Counters, report.Container, "container", "containers")

	return base, true
}

func addKubernetesLabelAndRank(base NodeSummary, n report.Node) NodeSummary {
	base.Label, _ = n.Latest.Lookup(kubernetes.Name)
	namespace, _ := n.Latest.Lookup(kubernetes.Namespace)
	base.Rank = namespace + "/" + base.Label
	return base
}

func podNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base = addKubernetesLabelAndRank(base, n)
	base.LabelMinor = pluralize(n.Counters, report.Container, "container", "containers")

	return base, true
}

var podGroupNodeTypeName = map[string]string{
	report.Deployment: "Deployment",
	report.DaemonSet:  "Daemon Set",
}

func podGroupNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base = addKubernetesLabelAndRank(base, n)
	base.Stack = true

	// NB: pods are the highest aggregation level for which we display
	// counts.
	count := pluralize(n.Counters, report.Pod, "pod", "pods")
	if typeName, ok := podGroupNodeTypeName[n.Topology]; ok {
		base.LabelMinor = fmt.Sprintf("%s of %s", typeName, count)
	} else {
		base.LabelMinor = count
	}

	return base, true
}

func ecsTaskNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(awsecs.TaskFamily)
	return base, true
}

func ecsServiceNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	_, base.Label, _ = report.ParseECSServiceNodeID(n.ID)
	base.Stack = true
	return base, true
}

func swarmServiceNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(docker.ServiceName)
	return base, true
}

func hostNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	var (
		hostname, _ = n.Latest.Lookup(host.HostName)
		parts       = strings.SplitN(hostname, ".", 2)
	)

	if len(parts) == 2 {
		base.Label, base.LabelMinor, base.Rank = parts[0], parts[1], parts[1]
	} else {
		base.Label = hostname
	}

	return base, true
}

func weaveNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	var (
		nickname, _ = n.Latest.Lookup(overlay.WeavePeerNickName)
	)

	_, peerName := report.ParseOverlayNodeID(n.ID)

	base.Label, base.LabelMinor = nickname, peerName

	return base, true
}

// groupNodeSummary renders the summary for a group node. n.Topology is
// expected to be of the form: group:container:hostname
func groupNodeSummary(base NodeSummary, r report.Report, n report.Node) (NodeSummary, bool) {
	parts := strings.Split(n.Topology, ":")
	if len(parts) != 3 {
		return NodeSummary{}, false
	}

	label, ok := n.Latest.Lookup(parts[2])
	if !ok {
		return NodeSummary{}, false
	}
	base.Label, base.Rank = label, label

	t, ok := r.Topology(parts[1])
	if ok && t.Label != "" {
		base.LabelMinor = pluralize(n.Counters, parts[1], t.Label, t.LabelPlural)
	}

	base.Shape = t.GetShape()
	base.Stack = true
	return base, true
}

func pluralize(counters report.Counters, key, singular, plural string) string {
	c, ok := counters.Lookup(key)
	if !ok {
		c = 0
	}
	if c == 1 {
		return fmt.Sprintf("%d %s", c, singular)
	}
	return fmt.Sprintf("%d %s", c, plural)
}

type nodeSummariesByID []NodeSummary

func (s nodeSummariesByID) Len() int           { return len(s) }
func (s nodeSummariesByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s nodeSummariesByID) Less(i, j int) bool { return s[i].ID < s[j].ID }

// NodeSummaries is a set of NodeSummaries indexed by ID.
type NodeSummaries map[string]NodeSummary

// Summaries converts RenderableNodes into a set of NodeSummaries
func Summaries(r report.Report, rns report.Nodes) NodeSummaries {

	result := NodeSummaries{}
	for id, node := range rns {
		if summary, ok := MakeNodeSummary(r, node); ok {
			for i, m := range summary.Metrics {
				summary.Metrics[i] = m.Summary()
			}
			result[id] = summary
		}
	}
	return result
}

// getRenderableContainerName obtains a user-friendly container name, to render in the UI
func getRenderableContainerName(nmd report.Node) string {
	for _, key := range []string{
		// Amazon's ecs-agent produces huge Docker container names, destructively
		// derived from mangling Container Definition names in Task
		// Definitions.
		//
		// However, the ecs-agent provides a label containing the original Container
		// Definition name.
		docker.LabelPrefix + AmazonECSContainerNameLabel,
		// Kubernetes also mangles its Docker container names and provides a
		// label with the original container name. However, note that this label
		// is only provided by Kubernetes versions >= 1.2 (see
		// https://github.com/kubernetes/kubernetes/pull/17234/ )
		docker.LabelPrefix + KubernetesContainerNameLabel,
		// Marathon doesn't set any Docker labels and this is the only meaningful
		// attribute we can find to make Scope useful without Mesos plugin
		docker.EnvPrefix + MarathonAppIDEnv,
		docker.ContainerName,
		docker.ContainerHostname,
	} {
		if label, ok := nmd.Latest.Lookup(key); ok {
			return label
		}
	}
	return ""
}
