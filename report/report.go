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
	Endpoint       = "endpoint"
	Process        = "process"
	Container      = "container"
	Pod            = "pod"
	Service        = "service"
	ContainerImage = "container_image"
	Host           = "host"
	Overlay        = "overlay"

	// Used when counting the number of containers
	ContainersKey = "containers"
)

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

	// ContainerImages nodes represent all Docker containers images on
	// hosts running probes. Metadata includes things like image id, name etc.
	// Edges are not present.
	ContainerImage Topology

	// Host nodes are physical hosts that run probes. Metadata includes things
	// like operating system, load, etc. The information is scraped by the
	// probes with each published report. Edges are not present.
	Host Topology

	// Overlay nodes are active peers in any software-defined network that's
	// overlaid on the infrastructure. The information is scraped by polling
	// their status endpoints. Edges could be present, but aren't currently.
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

// MakeReport makes a clean report, ready to Merge() other reports into.
func MakeReport() Report {
	return Report{
		Endpoint:       MakeTopology(),
		Process:        MakeTopology(),
		Container:      MakeTopology(),
		ContainerImage: MakeTopology(),
		Host:           MakeTopology(),
		Pod:            MakeTopology(),
		Service:        MakeTopology(),
		Overlay:        MakeTopology(),
		Sampling:       Sampling{},
		Window:         0,
		Plugins:        xfer.MakePluginSpecs(),
		ID:             fmt.Sprintf("%d", rand.Int63()),
	}
}

// Copy returns a value copy of the report.
func (r Report) Copy() Report {
	return Report{
		Endpoint:       r.Endpoint.Copy(),
		Process:        r.Process.Copy(),
		Container:      r.Container.Copy(),
		ContainerImage: r.ContainerImage.Copy(),
		Host:           r.Host.Copy(),
		Pod:            r.Pod.Copy(),
		Service:        r.Service.Copy(),
		Overlay:        r.Overlay.Copy(),
		Sampling:       r.Sampling,
		Window:         r.Window,
		Plugins:        r.Plugins.Copy(),
		ID:             fmt.Sprintf("%d", rand.Int63()),
	}
}

// Merge merges another Report into the receiver and returns the result. The
// original is not modified.
func (r Report) Merge(other Report) Report {
	cp := r.Copy()
	cp.Endpoint = r.Endpoint.Merge(other.Endpoint)
	cp.Process = r.Process.Merge(other.Process)
	cp.Container = r.Container.Merge(other.Container)
	cp.ContainerImage = r.ContainerImage.Merge(other.ContainerImage)
	cp.Host = r.Host.Merge(other.Host)
	cp.Pod = r.Pod.Merge(other.Pod)
	cp.Service = r.Service.Merge(other.Service)
	cp.Overlay = r.Overlay.Merge(other.Overlay)
	cp.Sampling = r.Sampling.Merge(other.Sampling)
	cp.Window += other.Window
	cp.Plugins = r.Plugins.Merge(other.Plugins)
	return cp
}

// Topologies returns a slice of Topologies in this report
func (r Report) Topologies() []Topology {
	return []Topology{
		r.Endpoint,
		r.Process,
		r.Container,
		r.ContainerImage,
		r.Pod,
		r.Service,
		r.Host,
		r.Overlay,
	}
}

// Topology gets a topology by name
func (r Report) Topology(name string) (Topology, bool) {
	t, ok := map[string]Topology{
		Endpoint:       r.Endpoint,
		Process:        r.Process,
		Container:      r.Container,
		ContainerImage: r.ContainerImage,
		Pod:            r.Pod,
		Service:        r.Service,
		Host:           r.Host,
		Overlay:        r.Overlay,
	}[name]
	return t, ok
}

// Validate checks the report for various inconsistencies.
func (r Report) Validate() error {
	var errs []string
	for _, topology := range r.Topologies() {
		if err := topology.Validate(); err != nil {
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
