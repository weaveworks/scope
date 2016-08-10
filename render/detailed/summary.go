package detailed

import (
	"fmt"
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
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

// Copy returns a value copy of the NodeSummaryGroup
func (g NodeSummaryGroup) Copy() NodeSummaryGroup {
	result := NodeSummaryGroup{
		TopologyID: g.TopologyID,
		Label:      g.Label,
		Columns:    g.Columns,
	}
	for _, node := range g.Nodes {
		result.Nodes = append(result.Nodes, node.Copy())
	}
	return result
}

// Column provides special json serialization for column ids, so they include
// their label for the frontend.
type Column struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	DefaultSort bool   `json:"defaultSort"`
}

// NodeSummary is summary information about a child for a Node.
type NodeSummary struct {
	ID         string               `json:"id"`
	Label      string               `json:"label"`
	LabelMinor string               `json:"label_minor"`
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

// MakeNodeSummary summarizes a node, if possible.
func MakeNodeSummary(r report.Report, n report.Node) (NodeSummary, bool) {
	renderers := map[string]func(NodeSummary, report.Node) (NodeSummary, bool){
		render.Pseudo:         pseudoNodeSummary,
		report.Process:        processNodeSummary,
		report.Container:      containerNodeSummary,
		report.ContainerImage: containerImageNodeSummary,
		report.Pod:            podNodeSummary,
		report.Service:        serviceNodeSummary,
		report.Deployment:     deploymentNodeSummary,
		report.ReplicaSet:     replicaSetNodeSummary,
		report.Host:           hostNodeSummary,
	}
	if renderer, ok := renderers[n.Topology]; ok {
		return renderer(baseNodeSummary(r, n), n)
	}
	if strings.HasPrefix(n.Topology, "group:") {
		return groupNodeSummary(baseNodeSummary(r, n), r, n)
	}
	return NodeSummary{}, false
}

// SummarizeMetrics returns a copy of the NodeSummary where the metrics are
// replaced with their summaries
func (n NodeSummary) SummarizeMetrics() NodeSummary {
	cp := n.Copy()
	for i, m := range cp.Metrics {
		cp.Metrics[i] = m.Summary()
	}
	return cp
}

// Copy returns a value copy of the NodeSummary
func (n NodeSummary) Copy() NodeSummary {
	result := NodeSummary{
		ID:         n.ID,
		Label:      n.Label,
		LabelMinor: n.LabelMinor,
		Rank:       n.Rank,
		Shape:      n.Shape,
		Stack:      n.Stack,
		Linkable:   n.Linkable,
		Adjacency:  n.Adjacency.Copy(),
	}
	for _, row := range n.Metadata {
		result.Metadata = append(result.Metadata, row.Copy())
	}
	for _, table := range n.Tables {
		result.Tables = append(result.Tables, table.Copy())
	}
	for _, row := range n.Metrics {
		result.Metrics = append(result.Metrics, row.Copy())
	}
	return result
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
		Adjacency: n.Adjacency.Copy(),
	}
}

func pseudoNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Pseudo = true
	base.Rank = n.ID

	if template, ok := map[string]struct{ Label, LabelMinor string }{
		render.TheInternetID:      {render.InboundMajor, ""},
		render.IncomingInternetID: {render.InboundMajor, render.InboundMinor},
		render.OutgoingInternetID: {render.OutboundMajor, render.OutboundMinor},
	}[n.ID]; ok {
		base.Label = template.Label
		base.LabelMinor = template.LabelMinor
		base.Shape = report.Cloud
		return base, true
	}

	// try rendering it as an uncontained node
	if strings.HasPrefix(n.ID, render.MakePseudoNodeID(render.UncontainedID)) {
		base.Label = render.UncontainedMajor
		base.LabelMinor = report.ExtractHostID(n)
		base.Shape = report.Square
		base.Stack = true
		return base, true
	}

	// try rendering it as an unmanaged node
	if strings.HasPrefix(n.ID, render.MakePseudoNodeID(render.UnmanagedID)) {
		base.Label = render.UnmanagedMajor
		base.Shape = report.Square
		base.Stack = true
		base.LabelMinor = report.ExtractHostID(n)
		return base, true
	}

	// try rendering it as an endpoint
	if addr, ok := n.Latest.Lookup(endpoint.Addr); ok {
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

	if c, ok := n.Counters.Lookup(report.Container); ok {
		if c == 1 {
			base.LabelMinor = fmt.Sprintf("%d container", c)
		} else {
			base.LabelMinor = fmt.Sprintf("%d containers", c)
		}
	}
	return base, true
}

func podNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(kubernetes.Name)
	namespace, _ := n.Latest.Lookup(kubernetes.Namespace)
	base.Rank = namespace + "/" + base.Label
	if c, ok := n.Counters.Lookup(report.Container); ok {
		if c == 1 {
			base.LabelMinor = fmt.Sprintf("%d container", c)
		} else {
			base.LabelMinor = fmt.Sprintf("%d containers", c)
		}
	}

	return base, true
}

func serviceNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(kubernetes.Name)
	namespace, _ := n.Latest.Lookup(kubernetes.Namespace)
	base.Rank = namespace + "/" + base.Label
	base.Stack = true

	// Services are always just a group of pods, so there's no counting multiple
	// services which might be grouped together.
	if p, ok := n.Counters.Lookup(report.Pod); ok {
		if p == 1 {
			base.LabelMinor = fmt.Sprintf("%d pod", p)
		} else {
			base.LabelMinor = fmt.Sprintf("%d pods", p)
		}
	}

	return base, true
}

func deploymentNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(kubernetes.Name)
	namespace, _ := n.Latest.Lookup(kubernetes.Namespace)
	base.Rank = namespace + "/" + base.Label
	base.Stack = true

	if p, ok := n.Counters.Lookup(report.Pod); ok {
		if p == 1 {
			base.LabelMinor = fmt.Sprintf("%d pod", p)
		} else {
			base.LabelMinor = fmt.Sprintf("%d pods", p)
		}
	}

	return base, true
}

func replicaSetNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(kubernetes.Name)
	namespace, _ := n.Latest.Lookup(kubernetes.Namespace)
	base.Rank = namespace + "/" + base.Label
	base.Stack = true

	if p, ok := n.Counters.Lookup(report.Pod); ok {
		if p == 1 {
			base.LabelMinor = fmt.Sprintf("%d pod", p)
		} else {
			base.LabelMinor = fmt.Sprintf("%d pods", p)
		}
	}

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
		if count, ok := n.Counters.Lookup(parts[1]); ok {
			if count == 1 {
				base.LabelMinor = fmt.Sprintf("%d %s", count, t.Label)
			} else {
				base.LabelMinor = fmt.Sprintf("%d %s", count, t.LabelPlural)
			}
		}
	}

	base.Shape = t.GetShape()
	base.Stack = true
	return base, true
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

// Copy returns a deep value-copy of NodeSummaries
func (n NodeSummaries) Copy() NodeSummaries {
	result := NodeSummaries{}
	for k, v := range n {
		result[k] = v.Copy()
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
