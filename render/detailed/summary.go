package detailed

import (
	"context"
	"fmt"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"

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

// BasicNodeSummary is basic summary information about a Node,
// sufficient for rendering links to the node.
type BasicNodeSummary struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	LabelMinor string `json:"labelMinor"`
	Rank       string `json:"rank"`
	Shape      string `json:"shape,omitempty"`
	Tag        string `json:"tag,omitempty"`
	Stack      bool   `json:"stack,omitempty"`
	Pseudo     bool   `json:"pseudo,omitempty"`
}

// NodeSummary is summary information about a Node.
type NodeSummary struct {
	BasicNodeSummary
	Metadata  []report.MetadataRow `json:"metadata,omitempty"`
	Parents   []Parent             `json:"parents,omitempty"`
	Metrics   []report.MetricRow   `json:"metrics,omitempty"`
	Tables    []report.Table       `json:"tables,omitempty"`
	Adjacency report.IDList        `json:"adjacency,omitempty"`
}

var renderers = map[string]func(BasicNodeSummary, report.Node) BasicNodeSummary{
	render.Pseudo:                pseudoNodeSummary,
	report.Process:               processNodeSummary,
	report.Container:             containerNodeSummary,
	report.ContainerImage:        containerImageNodeSummary,
	report.Pod:                   podNodeSummary,
	report.Service:               podGroupNodeSummary,
	report.Deployment:            podGroupNodeSummary,
	report.DaemonSet:             podGroupNodeSummary,
	report.StatefulSet:           podGroupNodeSummary,
	report.CronJob:               podGroupNodeSummary,
	report.Job:                   podGroupNodeSummary,
	report.ECSTask:               ecsTaskNodeSummary,
	report.ECSService:            ecsServiceNodeSummary,
	report.SwarmService:          swarmServiceNodeSummary,
	report.Host:                  hostNodeSummary,
	report.Overlay:               weaveNodeSummary,
	report.Endpoint:              nil, // Do not render
	report.PersistentVolume:      persistentVolumeNodeSummary,
	report.PersistentVolumeClaim: persistentVolumeClaimNodeSummary,
	report.StorageClass:          storageClassNodeSummary,
	report.VolumeSnapshot:        volumeSnapshotNodeSummary,
	report.VolumeSnapshotData:    volumeSnapshotDataNodeSummary,
}

// For each report.Topology, map to a 'primary' API topology. This can then be used in a variety of places.
var primaryAPITopology = map[string]string{
	report.Process:               "processes",
	report.Container:             "containers",
	report.ContainerImage:        "containers-by-image",
	report.Pod:                   "pods",
	report.Deployment:            "kube-controllers",
	report.DaemonSet:             "kube-controllers",
	report.StatefulSet:           "kube-controllers",
	report.CronJob:               "kube-controllers",
	report.Job:                   "kube-controllers",
	report.Service:               "services",
	report.ECSTask:               "ecs-tasks",
	report.ECSService:            "ecs-services",
	report.SwarmService:          "swarm-services",
	report.Host:                  "hosts",
	report.PersistentVolume:      "pods",
	report.PersistentVolumeClaim: "pods",
	report.StorageClass:          "pods",
	report.VolumeSnapshot:        "pods",
	report.VolumeSnapshotData:    "pods",
}

// MakeBasicNodeSummary returns a basic summary of a node, if
// possible. This summary is sufficient for rendering links to the node.
func MakeBasicNodeSummary(r report.Report, n report.Node) (BasicNodeSummary, bool) {
	summary := BasicNodeSummary{ // This is unlikely to look very good, but is a reasonable fallback
		ID:    n.ID,
		Label: n.ID,
		Shape: report.Triangle,
	}
	if t, ok := r.Topology(n.Topology); ok {
		summary.Shape = t.GetShape()
		summary.Tag = t.Tag
	}

	// Do we have a renderer for the topology?
	if renderer, ok := renderers[n.Topology]; ok {
		if renderer == nil { // we don't want to render this
			return summary, false
		}
		return renderer(summary, n), true
	}

	// Is it a group topology?
	if strings.HasPrefix(n.Topology, "group:") {
		return groupNodeSummary(summary, r, n), true
	}

	// Is it any known topology?
	if _, ok := r.Topology(n.Topology); ok {
		// We should never get here, since all known topologies are in
		// 'renderers'.
		return summary, true
	}

	// We have no idea how to render this.
	return summary, false
}

