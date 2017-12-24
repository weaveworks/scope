package report

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/common/xfer"
)

// Names of the various topologies.
const (
	Endpoint       = "endpoint"
	Process        = "process"
	Container      = "container"
	Pod            = "pod"
	Service        = "service"
	Deployment     = "deployment"
	ReplicaSet     = "replica_set"
	DaemonSet      = "daemon_set"
	StatefulSet    = "stateful_set"
	CronJob        = "cron_job"
	ContainerImage = "container_image"
	Host           = "host"
	Overlay        = "overlay"
	ECSService     = "ecs_service"
	ECSTask        = "ecs_task"
	SwarmService   = "swarm_service"

	// Shapes used for different nodes
	Circle   = "circle"
	Triangle = "triangle"
	Square   = "square"
	Pentagon = "pentagon"
	Hexagon  = "hexagon"
	Heptagon = "heptagon"
	Octagon  = "octagon"
	Cloud    = "cloud"

	// Used when counting the number of containers
	ContainersKey = "containers"
)

// topologyNames are the names of all report topologies.
var topologyNames = []string{
	Endpoint,
	Process,
	Container,
	ContainerImage,
	Pod,
	Service,
	Deployment,
	ReplicaSet,
	DaemonSet,
	StatefulSet,
	CronJob,
	Host,
	Overlay,
	ECSTask,
	ECSService,
	SwarmService,
}

// Report is the core data type. It's produced by probes, and consumed and
// stored by apps. It's composed of multiple topologies, each representing
// a different (related, but not equivalent) view of the network.
type Report struct {
	// Endpoint nodes are individual (address, port) tuples on each host.
	// They come from inspecting active connections and can (theoretically)
	// be traced back to a process. Edges are present.
	Endpoint Topology

	// Process nodes are processes on each host. Edges are not present.
	Process Topology

	// Container nodes represent all Docker containers on hosts running probes.
	// Metadata includes things like containter id, name, image id etc.
	// Edges are not present.
	Container Topology

	// Pod nodes represent all Kubernetes pods running on hosts running probes.
	// Metadata includes things like pod id, name etc. Edges are not
	// present.
	Pod Topology

	// Service nodes represent all Kubernetes services running on hosts running probes.
	// Metadata includes things like service id, name etc. Edges are not
	// present.
	Service Topology

	// Deployment nodes represent all Kubernetes deployments running on hosts running probes.
	// Metadata includes things like deployment id, name etc. Edges are not
	// present.
	Deployment Topology

	// ReplicaSet nodes represent all Kubernetes ReplicaSets running on hosts running probes.
	// Metadata includes things like ReplicaSet id, name etc. Edges are not
	// present.
	ReplicaSet Topology

	// DaemonSet nodes represent all Kubernetes DaemonSets running on hosts running probes.
	// Metadata includes things like DaemonSet id, name etc. Edges are not
	// present.
	DaemonSet Topology

	// StatefulSet nodes represent all Kubernetes Stateful Sets running on hosts running probes.
	// Metadata includes things like Stateful Set id, name, etc. Edges are not
	// present.
	StatefulSet Topology

	// CronJob nodes represent all Kubernetes Cron Jobs running on hosts running probes.
	// Metadata includes things like Cron Job id, name, etc. Edges are not
	// present.
	CronJob Topology

	// ContainerImages nodes represent all Docker containers images on
	// hosts running probes. Metadata includes things like image id, name etc.
	// Edges are not present.
	ContainerImage Topology

	// Host nodes are physical hosts that run probes. Metadata includes things
	// like operating system, load, etc. The information is scraped by the
	// probes with each published report. Edges are not present.
	Host Topology

	// ECS Task nodes are AWS ECS tasks, which represent a group of containers.
	// Metadata is limited for now, more to come later. Edges are not present.
	ECSTask Topology

	// ECS Service nodes are AWS ECS services, which represent a specification for a
	// desired count of tasks with a task definition template.
	// Metadata is limited for now, more to come later. Edges are not present.
	ECSService Topology

	// Swarm Service nodes are Docker Swarm services, which represent a specification for a
	// group of tasks (either one per host, or a desired count).
	// Edges are not present.
	SwarmService Topology

	// Overlay nodes are active peers in any software-defined network that's
	// overlaid on the infrastructure. The information is scraped by polling
	// their status endpoints. Edges are present.
	Overlay Topology

	// Sampling data for this report.
	Sampling Sampling

	// Window is the amount of time that this report purports to represent.
	// Windows must be carefully merged. They should only be added when
	// reports cover non-overlapping periods of time. By default, we assume
	// that's true, and add windows in merge operations. When that's not true,
	// such as in the app, we expect the component to overwrite the window
	// before serving it to consumers.
	Window time.Duration

	// Shortcut reports should be propagated to the UI as quickly as possible,
	// bypassing the usual spy interval, publish interval and app ws interval.
	Shortcut bool

	Plugins xfer.PluginSpecs

	// ID a random identifier for this report, used when caching
	// rendered views of the report.  Reports with the same id
	// must be equal, but we don't require that equal reports have
	// the same id.
	ID string `deepequal:"skip"`
}

