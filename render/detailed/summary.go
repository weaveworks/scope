package detailed

import (
	"fmt"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

// NodeSummaryGroup is a topology-typed group of children for a Node.
type NodeSummaryGroup struct {
	Label      string        `json:"label"`
	Nodes      []NodeSummary `json:"nodes"`
	TopologyID string        `json:"topologyId"`
	Columns    []string      `json:"columns"`
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

// NodeSummary is summary information about a child for a Node.
type NodeSummary struct {
	ID       string        `json:"id"`
	Label    string        `json:"label"`
	Linkable bool          `json:"linkable"` // Whether this node can be linked-to
	Metadata []MetadataRow `json:"metadata,omitempty"`
	Metrics  []MetricRow   `json:"metrics,omitempty"`
}

// MakeNodeSummary summarizes a node, if possible.
func MakeNodeSummary(n report.Node) (NodeSummary, bool) {
	renderers := map[string]func(report.Node) NodeSummary{
		report.Process:        processNodeSummary,
		report.Container:      containerNodeSummary,
		report.ContainerImage: containerImageNodeSummary,
		report.Pod:            podNodeSummary,
		report.Host:           hostNodeSummary,
		"group":               groupNodeSummary,
	}
	if renderer, ok := renderers[n.Topology]; ok {
		return renderer(n), true
	}
	return NodeSummary{}, false
}

// Copy returns a value copy of the NodeSummary
func (n NodeSummary) Copy() NodeSummary {
	result := NodeSummary{
		ID:       n.ID,
		Label:    n.Label,
		Linkable: n.Linkable,
	}
	for _, row := range n.Metadata {
		result.Metadata = append(result.Metadata, row.Copy())
	}
	for _, row := range n.Metrics {
		result.Metrics = append(result.Metrics, row.Copy())
	}
	return result
}

func processNodeSummary(nmd report.Node) NodeSummary {
	var (
		id               string
		label, nameFound = nmd.Metadata[process.Name]
	)
	if pid, ok := nmd.Metadata[process.PID]; ok {
		if !nameFound {
			label = fmt.Sprintf("(%s)", pid)
		}
		id = render.MakeProcessID(report.ExtractHostID(nmd), pid)
	}
	_, isConnected := nmd.Metadata[render.IsConnected]
	return NodeSummary{
		ID:       id,
		Label:    label,
		Linkable: isConnected,
		Metadata: processNodeMetadata(nmd),
		Metrics:  processNodeMetrics(nmd),
	}
}

func containerNodeSummary(nmd report.Node) NodeSummary {
	label, _ := render.GetRenderableContainerName(nmd)
	return NodeSummary{
		ID:       render.MakeContainerID(nmd.Metadata[docker.ContainerID]),
		Label:    label,
		Linkable: true,
		Metadata: containerNodeMetadata(nmd),
		Metrics:  containerNodeMetrics(nmd),
	}
}

func containerImageNodeSummary(nmd report.Node) NodeSummary {
	imageName := nmd.Metadata[docker.ImageName]
	return NodeSummary{
		ID:       render.MakeContainerImageID(render.ImageNameWithoutVersion(imageName)),
		Label:    imageName,
		Linkable: true,
		Metadata: containerImageNodeMetadata(nmd),
	}
}

func groupNodeSummary(nmd report.Node) NodeSummary {
	key := nmd.Metadata["group_key"]
	value := nmd.Metadata["group_value"]
	label := nmd.Metadata["group_label"]
	return NodeSummary{
		ID:       render.MakeGroupID(key, value),
		Label:    label,
		Linkable: true,
		Metadata: groupNodeMetadata(nmd),
	}
}

func podNodeSummary(nmd report.Node) NodeSummary {
	return NodeSummary{
		ID:       render.MakePodID(nmd.Metadata[kubernetes.PodID]),
		Label:    nmd.Metadata[kubernetes.PodName],
		Linkable: true,
		Metadata: podNodeMetadata(nmd),
	}
}

func hostNodeSummary(nmd report.Node) NodeSummary {
	return NodeSummary{
		ID:       render.MakeHostID(nmd.Metadata[host.HostName]),
		Label:    nmd.Metadata[host.HostName],
		Linkable: true,
		Metadata: hostNodeMetadata(nmd),
		Metrics:  hostNodeMetrics(nmd),
	}
}

type nodeSummariesByID []NodeSummary

func (s nodeSummariesByID) Len() int           { return len(s) }
func (s nodeSummariesByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s nodeSummariesByID) Less(i, j int) bool { return s[i].ID < s[j].ID }
