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
	Circle   = "circle"
	Square   = "square"
	Heptagon = "heptagon"
	Hexagon  = "hexagon"
	Cloud    = "cloud"
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
func MakeNodeSummary(n report.Node) (NodeSummary, bool) {
	renderers := map[string]func(NodeSummary, report.Node) (NodeSummary, bool){
		render.Pseudo:         pseudoNodeSummary,
		report.Process:        processNodeSummary,
		report.Container:      containerNodeSummary,
		report.ContainerImage: containerImageNodeSummary,
		report.Pod:            podNodeSummary,
		report.Service:        serviceNodeSummary,
		report.Host:           hostNodeSummary,
	}
	if renderer, ok := renderers[n.Topology]; ok {
		if summary, ok := baseNodeSummary(NodeSummary{}, n); ok {
			return renderer(summary, n)
		}
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
	for _, row := range n.DockerLabels {
		result.DockerLabels = append(result.DockerLabels, row.Copy())
	}
	for _, row := range n.Metrics {
		result.Metrics = append(result.Metrics, row.Copy())
	}
	return result
}

func baseNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.ID = n.ID
	base.Shape = Circle
	base.Linkable = true
	base.Metadata = NodeMetadata(n)
	base.DockerLabels = NodeDockerLabels(n)
	base.Metrics = NodeMetrics(n)
	base.Adjacency = n.Adjacency.Copy()
	return base, true
}

func pseudoNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Pseudo = true
	base.Rank = n.ID

	if template, ok := map[string]struct{ Label, LabelMinor, Shape string }{
		render.TheInternetID:      {render.InboundMajor, "", Cloud},
		render.IncomingInternetID: {render.InboundMajor, render.InboundMinor, Cloud},
		render.OutgoingInternetID: {render.OutboundMajor, render.OutboundMinor, Cloud},
	}[n.ID]; ok {
		base.Label = template.Label
		base.LabelMinor = template.LabelMinor
		base.Shape = template.Shape
		return base, true
	}

	// try rendering it as an uncontained node
	if strings.HasPrefix(n.ID, render.MakePseudoNodeID(render.UncontainedID)) {
		base.Label = render.UncontainedMajor
		base.Shape = Circle
		base.LabelMinor = report.ExtractHostID(n)
		return base, true
	}

	// try rendering it as an endpoint
	if addr, ok := n.Latest.Lookup(endpoint.Addr); ok {
		base.Label = addr
		return base, true
	}

	return NodeSummary{}, false
}

func processNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(process.Name)
	base.Rank, _ = n.Latest.Lookup(process.Name)
	base.Shape = Square

	if p, ok := n.Counters.Lookup(report.Process); ok {
		base.Stack = true
		if p == 1 {
			base.LabelMinor = fmt.Sprintf("%d process", p)
		} else {
			base.LabelMinor = fmt.Sprintf("%d processes", p)
		}
	} else {
		pid, ok := n.Latest.Lookup(process.PID)
		if !ok {
			return NodeSummary{}, false
		}
		if containerName, ok := n.Latest.Lookup(docker.ContainerName); ok {
			base.LabelMinor = fmt.Sprintf("%s (%s:%s)", report.ExtractHostID(n), containerName, pid)
		} else {
			base.LabelMinor = fmt.Sprintf("%s (%s)", report.ExtractHostID(n), pid)
		}
	}

	_, isConnected := n.Latest.Lookup(render.IsConnected)
	base.Linkable = isConnected
	return base, true
}

func containerNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = render.GetRenderableContainerName(n)

	if c, ok := n.Counters.Lookup(report.Container); ok {
		base.Stack = true
		if c == 1 {
			base.LabelMinor = fmt.Sprintf("%d container", c)
		} else {
			base.LabelMinor = fmt.Sprintf("%d containers", c)
		}
	} else {
		base.LabelMinor = report.ExtractHostID(n)
	}

	if imageName, ok := n.Latest.Lookup(docker.ImageName); ok {
		base.Rank = render.ImageNameWithoutVersion(imageName)
	}

	base.Shape = Hexagon
	return base, true
}

func containerImageNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	imageName, ok := n.Latest.Lookup(docker.ImageName)
	if !ok {
		return NodeSummary{}, false
	}

	imageNameWithoutVersion := render.ImageNameWithoutVersion(imageName)
	base.Label = imageNameWithoutVersion
	base.Rank = imageNameWithoutVersion
	base.Shape = Hexagon
	base.Stack = true

	if base.Label == "<none>" {
		base.Label, _ = n.Latest.Lookup(docker.ImageID)
		if len(base.Label) > 12 {
			base.Label = base.Label[:12]
		}
	}

	if i, ok := n.Counters.Lookup(report.ContainerImage); ok {
		if i == 1 {
			base.LabelMinor = fmt.Sprintf("%d image", i)
		} else {
			base.LabelMinor = fmt.Sprintf("%d images", i)
		}
	} else if c, ok := n.Counters.Lookup(report.Container); ok {
		if c == 1 {
			base.LabelMinor = fmt.Sprintf("%d container", c)
		} else {
			base.LabelMinor = fmt.Sprintf("%d containers", c)
		}
	}
	return base, true
}

func podNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(kubernetes.PodName)
	base.Rank, _ = n.Latest.Lookup(kubernetes.PodID)
	base.Shape = Heptagon

	if p, ok := n.Counters.Lookup(report.Pod); ok {
		base.Stack = true
		if p == 1 {
			base.LabelMinor = fmt.Sprintf("%d pod", p)
		} else {
			base.LabelMinor = fmt.Sprintf("%d pods", p)
		}
	} else if c, ok := n.Counters.Lookup(report.Container); ok {
		if c == 1 {
			base.LabelMinor = fmt.Sprintf("%d container", c)
		} else {
			base.LabelMinor = fmt.Sprintf("%d containers", c)
		}
	}

	return base, true
}

func serviceNodeSummary(base NodeSummary, n report.Node) (NodeSummary, bool) {
	base.Label, _ = n.Latest.Lookup(kubernetes.ServiceName)
	base.Rank, _ = n.Latest.Lookup(kubernetes.ServiceID)
	base.Shape = Heptagon
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

	if h, ok := n.Counters.Lookup(report.Host); ok {
		base.Stack = true
		if h == 1 {
			base.LabelMinor = fmt.Sprintf("%d host", h)
		} else {
			base.LabelMinor = fmt.Sprintf("%d hosts", h)
		}
	}

	base.Shape = Circle
	return base, true
}

type nodeSummariesByID []NodeSummary

func (s nodeSummariesByID) Len() int           { return len(s) }
func (s nodeSummariesByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s nodeSummariesByID) Less(i, j int) bool { return s[i].ID < s[j].ID }

// NodeSummaries is a set of NodeSummaries indexed by ID.
type NodeSummaries map[string]NodeSummary

// Summaries converts RenderableNodes into a set of NodeSummaries
func Summaries(rns report.Nodes) NodeSummaries {
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

// Copy returns a deep value-copy of NodeSummaries
func (n NodeSummaries) Copy() NodeSummaries {
	result := NodeSummaries{}
	for k, v := range n {
		result[k] = v.Copy()
	}
	return result
}
