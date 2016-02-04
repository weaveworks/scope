package detailed

import (
	"fmt"

	"github.com/ugorji/go/codec"

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
type Column string

// CodecEncodeSelf implements codec.Selfer
func (c *Column) CodecEncodeSelf(encoder *codec.Encoder) {
	in := map[string]string{"id": string(*c), "label": Label(string(*c))}
	encoder.Encode(in)
}

// CodecDecodeSelf implements codec.Selfer
func (c *Column) CodecDecodeSelf(decoder *codec.Decoder) {
	m := map[string]string{}
	decoder.Decode(&m)
	*c = Column(m["id"])
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (Column) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*Column) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}

// NodeSummary is summary information about a child for a Node.
type NodeSummary struct {
	ID           string        `json:"id"`
	Label        string        `json:"label"`
	Linkable     bool          `json:"linkable"` // Whether this node can be linked-to
	Metadata     []MetadataRow `json:"metadata,omitempty"`
	DockerLabels []MetadataRow `json:"docker_labels,omitempty"`
	Metrics      []MetricRow   `json:"metrics,omitempty"`
}

// MakeNodeSummary summarizes a node, if possible.
func MakeNodeSummary(n report.Node) (NodeSummary, bool) {
	renderers := map[string]func(report.Node) NodeSummary{
		report.Process:        processNodeSummary,
		report.Container:      containerNodeSummary,
		report.ContainerImage: containerImageNodeSummary,
		report.Pod:            podNodeSummary,
		report.Host:           hostNodeSummary,
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
	for _, row := range n.DockerLabels {
		result.DockerLabels = append(result.DockerLabels, row.Copy())
	}
	for _, row := range n.Metrics {
		result.Metrics = append(result.Metrics, row.Copy())
	}
	return result
}

func baseNodeSummary(id, label string, linkable bool, nmd report.Node) NodeSummary {
	return NodeSummary{
		ID:           id,
		Label:        label,
		Linkable:     linkable,
		Metadata:     NodeMetadata(nmd),
		DockerLabels: NodeDockerLabels(nmd),
		Metrics:      NodeMetrics(nmd),
	}
}

func processNodeSummary(nmd report.Node) NodeSummary {
	var (
		id               string
		label, nameFound = nmd.Latest.Lookup(process.Name)
	)
	if pid, ok := nmd.Latest.Lookup(process.PID); ok {
		if !nameFound {
			label = fmt.Sprintf("(%s)", pid)
		}
		id = render.MakeProcessID(report.ExtractHostID(nmd), pid)
	}
	_, isConnected := nmd.Latest.Lookup(render.IsConnected)
	return baseNodeSummary(id, label, isConnected, nmd)
}

func containerNodeSummary(nmd report.Node) NodeSummary {
	label, _ := render.GetRenderableContainerName(nmd)
	containerID, _ := nmd.Latest.Lookup(docker.ContainerID)
	return baseNodeSummary(render.MakeContainerID(containerID), label, true, nmd)
}

func containerImageNodeSummary(nmd report.Node) NodeSummary {
	imageName, _ := nmd.Latest.Lookup(docker.ImageName)
	return baseNodeSummary(render.MakeContainerImageID(render.ImageNameWithoutVersion(imageName)), imageName, true, nmd)
}

func podNodeSummary(nmd report.Node) NodeSummary {
	podID, _ := nmd.Latest.Lookup(kubernetes.PodID)
	podName, _ := nmd.Latest.Lookup(kubernetes.PodName)
	return baseNodeSummary(render.MakePodID(podID), podName, true, nmd)
}

func hostNodeSummary(nmd report.Node) NodeSummary {
	hostName, _ := nmd.Latest.Lookup(host.HostName)
	return baseNodeSummary(render.MakeHostID(hostName), hostName, true, nmd)
}

type nodeSummariesByID []NodeSummary

func (s nodeSummariesByID) Len() int           { return len(s) }
func (s nodeSummariesByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s nodeSummariesByID) Less(i, j int) bool { return s[i].ID < s[j].ID }
