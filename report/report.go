package report

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/weaveworks/scope/common/xfer"
)

// Names of the various topologies.
const (
	Endpoint              = "endpoint"
	Process               = "process"
	Container             = "container"
	Pod                   = "pod"
	Service               = "service"
	Deployment            = "deployment"
	ReplicaSet            = "replica_set"
	DaemonSet             = "daemon_set"
	StatefulSet           = "stateful_set"
	CronJob               = "cron_job"
	Namespace             = "namespace"
	ContainerImage        = "container_image"
	Host                  = "host"
	Overlay               = "overlay"
	ECSService            = "ecs_service"
	ECSTask               = "ecs_task"
	SwarmService          = "swarm_service"
	PersistentVolume      = "persistent_volume"
	PersistentVolumeClaim = "persistent_volume_claim"
	StorageClass          = "storage_class"
	VolumeSnapshot        = "volume_snapshot"
	VolumeSnapshotData    = "volume_snapshot_data"
	Job                   = "job"

	// Shapes used for different nodes
	Circle         = "circle"
	Triangle       = "triangle"
	Square         = "square"
	Pentagon       = "pentagon"
	Hexagon        = "hexagon"
	Heptagon       = "heptagon"
	Octagon        = "octagon"
	Cloud          = "cloud"
	Cylinder       = "cylinder"
	DottedCylinder = "dottedcylinder"
	StorageSheet   = "sheet"
	Camera         = "camera"
	DottedTriangle = "dottedtriangle"

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
	Namespace,
	Host,
	Overlay,
	ECSTask,
	ECSService,
	SwarmService,
	PersistentVolume,
	PersistentVolumeClaim,
	StorageClass,
	VolumeSnapshot,
	VolumeSnapshotData,
	Job,
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

	// Namespace nodes represent all Kubernetes Namespaces running on hosts running probes.
	// Metadata includes things like Namespace id, name, etc. Edges are not
	// present.
	Namespace Topology

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

	// Persistent Volume nodes represent all Kubernetes Persistent Volumes running on hosts running probes.
	// Metadata is limited for now, more to come later.
	PersistentVolume Topology

	// Persistent Volume Claim nodes represent all Kubernetes Persistent Volume Claims running on hosts running probes.
	// Metadata is limited for now, more to come later.
	PersistentVolumeClaim Topology

	// Storage Class represent all kubernetes Storage Classes on hosts running probes.
	// Metadata is limited for now, more to come later.
	StorageClass Topology

	// VolumeSnapshot represent all Kubernetes Volume Snapshots on hosts running probes.
	VolumeSnapshot Topology

	// VolumeSnapshotData represent all Kubernetes Volume Snapshot Data on hosts running probes.
	VolumeSnapshotData Topology

	// Job represent all Kubernetes Job on hosts running probes.
	Job Topology

	DNS DNSRecords `json:"DNS,omitempty" deepequal:"nil==empty"`
	// Backwards-compatibility for an accident in commit 951629a / release 1.11.6.
	BugDNS DNSRecords `json:"nodes,omitempty"`

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

		Namespace: MakeTopology(),

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

		PersistentVolume: MakeTopology().
			WithShape(Cylinder).
			WithLabel("persistent volume", "persistent volumes"),

		PersistentVolumeClaim: MakeTopology().
			WithShape(DottedCylinder).
			WithLabel("persistent volume claim", "persistent volume claims"),

		StorageClass: MakeTopology().
			WithShape(StorageSheet).
			WithLabel("storage class", "storage classes"),

		VolumeSnapshot: MakeTopology().
			WithShape(DottedCylinder).
			WithTag(Camera).
			WithLabel("volume snapshot", "volume snapshots"),

		VolumeSnapshotData: MakeTopology().
			WithShape(Cylinder).
			WithTag(Camera).
			WithLabel("volume snapshot data", "volume snapshot data"),

		Job: MakeTopology().
			WithShape(DottedTriangle).
			WithLabel("job", "jobs"),

		DNS: DNSRecords{},

		Sampling: Sampling{},
		Window:   0,
		Plugins:  xfer.MakePluginSpecs(),
		ID:       fmt.Sprintf("%d", rand.Int63()),
	}
}

