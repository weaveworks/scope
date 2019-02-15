package detailed

import (
	"sort"
	"time"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/probe/awsecs"
	"github.com/weaveworks/scope/probe/docker"
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

// RenderContext carries contextual data that is needed when rendering parts of the report.
type RenderContext struct {
	report.Report
	MetricsGraphURL string
}

// MakeNode transforms a renderable node to a detailed node. It uses
// aggregate metadata, plus the set of origin node IDs, to produce tables.
func MakeNode(topologyID string, rc RenderContext, hideCommandLineArguments bool, ns report.Nodes, n report.Node) Node {
	summary, _ := MakeNodeSummary(rc, hideCommandLineArguments, n)
	return Node{
		NodeSummary: summary,
		Controls:    controls(rc.Report, n),
		Children:    children(rc, hideCommandLineArguments, n),
		Connections: []ConnectionsSummary{
			incomingConnectionsSummary(topologyID, rc.Report, n, ns),
			outgoingConnectionsSummary(topologyID, rc.Report, n, ns),
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

// We only need to include topologies here where the nodes may appear
// as children of other nodes in some topology.
var nodeSummaryGroupSpecs = []struct {
	topologyID string
	NodeSummaryGroup
}{
	{
		topologyID: report.Pod,
		NodeSummaryGroup: NodeSummaryGroup{
			Label: "Pods",
			Columns: []Column{
				{ID: kubernetes.State, Label: "State"},
				{ID: report.Container, Label: "# Containers", Datatype: report.Number},
				{ID: kubernetes.IP, Label: "IP", Datatype: report.IP},
			},
		},
	},
	{
		topologyID: report.ECSTask,
		NodeSummaryGroup: NodeSummaryGroup{
			Label: "Tasks",
			Columns: []Column{
				{ID: awsecs.CreatedAt, Label: "Created At", Datatype: report.DateTime},
			},
		},
	},
	{
		topologyID: report.Container,
		NodeSummaryGroup: NodeSummaryGroup{
			Label: "Containers",
			Columns: []Column{
				{ID: docker.CPUTotalUsage, Label: "CPU", Datatype: report.Number},
				{ID: docker.MemoryUsage, Label: "Memory", Datatype: report.Number},
			},
		},
	},
	{
		topologyID: report.Process,
		NodeSummaryGroup: NodeSummaryGroup{
			Label: "Processes",
			Columns: []Column{
				{ID: process.PID, Label: "PID", Datatype: report.Number},
				{ID: process.CPUUsage, Label: "CPU", Datatype: report.Number},
				{ID: process.MemoryUsage, Label: "Memory", Datatype: report.Number},
			},
		},
	},
	{
		topologyID: report.ContainerImage,
		NodeSummaryGroup: NodeSummaryGroup{
			Label:   "Container images",
			Columns: []Column{},
		},
	},
	{
		topologyID: report.PersistentVolume,
		NodeSummaryGroup: NodeSummaryGroup{
			Label:   "Persistent Volumes",
			Columns: []Column{},
		},
	},
	{
		topologyID: report.PersistentVolumeClaim,
		NodeSummaryGroup: NodeSummaryGroup{
			Label:   "Persistent Volume Claims",
			Columns: []Column{},
		},
	},
	{
		topologyID: report.StorageClass,
		NodeSummaryGroup: NodeSummaryGroup{
			Label:   "Storage Classes",
			Columns: []Column{},
		},
	},
	{
		topologyID: report.VolumeSnapshot,
		NodeSummaryGroup: NodeSummaryGroup{
			Label:   "Volume Snapshots",
			Columns: []Column{},
		},
	},
	{
		topologyID: report.VolumeSnapshotData,
		NodeSummaryGroup: NodeSummaryGroup{
			Label:   "Volume Snapshot Data",
			Columns: []Column{},
		},
	},
}

func children(rc RenderContext, hideCommandLineArguments bool, n report.Node) []NodeSummaryGroup {
	summaries := map[string][]NodeSummary{}
	n.Children.ForEach(func(child report.Node) {
		if child.ID == n.ID {
			return
		}
		summary, ok := MakeNodeSummary(rc, hideCommandLineArguments, child)
		if !ok {
			return
		}
		summaries[child.Topology] = append(summaries[child.Topology], summary.SummarizeMetrics())
	})

	nodeSummaryGroups := []NodeSummaryGroup{}
	// Apply specific group specs in the order they're listed
	for _, spec := range nodeSummaryGroupSpecs {
		if len(summaries[spec.topologyID]) == 0 {
			continue
		}
		apiTopology, ok := primaryAPITopology[spec.topologyID]
		if !ok {
			continue
		}
		sort.Sort(nodeSummariesByID(summaries[spec.topologyID]))
		group := spec.NodeSummaryGroup
		group.Nodes = summaries[spec.topologyID]
		group.TopologyID = apiTopology
		nodeSummaryGroups = append(nodeSummaryGroups, group)
		delete(summaries, spec.topologyID)
	}
	// As a fallback, in case a topology has no group spec defined, add any remaining at the end
	for topologyID, nodeSummaries := range summaries {
		if len(nodeSummaries) == 0 {
			continue
		}
		topology, ok := rc.Topology(topologyID)
		if !ok {
			continue
		}
		apiTopology, ok := primaryAPITopology[topologyID]
		if !ok {
			continue
		}
		sort.Sort(nodeSummariesByID(nodeSummaries))
		group := NodeSummaryGroup{
			TopologyID: apiTopology,
			Label:      topology.LabelPlural,
			Columns:    []Column{},
		}
		nodeSummaryGroups = append(nodeSummaryGroups, group)
	}

	return nodeSummaryGroups
}