// RenderContext carries contextual data that is needed when rendering parts of the report.
type RenderContext struct {
	Report
	MetricsGraphURL string
}

// MakeReport makes a clean report, ready to Merge() other reports into.
func MakeReport() Report {
	return Report{
		Endpoint: MakeTopology(),

		Process: MakeTopology().
			WithShape(Square).
			WithLabel("process", "processes"),

		Container: MakeTopology().
			WithShape(Hexagon).
			WithLabel("container", "containers"),

		ContainerImage: MakeTopology().
			WithShape(Hexagon).
			WithLabel("image", "images"),

		Host: MakeTopology().
			WithShape(Circle).
			WithLabel("host", "hosts"),

		Pod: MakeTopology().
			WithShape(Heptagon).
			WithLabel("pod", "pods"),

		Service: MakeTopology().
			WithShape(Heptagon).
			WithLabel("service", "services"),

		Deployment: MakeTopology().
			WithShape(Heptagon).
			WithLabel("deployment", "deployments"),

		ReplicaSet: MakeTopology().
			WithShape(Triangle).
			WithLabel("replica set", "replica sets"),

		DaemonSet: MakeTopology().
			WithShape(Pentagon).
			WithLabel("daemonset", "daemonsets"),

		StatefulSet: MakeTopology().
			WithShape(Octagon).
			WithLabel("stateful set", "stateful sets"),

		CronJob: MakeTopology().
			WithShape(Triangle).
			WithLabel("cron job", "cron jobs"),

		Overlay: MakeTopology().
			WithShape(Circle).
			WithLabel("peer", "peers"),

		ECSTask: MakeTopology().
			WithShape(Heptagon).
			WithLabel("task", "tasks"),

		ECSService: MakeTopology().
			WithShape(Heptagon).
			WithLabel("service", "services"),

		SwarmService: MakeTopology().
			WithShape(Heptagon).
			WithLabel("service", "services"),

		Sampling: Sampling{},
		Window:   0,
		Plugins:  xfer.MakePluginSpecs(),
		ID:       fmt.Sprintf("%d", rand.Int63()),
	}
}

// Copy returns a value copy of the report.
func (r Report) Copy() Report {
	newReport := Report{
		Sampling: r.Sampling,
		Window:   r.Window,
		Plugins:  r.Plugins.Copy(),
		ID:       fmt.Sprintf("%d", rand.Int63()),
	}
	newReport.WalkPairedTopologies(&r, func(newTopology, oldTopology *Topology) {
		*newTopology = oldTopology.Copy()
	})
	return newReport
}

// Merge merges another Report into the receiver and returns the result. The
// original is not modified.
func (r Report) Merge(other Report) Report {
	newReport := r.Copy()
	newReport.Sampling = newReport.Sampling.Merge(other.Sampling)
	newReport.Window = newReport.Window + other.Window
	newReport.Plugins = newReport.Plugins.Merge(other.Plugins)
	newReport.WalkPairedTopologies(&other, func(ourTopology, theirTopology *Topology) {
		*ourTopology = ourTopology.Merge(*theirTopology)
	})
	return newReport
}

// WalkTopologies iterates through the Topologies of the report,
// potentially modifying them
func (r *Report) WalkTopologies(f func(*Topology)) {
	for _, name := range topologyNames {
		f(r.topology(name))
	}
}

// WalkNamedTopologies iterates through the Topologies of the report,
// potentially modifying them.
func (r *Report) WalkNamedTopologies(f func(string, *Topology)) {
	for _, name := range topologyNames {
		f(name, r.topology(name))
	}
}

// WalkPairedTopologies iterates through the Topologies of this and another report,
// potentially modifying one or both.
func (r *Report) WalkPairedTopologies(o *Report, f func(*Topology, *Topology)) {
	for _, name := range topologyNames {
		f(r.topology(name), o.topology(name))
	}
}

// topology returns a reference to one of the report's topologies,
// selected by name.
func (r *Report) topology(name string) *Topology {
	switch name {
	case Endpoint:
		return &r.Endpoint
	case Process:
		return &r.Process
	case Container:
		return &r.Container
	case ContainerImage:
		return &r.ContainerImage
	case Pod:
		return &r.Pod
	case Service:
		return &r.Service
	case Deployment:
		return &r.Deployment
	case ReplicaSet:
		return &r.ReplicaSet
	case DaemonSet:
		return &r.DaemonSet
	case StatefulSet:
		return &r.StatefulSet
	case CronJob:
		return &r.CronJob
	case Host:
		return &r.Host
	case Overlay:
		return &r.Overlay
	case ECSTask:
		return &r.ECSTask
	case ECSService:
		return &r.ECSService
	case SwarmService:
		return &r.SwarmService
	}
	return nil
}