// Copy returns a value copy of the report.
func (r Report) Copy() Report {
	newReport := Report{
		DNS:      r.DNS.Copy(),
		Sampling: r.Sampling,
		Window:   r.Window,
		Shortcut: r.Shortcut,
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
	newReport.UnsafeMerge(other)
	return newReport
}

// UnsafeMerge merges another Report into the receiver. The original is modified.
func (r *Report) UnsafeMerge(other Report) {
	r.DNS = r.DNS.Merge(other.DNS)
	r.Sampling = r.Sampling.Merge(other.Sampling)
	r.Window = r.Window + other.Window
	r.Plugins = r.Plugins.Merge(other.Plugins)
	r.WalkPairedTopologies(&other, func(ourTopology, theirTopology *Topology) {
		ourTopology.UnsafeMerge(*theirTopology)
	})
}

// UnsafeUnMerge removes any information from r that would be added by merging other.
// The original is modified.
func (r *Report) UnsafeUnMerge(other Report) {
	// TODO: DNS, Sampling, Plugins
	r.Window = r.Window - other.Window
	r.WalkPairedTopologies(&other, func(ourTopology, theirTopology *Topology) {
		ourTopology.UnsafeUnMerge(*theirTopology)
	})
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
	case Namespace:
		return &r.Namespace
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
	case PersistentVolume:
		return &r.PersistentVolume
	case PersistentVolumeClaim:
		return &r.PersistentVolumeClaim
	case StorageClass:
		return &r.StorageClass
	case VolumeSnapshot:
		return &r.VolumeSnapshot
	case VolumeSnapshotData:
		return &r.VolumeSnapshotData
	case Job:
		return &r.Job
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

// DropTopologiesOver - as a protection against overloading the app
// server, drop topologies that have really large node counts. In
// practice we only see this with runaway numbers of zombie processes.
func (r Report) DropTopologiesOver(limit int) Report {
	r.WalkNamedTopologies(func(name string, topology *Topology) {
		if topology != nil && len(topology.Nodes) > limit {
			topology.Nodes = Nodes{}
		}
	})
	return r
}

// Upgrade returns a new report based on a report received from the old probe.
//
func (r Report) Upgrade() Report {
	return r.upgradePodNodes().upgradeNamespaces().upgradeDNSRecords()
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

func (r Report) upgradeNamespaces() Report {
	if len(r.Namespace.Nodes) > 0 {
		return r
	}

	namespaces := map[string]struct{}{}
	for _, t := range []Topology{r.Pod, r.Service, r.Deployment, r.DaemonSet, r.StatefulSet, r.CronJob} {
		for _, n := range t.Nodes {
			if state, ok := n.Latest.Lookup(KubernetesState); ok && state == "deleted" {
				continue
			}
			if namespace, ok := n.Latest.Lookup(KubernetesNamespace); ok {
				namespaces[namespace] = struct{}{}
			}
		}
	}

	nodes := make(Nodes, len(namespaces))
	for ns := range namespaces {
		// Namespace ID:
		// Probes did not use to report namespace ids, but since creating a report node requires an id,
		// the namespace name, which is unique, is passed to `MakeNamespaceNodeID`
		namespaceID := MakeNamespaceNodeID(ns)
		nodes[namespaceID] = MakeNodeWith(namespaceID, map[string]string{KubernetesName: ns})
	}
	r.Namespace.Nodes = nodes

	return r
}

func (r Report) upgradeDNSRecords() Report {
	// For release 1.11.6, probes accidentally sent DNS records labeled "nodes".
	// Translate the incorrect version here. Accident was in commit 951629a.
	if len(r.BugDNS) > 0 {
		r.DNS = r.BugDNS
		r.BugDNS = nil
	}
	if len(r.DNS) > 0 {
		return r
	}
	dns := make(DNSRecords)
	for endpointID, endpoint := range r.Endpoint.Nodes {
		_, addr, _, ok := ParseEndpointNodeID(endpointID)
		snoopedNames, foundS := endpoint.Sets.Lookup(SnoopedDNSNames)
		reverseNames, foundR := endpoint.Sets.Lookup(ReverseDNSNames)
		if ok && (foundS || foundR) {
			// Add address and names to report-level map
			if existing, found := dns[addr]; found {
				var sUnchanged, rUnchanged bool
				snoopedNames, sUnchanged = snoopedNames.Merge(existing.Forward)
				reverseNames, rUnchanged = reverseNames.Merge(existing.Reverse)
				if sUnchanged && rUnchanged {
					continue
				}
			}
			dns[addr] = DNSRecord{Forward: snoopedNames, Reverse: reverseNames}
		}
	}
	r.DNS = dns
	return r
}

// Summary returns a human-readable string summarising the contents, for diagnostic purposes
func (r Report) Summary() string {
	ret := ""
	if len(r.Host.Nodes) == 1 {
		for k := range r.Host.Nodes {
			ret = k + ": "
		}
	}
	count := 0
	r.WalkNamedTopologies(func(n string, t *Topology) {
		if len(t.Nodes) > 0 {
			count++
			if count > 1 {
				ret = ret + ", "
			}
			ret = ret + fmt.Sprintf("%s:%d", n, len(t.Nodes))
		}
	})
	return ret
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
