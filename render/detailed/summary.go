package detailed

import (
	"fmt"
	"strings"

	"github.com/weaveworks/scope/probe/awsecs"
	"github.com/weaveworks/scope/probe/docker"
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
	report.StatefulSet:    podGroupNodeSummary,
	report.CronJob:        podGroupNodeSummary,
	report.ECSTask:        ecsTaskNodeSummary,
	report.ECSService:     ecsServiceNodeSummary,
	report.SwarmService:   swarmServiceNodeSummary,
	report.Host:           hostNodeSummary,
	report.Overlay:        weaveNodeSummary,
	report.Endpoint:       nil, // Do not render
}

var templates = map[string]struct{ Label, LabelMinor string }{
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
	report.StatefulSet:    "kube-controllers",
	report.CronJob:        "kube-controllers",
	report.Service:        "services",
	report.ECSTask:        "ecs-tasks",
	report.ECSService:     "ecs-services",
	report.SwarmService:   "swarm-services",
	report.Host:           "hosts",
}

// MakeNodeSummary summarizes a node, if possible.
func MakeNodeSummary(rc report.RenderContext, n report.Node) (NodeSummary, bool) {
	r := rc.Report
	if renderer, ok := renderers[n.Topology]; ok {
		// Skip (and don't fall through to fallback) if renderer maps to nil
		if renderer != nil {
			summary, b := renderer(baseNodeSummary(r, n), n)
			return RenderMetricURLs(summary, n, rc.MetricsGraphURL), b
		}
	} else if _, ok := rc.Topology(n.Topology); ok {
		summary := baseNodeSummary(r, n)
		summary.Label = n.ID // This is unlikely to look very good, but is a reasonable fallback
		return summary, true
	}
	if strings.HasPrefix(n.Topology, "group:") {
		summary, b := groupNodeSummary(baseNodeSummary(r, n), r, n)
		return RenderMetricURLs(summary, n, rc.MetricsGraphURL), b
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
	summary := NodeSummary{
		ID:        n.ID,
		Linkable:  true,
		Parents:   Parents(r, n),
		Adjacency: n.Adjacency,
	}
	if t, ok := r.Topology(n.Topology); ok {
		summary.Shape = t.GetShape()
	}
	if _, ok := n.Counters.Lookup(n.Topology); ok {
		// This is a group of nodes, so no metadata, metrics, tables
		return summary
	}
	if topology, ok := r.Topology(n.Topology); ok {
		summary.Metadata = topology.MetadataTemplates.MetadataRows(n)
		summary.Metrics = topology.MetricTemplates.MetricRows(n)
		summary.Tables = topology.TableTemplates.Tables(n)
	}
	return summary
}

func pseudoNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	pseudoID, _ := render.ParsePseudoNodeID(n.ID)
	base.Pseudo = true
	base.Rank = pseudoID

	switch {
	case render.IsInternetNode(n):
		// render as an internet node
		template := templates[n.ID]
		base.Label = template.Label
		base.LabelMinor = template.LabelMinor
		base.Shape = report.Cloud
	case strings.HasPrefix(n.ID, render.ServiceNodeIDPrefix):
		// render as a known service node
		base.Label = n.ID[len(render.ServiceNodeIDPrefix):]
		base.LabelMinor = ""
		base.Shape = report.Cloud
	case strings.HasPrefix(n.ID, render.UncontainedIDPrefix):
		// render as an uncontained node
		base.Label = render.UncontainedMajor
		base.LabelMinor = report.ExtractHostID(n)
		base.Shape = report.Square
		base.Stack = true
	case strings.HasPrefix(n.ID, render.UnmanagedIDPrefix):
		// render as an unmanaged node
		base.Label = render.UnmanagedMajor
		base.LabelMinor = report.ExtractHostID(n)
		base.Shape = report.Square
		base.Stack = true
	default:
		// try rendering it as an endpoint
		if _, addr, _, ok := report.ParseEndpointNodeID(n.ID); ok {
			base.Label = addr
			base.Shape = report.Circle
		} else {
			// last resort
			base.Label = pseudoID
		}
	}
	return base, true
}

func processNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	var (
		hostID, pid, _   = report.ParseProcessNodeID(n.ID)
		processName, _   = n.Latest.Lookup(process.Name)
		containerName, _ = n.Latest.Lookup(docker.ContainerName)
	)
	switch {
	case processName != "" && containerName != "":
		base.Label = processName
		base.LabelMinor = fmt.Sprintf("%s (%s:%s)", hostID, containerName, pid)
		base.Rank = processName
	case processName != "":
		base.Label = processName
		base.LabelMinor = fmt.Sprintf("%s (%s)", hostID, pid)
		base.Rank = processName
	case containerName != "":
		base.Label = pid
		base.LabelMinor = fmt.Sprintf("%s (%s)", hostID, containerName)
		base.Rank = hostID
	default:
		base.Label = pid
		base.LabelMinor = hostID
		base.Rank = hostID
	}
	base.Linkable = render.IsConnected(n)
	return base, true
}

func containerNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	var (
		containerName = getRenderableContainerName(n)
		hostName      = report.ExtractHostID(n)
		imageName, _  = n.Latest.Lookup(docker.ImageName)
	)
	base.Label = containerName
	base.LabelMinor = hostName
	if imageName != "" {
		base.Rank = docker.ImageNameWithoutVersion(imageName)
	} else if hostName != "" {
		base.Rank = hostName
	} else {
		base.Rank = base.Label
	}
	return base, true
}

func containerImageNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	var (
		imageName, _            = n.Latest.Lookup(docker.ImageName)
		imageNameWithoutVersion = docker.ImageNameWithoutVersion(imageName)
	)
	switch {
	case imageNameWithoutVersion != "" && imageNameWithoutVersion != ImageNameNone:
		base.Label = imageNameWithoutVersion
	case imageName != "" && imageName != ImageNameNone:
		base.Label = imageName
	default:
		// The id can be an image id or an image name. Ideally we'd
		// truncate the former but not the latter, but short of
		// heuristic regexp match we cannot tell the difference.
		base.Label, _ = report.ParseContainerImageNodeID(n.ID)
	}
	base.LabelMinor = pluralize(n.Counters, report.Container, "container", "containers")
	base.Rank = base.Label
	base.Stack = true
	return base, true
}

func addKubernetesLabelAndRank(base NodeSummary, n report.Node) NodeSummary {
	var (
		name, _      = n.Latest.Lookup(kubernetes.Name)
		namespace, _ = n.Latest.Lookup(kubernetes.Namespace)
	)
	if name != "" {
		base.Label = name
	} else {
		base.Label, _, _ = report.ParseNodeID(n.ID)
	}
	base.Rank = namespace + "/" + base.Label
	return base
}

func podNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base = addKubernetesLabelAndRank(base, n)
	base.LabelMinor = pluralize(n.Counters, report.Container, "container", "containers")

	return base, true
}

var podGroupNodeTypeName = map[string]string{
	report.Deployment:  "Deployment",
	report.DaemonSet:   "DaemonSet",
	report.StatefulSet: "StatefulSet",
	report.CronJob:     "CronJob",
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
	if base.Label == "" {
		base.Label, _ = report.ParseECSTaskNodeID(n.ID)
	}
	return base, true
}

func ecsServiceNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	_, base.Label, _ = report.ParseECSServiceNodeID(n.ID)
	base.Stack = true
	return base, true
}

func swarmServiceNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(docker.ServiceName)
	if base.Label == "" {
		base.Label, _ = report.ParseSwarmServiceNodeID(n.ID)
	}
	return base, true
}

func hostNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	var (
		hostname, _ = report.ParseHostNodeID(n.ID)
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
		_, peerName = report.ParseOverlayNodeID(n.ID)
	)
	if nickname != "" {
		base.Label = nickname
	} else {
		base.Label = peerName
	}
	base.LabelMinor = peerName
	return base, true
}

// groupNodeSummary renders the summary for a group node. n.Topology is
// expected to be of the form: group:container:hostname
func groupNodeSummary(base NodeSummary, r report.Report, n report.Node) (NodeSummary, bool) {
	topology, key, ok := render.ParseGroupNodeTopology(n.Topology)
	if !ok {
		return NodeSummary{}, false
	}

	label, ok := n.Latest.Lookup(key)
	if !ok {
		return NodeSummary{}, false
	}
	base.Label, base.Rank = label, label

	if t, ok := r.Topology(topology); ok {
		base.Shape = t.GetShape()
		if t.Label != "" {
			base.LabelMinor = pluralize(n.Counters, topology, t.Label, t.LabelPlural)
		}
	}
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
func Summaries(rc report.RenderContext, rns report.Nodes) NodeSummaries {

	result := NodeSummaries{}
	for id, node := range rns {
		if summary, ok := MakeNodeSummary(rc, node); ok {
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
	containerID, _ := report.ParseContainerNodeID(nmd.ID)
	if len(containerID) > 12 {
		containerID = containerID[:12]
	}
	return containerID
}