// MakeNodeSummary summarizes a node, if possible.
func MakeNodeSummary(rc RenderContext, n report.Node) (NodeSummary, bool) {
	base, ok := MakeBasicNodeSummary(rc.Report, n)
	if !ok {
		return NodeSummary{}, false
	}
	summary := NodeSummary{
		BasicNodeSummary: base,
		Parents:          Parents(rc.Report, n),
		Adjacency:        n.Adjacency,
	}
	// Only include metadata, metrics, tables when it's not a group node
	if n.CountChildrenOfTopology(n.Topology) == 0 {
		if topology, ok := rc.Topology(n.Topology); ok {
			summary.Metadata = topology.MetadataTemplates.MetadataRows(n)
			summary.Metrics = topology.MetricTemplates.MetricRows(n)
			summary.Tables = topology.TableTemplates.Tables(n)
		}
	}
	return RenderMetricURLs(summary, n, rc.Report, rc.MetricsGraphURL), true
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

func pseudoNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	pseudoID, _ := render.ParsePseudoNodeID(n.ID)
	base.Pseudo = true
	base.Rank = pseudoID

	switch {
	case n.ID == render.IncomingInternetID:
		// render as an internet node
		base.Label = render.InboundMajor
		base.LabelMinor = render.InboundMinor
		base.Shape = report.Cloud
	case n.ID == render.OutgoingInternetID:
		// render as an internet node
		base.Label = render.OutboundMajor
		base.LabelMinor = render.OutboundMinor
		base.Shape = report.Cloud
	case strings.HasPrefix(n.ID, render.ServiceNodeIDPrefix):
		// render as a known service node
		base.Label = n.ID[len(render.ServiceNodeIDPrefix):]
		base.LabelMinor = ""
		base.Shape = report.Cloud
	case strings.HasPrefix(n.ID, render.UncontainedIDPrefix):
		// render as an uncontained node
		base.Label = render.UncontainedMajor
		base.LabelMinor = n.ID[len(render.UncontainedIDPrefix):]
		base.Shape = report.Square
		base.Stack = true
	case strings.HasPrefix(n.ID, render.UnmanagedIDPrefix):
		// render as an unmanaged node
		base.Label = render.UnmanagedMajor
		base.LabelMinor = n.ID[len(render.UnmanagedIDPrefix):]
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
	return base
}

func processNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
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
	return base
}

func containerNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	var (
		containerName = getRenderableContainerName(n)
		hostName      = report.ExtractHostID(n)
		imageName, _  = n.Latest.Lookup(docker.ImageName)
	)
	base.Label = containerName
	base.LabelMinor = hostName
	if imageName != "" {
		base.Rank = docker.ImageNameWithoutTag(imageName)
	} else if hostName != "" {
		base.Rank = hostName
	} else {
		base.Rank = base.Label
	}
	return base
}

func containerImageNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	var (
		imageName, _        = n.Latest.Lookup(docker.ImageName)
		imageNameWithoutTag = docker.ImageNameWithoutTag(imageName)
	)
	switch {
	case imageNameWithoutTag != "" && imageNameWithoutTag != ImageNameNone:
		base.Label = imageNameWithoutTag
	case imageName != "" && imageName != ImageNameNone:
		base.Label = imageName
	default:
		// The id can be an image id or an image name. Ideally we'd
		// truncate the former but not the latter, but short of
		// heuristic regexp match we cannot tell the difference.
		base.Label, _ = report.ParseContainerImageNodeID(n.ID)
	}
	base.LabelMinor = pluralize(n, report.Container, "container", "containers")
	base.Rank = base.Label
	base.Stack = true
	return base
}

