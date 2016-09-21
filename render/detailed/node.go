package detailed

import (
	"sort"
	"time"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Node is the data type that's yielded to the JavaScript layer when
// we want deep information about an individual node.
type Node struct {
	NodeSummary
	Controls    []ControlInstance    `json:"controls"`
	Children    []NodeSummaryGroup   `json:"children,omitempty"`
	Connections []ConnectionsSummary `json:"connections,omitempty"`
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
	Rank    int    `json:"rank"`
}

// CodecEncodeSelf marshals this ControlInstance. It takes the basic Metric
// rendering, then adds some row-specific fields.
func (c *ControlInstance) CodecEncodeSelf(encoder *codec.Encoder) {
	encoder.Encode(wiredControlInstance{
		ProbeID: c.ProbeID,
		NodeID:  c.NodeID,
		ID:      c.Control.ID,
		Human:   c.Control.Human,
		Icon:    c.Control.Icon,
		Rank:    c.Control.Rank,
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
			Rank:  in.Rank,
		},
	}
}

// MakeNode transforms a renderable node to a detailed node. It uses
// aggregate metadata, plus the set of origin node IDs, to produce tables.
func MakeNode(topologyID string, r report.Report, ns report.Nodes, n report.Node) Node {
	summary, _ := MakeNodeSummary(r, n)
	return Node{
		NodeSummary: summary,
		Controls:    controls(r, n),
		Children:    children(r, n),
		Connections: []ConnectionsSummary{
			incomingConnectionsSummary(topologyID, r, n, ns),
			outgoingConnectionsSummary(topologyID, r, n, ns),
		},
	}
}

func controlsFor(topology report.Topology, nodeID string) []ControlInstance {
	result := []ControlInstance{}
	node, ok := topology.Nodes[nodeID]
	if !ok {
		return result
	}
	probeID, ok := node.Latest.Lookup(report.ControlProbeID)
	if !ok {
		return result
	}
	node.LatestControls.ForEach(func(controlID string, _ time.Time, data report.NodeControlData) {
		if data.Dead {
			return
		}
		if control, ok := topology.Controls[controlID]; ok {
			result = append(result, ControlInstance{
				ProbeID: probeID,
				NodeID:  nodeID,
				Control: control,
			})
		}
	})
	return result
}

func controls(r report.Report, n report.Node) []ControlInstance {
	if t, ok := r.Topology(n.Topology); ok {
		return controlsFor(t, n.ID)
	}
	return []ControlInstance{}
}

var (
	nodeSummaryGroupSpecs = []struct {
		topologyID string
		NodeSummaryGroup
	}{
		{
			topologyID: report.Host,
			NodeSummaryGroup: NodeSummaryGroup{
				TopologyID: "hosts",
				Label:      "Hosts",
				Columns: []Column{
					{ID: host.CPUUsage, Label: "CPU", DataType: "number"},
					{ID: host.MemoryUsage, Label: "Memory", DataType: "number"},
				},
			},
		},
		{
			topologyID: report.Service,
			NodeSummaryGroup: NodeSummaryGroup{
				TopologyID: "services",
				Label:      "Services",
				Columns: []Column{
					{ID: report.Pod, Label: "# Pods", DataType: "number"},
					{ID: kubernetes.IP, Label: "IP"},
				},
			},
		},
		{
			topologyID: report.ReplicaSet,
			NodeSummaryGroup: NodeSummaryGroup{
				TopologyID: "replica-sets",
				Label:      "Replica Sets",
				Columns: []Column{
					{ID: report.Pod, Label: "# Pods", DataType: "number"},
					{ID: kubernetes.ObservedGeneration, Label: "Observed Gen.", DataType: "number"},
				},
			},
		},
		{
			topologyID: report.Pod,
			NodeSummaryGroup: NodeSummaryGroup{
				TopologyID: "pods",
				Label:      "Pods",

				Columns: []Column{
					{ID: kubernetes.State, Label: "State"},
					{ID: report.Container, Label: "# Containers", DataType: "number"},
					{ID: kubernetes.IP, Label: "IP"},
				},
			},
		},
		{
			topologyID: report.Container,
			NodeSummaryGroup: NodeSummaryGroup{
				TopologyID: "containers",
				Label:      "Containers", Columns: []Column{
					{ID: docker.CPUTotalUsage, Label: "CPU", DataType: "number"},
					{ID: docker.MemoryUsage, Label: "Memory", DataType: "number"},
				},
			},
		},
		{
			topologyID: report.Process,
			NodeSummaryGroup: NodeSummaryGroup{
				TopologyID: "processes",
				Label:      "Processes", Columns: []Column{
					{ID: process.PID, Label: "PID", DataType: "number"},
					{ID: process.CPUUsage, Label: "CPU", DataType: "number"},
					{ID: process.MemoryUsage, Label: "Memory", DataType: "number"},
				},
			},
		},
		{
			topologyID: report.ContainerImage,
			NodeSummaryGroup: NodeSummaryGroup{
				TopologyID: "containers-by-image",
				Label:      "Container Images",
				Columns: []Column{
					{ID: report.Container, Label: "# Containers", DefaultSort: true, DataType: "number"},
				},
			},
		},
	}
)

func children(r report.Report, n report.Node) []NodeSummaryGroup {
	summaries := map[string][]NodeSummary{}
	n.Children.ForEach(func(child report.Node) {
		if child.ID == n.ID {
			return
		}
		summary, ok := MakeNodeSummary(r, child)
		if !ok {
			return
		}
		summaries[child.Topology] = append(summaries[child.Topology], summary.SummarizeMetrics())
	})

	nodeSummaryGroups := []NodeSummaryGroup{}
	for _, spec := range nodeSummaryGroupSpecs {
		if len(summaries[spec.topologyID]) > 0 {
			sort.Sort(nodeSummariesByID(summaries[spec.TopologyID]))
			group := spec.NodeSummaryGroup
			group.Nodes = summaries[spec.topologyID]
			nodeSummaryGroups = append(nodeSummaryGroups, group)
		}
	}

	return nodeSummaryGroups
}
