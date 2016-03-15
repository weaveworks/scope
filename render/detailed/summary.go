package detailed

import (
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
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

// MakeColumn makes a Column by looking up the label by id.
func MakeColumn(id string) Column {
	return Column{
		ID:    id,
		Label: Label(id),
	}
}

// NodeSummary is summary information about a child for a Node.
type NodeSummary struct {
	ID           string        `json:"id"`
	Label        string        `json:"label"`
	LabelMinor   string        `json:"label_minor"`
	Rank         string        `json:"rank"`
	Shape        string        `json:"shape,omitempty"`
	Stack        bool          `json:"stack,omitempty"`
	Linkable     bool          `json:"linkable,omitempty"` // Whether this node can be linked-to
	Pseudo       bool          `json:"pseudo,omitempty"`
	Metadata     []MetadataRow `json:"metadata,omitempty"`
	DockerLabels []MetadataRow `json:"docker_labels,omitempty"`
	Metrics      []MetricRow   `json:"metrics,omitempty"`
	Adjacency    report.IDList `json:"adjacency,omitempty"`
}

// MakeNodeSummary summarizes a node, if possible.
func MakeNodeSummary(n render.RenderableNode) (NodeSummary, bool) {
	renderers := map[string]func(report.Node) NodeSummary{
		report.Process:        processNodeSummary,
		report.Container:      containerNodeSummary,
		report.ContainerImage: containerImageNodeSummary,
		report.Pod:            podNodeSummary,
		report.Host:           hostNodeSummary,
	}
	var summary NodeSummary
	if renderer, ok := renderers[n.Topology]; ok {
		summary = renderer(n.Node)
	} else if n.Pseudo {
		summary = pseudoNodeSummary(n.Node)
	} else {
		return NodeSummary{}, false
	}
	summary.ID = n.ID
	summary.Label = n.Label
	summary.LabelMinor = n.LabelMinor
	summary.Rank = n.Rank
	summary.Shape = n.Shape
	summary.Stack = n.Stack
	summary.Adjacency = n.Node.Adjacency.Copy()
	return summary, true
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
	for _, row := range n.DockerLabels {
		result.DockerLabels = append(result.DockerLabels, row.Copy())
	}
	for _, row := range n.Metrics {
		result.Metrics = append(result.Metrics, row.Copy())
	}
	return result
}

func baseNodeSummary(linkable bool, nmd report.Node) NodeSummary {
	return NodeSummary{
		Linkable:     linkable,
		Metadata:     NodeMetadata(nmd),
		DockerLabels: NodeDockerLabels(nmd),
		Metrics:      NodeMetrics(nmd),
	}
}

func pseudoNodeSummary(nmd report.Node) NodeSummary {
	n := baseNodeSummary(true, nmd)
	n.Pseudo = true
	return n
}

func processNodeSummary(nmd report.Node) NodeSummary {
	_, isConnected := nmd.Latest.Lookup(render.IsConnected)
	return baseNodeSummary(isConnected, nmd)
}

func containerNodeSummary(nmd report.Node) NodeSummary      { return baseNodeSummary(true, nmd) }
func containerImageNodeSummary(nmd report.Node) NodeSummary { return baseNodeSummary(true, nmd) }
func podNodeSummary(nmd report.Node) NodeSummary            { return baseNodeSummary(true, nmd) }
func hostNodeSummary(nmd report.Node) NodeSummary           { return baseNodeSummary(true, nmd) }

type nodeSummariesByID []NodeSummary

func (s nodeSummariesByID) Len() int           { return len(s) }
func (s nodeSummariesByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s nodeSummariesByID) Less(i, j int) bool { return s[i].ID < s[j].ID }

type NodeSummaries map[string]NodeSummary

func Summaries(rns render.RenderableNodes) NodeSummaries {
	result := NodeSummaries{}
	for id, node := range rns {
		if summary, ok := MakeNodeSummary(node); ok {
			for i, m := range summary.Metrics {
				summary.Metrics[i] = m.Summary()
			}
			result[id] = summary
		}
	}
	return result
}