// Topology returns one of the report's topologies, selected by name.
func (r Report) Topology(name string) (Topology, bool) {
	if t := r.topology(name); t != nil {
		return *t, true
	}
	return Topology{}, false
}

// Validate checks the report for various inconsistencies.
func (r Report) Validate() error {
	var errs []string
	for _, name := range topologyNames {
		if err := r.topology(name).Validate(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if r.Sampling.Count > r.Sampling.Total {
		errs = append(errs, fmt.Sprintf("sampling count (%d) bigger than total (%d)", r.Sampling.Count, r.Sampling.Total))
	}
	if len(errs) > 0 {
		return fmt.Errorf("%d error(s): %s", len(errs), strings.Join(errs, "; "))
	}
	return nil
}

// Upgrade returns a new report based on a report received from the old probe.
//
// This for now creates node's LatestControls from Controls.
func (r Report) Upgrade() Report {
	return r.upgradeLatestControls().upgradePodNodes()
}

func (r Report) upgradeLatestControls() Report {
	needUpgrade := false
	r.WalkTopologies(func(topology *Topology) {
		for _, node := range topology.Nodes {
			if node.LatestControls.Size() == 0 && len(node.Controls.Controls) > 0 {
				needUpgrade = true
			}
		}
	})

	if !needUpgrade {
		return r
	}

	cp := r.Copy()
	ncd := NodeControlData{
		Dead: false,
	}
	cp.WalkTopologies(func(topology *Topology) {
		n := Nodes{}
		for name, node := range topology.Nodes {
			if node.LatestControls.Size() == 0 && len(node.Controls.Controls) > 0 {
				for _, control := range node.Controls.Controls {
					node.LatestControls = node.LatestControls.Set(control, node.Controls.Timestamp, ncd)
				}
			}
			n[name] = node
		}
		topology.Nodes = n
	})

	return cp
}

func (r Report) upgradePodNodes() Report {
	// At the same time the probe stopped reporting replicasets,
	// it also started reporting deployments as pods' parents
	if len(r.ReplicaSet.Nodes) == 0 {
		return r
	}

	// For each pod, we check for any replica sets, and merge any deployments they point to
	// into a replacement Parents value.
	nodes := Nodes{}
	for podID, pod := range r.Pod.Nodes {
		if replicaSetIDs, ok := pod.Parents.Lookup(ReplicaSet); ok {
			newParents := pod.Parents.Delete(ReplicaSet)
			for _, replicaSetID := range replicaSetIDs {
				if replicaSet, ok := r.ReplicaSet.Nodes[replicaSetID]; ok {
					if deploymentIDs, ok := replicaSet.Parents.Lookup(Deployment); ok {
						newParents = newParents.Add(Deployment, deploymentIDs)
					}
				}
			}
			// newParents contains a copy of the current parents without replicasets,
			// PruneParents().WithParents() ensures replicasets are actually deleted
			pod = pod.PruneParents().WithParents(newParents)
		}
		nodes[podID] = pod
	}
	r.Pod.Nodes = nodes

	return r
}

// BackwardCompatible returns a new backward-compatible report.
//
// This for now creates node's Controls from LatestControls.
func (r Report) BackwardCompatible() Report {
	now := mtime.Now()
	cp := r.Copy()
	cp.WalkTopologies(func(topology *Topology) {
		n := Nodes{}
		for name, node := range topology.Nodes {
			var controls []string
			node.LatestControls.ForEach(func(k string, _ time.Time, v NodeControlData) {
				if !v.Dead {
					controls = append(controls, k)
				}
			})
			if len(controls) > 0 {
				node.Controls = NodeControls{
					Timestamp: now,
					Controls:  MakeStringSet(controls...),
				}
			}
			n[name] = node
		}
		topology.Nodes = n
	})
	return cp
}

// Sampling describes how the packet data sources for this report were
// sampled. It can be used to calculate effective sample rates. We can't
// just put the rate here, because that can't be accurately merged. Counts
// in e.g. edge metadata structures have already been adjusted to
// compensate for the sample rate.
type Sampling struct {
	Count uint64 // observed and processed
	Total uint64 // observed overall
}

// Rate returns the effective sampling rate.
func (s Sampling) Rate() float64 {
	if s.Total <= 0 {
		return 1.0
	}
	return float64(s.Count) / float64(s.Total)
}

// Merge combines two sampling structures via simple addition and returns the
// result. The original is not modified.
func (s Sampling) Merge(other Sampling) Sampling {
	return Sampling{
		Count: s.Count + other.Count,
		Total: s.Total + other.Total,
	}
}

const (
	// HostNodeID is a metadata foreign key, linking a node in any topology to
	// a node in the host topology. That host node is the origin host, where
	// the node was originally detected.
	HostNodeID = "host_node_id"
	// ControlProbeID is the random ID of the probe which controls the specific node.
	ControlProbeID = "control_probe_id"
)