func addKubernetesLabelAndRank(base BasicNodeSummary, n report.Node) BasicNodeSummary {
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

func podNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	base = addKubernetesLabelAndRank(base, n)
	base.LabelMinor = pluralize(n, report.Container, "container", "containers")
	return base
}

var podGroupNodeTypeName = map[string]string{
	report.Deployment:  "Deployment",
	report.DaemonSet:   "DaemonSet",
	report.StatefulSet: "StatefulSet",
	report.CronJob:     "CronJob",
	report.Job:         "Job",
}

func podGroupNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	base = addKubernetesLabelAndRank(base, n)
	base.Stack = true
	// NB: pods are the highest aggregation level for which we display
	// counts.
	count := pluralize(n, report.Pod, "pod", "pods")
	if typeName, ok := podGroupNodeTypeName[n.Topology]; ok {
		base.LabelMinor = fmt.Sprintf("%s of %s", typeName, count)
	} else {
		base.LabelMinor = count
	}
	return base
}

func ecsTaskNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	base.Label, _ = n.Latest.Lookup(awsecs.TaskFamily)
	if base.Label == "" {
		base.Label, _ = report.ParseECSTaskNodeID(n.ID)
	}
	return base
}

func ecsServiceNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	_, base.Label, _ = report.ParseECSServiceNodeID(n.ID)
	base.Stack = true
	return base
}

func swarmServiceNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	base.Label, _ = n.Latest.Lookup(docker.ServiceName)
	if base.Label == "" {
		base.Label, _ = report.ParseSwarmServiceNodeID(n.ID)
	}
	return base
}

func hostNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	var (
		hostname, _ = report.ParseHostNodeID(n.ID)
		parts       = strings.SplitN(hostname, ".", 2)
	)
	if len(parts) == 2 {
		base.Label, base.LabelMinor, base.Rank = parts[0], parts[1], parts[1]
	} else {
		base.Label = hostname
	}
	return base
}

func weaveNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
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
	return base
}

func persistentVolumeNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	base = addKubernetesLabelAndRank(base, n)
	return base
}

func persistentVolumeClaimNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	base = addKubernetesLabelAndRank(base, n)
	return base
}

func storageClassNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	base = addKubernetesLabelAndRank(base, n)
	return base
}

func volumeSnapshotNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	base = addKubernetesLabelAndRank(base, n)
	return base
}

func volumeSnapshotDataNodeSummary(base BasicNodeSummary, n report.Node) BasicNodeSummary {
	base = addKubernetesLabelAndRank(base, n)
	return base
}

// groupNodeSummary renders the summary for a group node. n.Topology is
// expected to be of the form: group:container:hostname
func groupNodeSummary(base BasicNodeSummary, r report.Report, n report.Node) BasicNodeSummary {
	base.Label, base.Rank = n.ID, n.ID
	if topology, _, ok := render.ParseGroupNodeTopology(n.Topology); ok {
		if t, ok := r.Topology(topology); ok {
			base.Shape = t.GetShape()
			base.Tag = t.Tag
			if t.Label != "" {
				base.LabelMinor = pluralize(n, topology, t.Label, t.LabelPlural)
			}
		}
	}
	base.Stack = true
	return base
}

func pluralize(n report.Node, key, singular, plural string) string {
	// either fetch a stored counter, or count the children directly
	c, ok := n.LookupCounter(key)
	if !ok {
		c = n.CountChildrenOfTopology(key)
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
func Summaries(ctx context.Context, rc RenderContext, rns report.Nodes) NodeSummaries {
	span, ctx := opentracing.StartSpanFromContext(ctx, "detailed.Summaries")
	defer span.Finish()

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
