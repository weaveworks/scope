package detailed

import (
	"sort"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

// Node is the data type that's yielded to the JavaScript layer when
// we want deep information about an individual node.
type Node struct {
	NodeSummary
	Rank     string             `json:"rank,omitempty"`
	Pseudo   bool               `json:"pseudo,omitempty"`
	Controls []ControlInstance  `json:"controls"`
	Children []NodeSummaryGroup `json:"children,omitempty"`
	Parents  []Parent           `json:"parents,omitempty"`
}

// ControlInstance contains a control description, and all the info
// needed to execute it.
type ControlInstance struct {
	ProbeID string
	NodeID  string
	Control report.Control
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (ControlInstance) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*ControlInstance) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}

type wiredControlInstance struct {
	ProbeID string `json:"probeId"`
	NodeID  string `json:"nodeId"`
	ID      string `json:"id"`
	Human   string `json:"human"`
	Icon    string `json:"icon"`
}

// CodecEncodeSelf marshals this MetricRow. It takes the basic Metric
// rendering, then adds some row-specific fields.
func (c *ControlInstance) CodecEncodeSelf(encoder *codec.Encoder) {
	encoder.Encode(wiredControlInstance{
		ProbeID: c.ProbeID,
		NodeID:  c.NodeID,
		ID:      c.Control.ID,
		Human:   c.Control.Human,
		Icon:    c.Control.Icon,
	})
}

// CodecDecodeSelf implements codec.Selfer
func (c *ControlInstance) CodecDecodeSelf(decoder *codec.Decoder) {
	var in wiredControlInstance
	decoder.Decode(&in)
	*c = ControlInstance{
		ProbeID: in.ProbeID,
		NodeID:  in.NodeID,
		Control: report.Control{
			ID:    in.ID,
			Human: in.Human,
			Icon:  in.Icon,
		},
	}
}

// MakeNode transforms a renderable node to a detailed node. It uses
// aggregate metadata, plus the set of origin node IDs, to produce tables.
func MakeNode(r report.Report, n render.RenderableNode) Node {
	summary, _ := MakeNodeSummary(n.Node)
	summary.ID = n.ID
	summary.Label = n.LabelMajor
	return Node{
		NodeSummary: summary,
		Rank:        n.Rank,
		Pseudo:      n.Pseudo,
		Controls:    controls(r, n),
		Children:    children(n),
		Parents:     Parents(r, n),
	}
}

func controlsFor(topology report.Topology, nodeID string) []ControlInstance {
	result := []ControlInstance{}
	node, ok := topology.Nodes[nodeID]
	if !ok {
		return result
	}

	for _, id := range node.Controls.Controls {
		if control, ok := topology.Controls[id]; ok {
			probeID, ok := node.Latest.Lookup(report.ControlProbeID)
			if !ok {
				continue
			}
			result = append(result, ControlInstance{
				ProbeID: probeID,
				NodeID:  nodeID,
				Control: control,
			})
		}
	}
	return result
}

func controls(r report.Report, n render.RenderableNode) []ControlInstance {
	if _, ok := r.Process.Nodes[n.ControlNode]; ok {
		return controlsFor(r.Process, n.ControlNode)
	} else if _, ok := r.Container.Nodes[n.ControlNode]; ok {
		return controlsFor(r.Container, n.ControlNode)
	} else if _, ok := r.ContainerImage.Nodes[n.ControlNode]; ok {
		return controlsFor(r.ContainerImage, n.ControlNode)
	} else if _, ok := r.Host.Nodes[n.ControlNode]; ok {
		return controlsFor(r.Host, n.ControlNode)
	}
	return []ControlInstance{}
}

var (
	nodeSummaryGroupSpecs = []struct {
		topologyID string
		NodeSummaryGroup
	}{
		{report.Host, NodeSummaryGroup{TopologyID: "hosts", Label: "Hosts", Columns: []Column{host.CPUUsage, host.MemoryUsage}}},
		{report.Pod, NodeSummaryGroup{TopologyID: "pods", Label: "Pods"}},
		{report.Container, NodeSummaryGroup{TopologyID: "containers", Label: "Containers", Columns: []Column{docker.CPUTotalUsage, docker.MemoryUsage}}},
		{report.Process, NodeSummaryGroup{TopologyID: "processes", Label: "Processes", Columns: []Column{process.PID, process.CPUUsage, process.MemoryUsage}}},
		{report.ContainerImage, NodeSummaryGroup{TopologyID: "containers-by-image", Label: "Container Images", Columns: []Column{render.ContainersKey}}},
	}
)

func children(n render.RenderableNode) []NodeSummaryGroup {
	summaries := map[string][]NodeSummary{}
	n.Children.ForEach(func(child report.Node) {
		if child.ID != n.ID {
			if summary, ok := MakeNodeSummary(child); ok {
				summaries[child.Topology] = append(summaries[child.Topology], summary)
			}
		}
	})

	nodeSummaryGroups := []NodeSummaryGroup{}
	for _, spec := range nodeSummaryGroupSpecs {
		if len(summaries[spec.topologyID]) > 0 {
			sort.Sort(nodeSummariesByID(summaries[spec.TopologyID]))
			group := spec.NodeSummaryGroup.Copy()
			group.Nodes = summaries[spec.topologyID]
			nodeSummaryGroups = append(nodeSummaryGroups, group)
		}
	}
	return nodeSummaryGroups
}
