package report

import (
	"net"
	"strings"
)

// Report is the core data type. It's produced by probes, and consumed and
// stored by apps. It's composed of multiple topologies, each representing
// a different (related, but not equivalent) view of the network.
type Report struct {
	// Endpoint nodes are individual (address, port) tuples on each host.
	// They come from inspecting active connections and can (theoretically)
	// be traced back to a process. Edges are present.
	Endpoint Topology

	// Address nodes are addresses (e.g. ifconfig) on each host. Certain
	// information may be present in this topology that can't be mapped to
	// endpoints (e.g. ICMP). Edges are present.
	Address Topology

	// Process nodes are processes on each host. Edges are not present.
	Process Topology

	// Container nodes represent all Docker containers on hosts running probes.
	// Metadata includes things like Docker image, name etc.
	// Edges are not present.
	Container Topology

	// Host nodes are physical hosts that run probes. Metadata includes things
	// like operating system, load, etc. The information is scraped by the
	// probes with each published report. Edges are not present.
	Host Topology

	// Overlay nodes are active peers in any software-defined network that's
	// overlaid on the infrastructure. The information is scraped by polling
	// their status endpoints. Edges could be present, but aren't currently.
	Overlay Topology
}

const (
	// HostNodeID is a metadata foreign key, linking a node in any topology to
	// a node in the host topology. That host node is the origin host, where
	// the node was originally detected.
	HostNodeID = "host_node_id"
)

// RenderableNode is the data type that's yielded to the JavaScript layer as
// an element of a topology. It should contain information that's relevant
// to rendering a node when there are many nodes visible at once.
type RenderableNode struct {
	ID         string            `json:"id"`                    //
	LabelMajor string            `json:"label_major"`           // e.g. "process", human-readable
	LabelMinor string            `json:"label_minor,omitempty"` // e.g. "hostname", human-readable, optional
	Rank       string            `json:"rank"`                  // to help the layout engine
	Pseudo     bool              `json:"pseudo,omitempty"`      // sort-of a placeholder node, for rendering purposes
	Adjacency  IDList            `json:"adjacency,omitempty"`   // Node IDs (in the same topology domain)
	Origins    IDList            `json:"origins,omitempty"`     // Core node IDs that contributed information
	Metadata   AggregateMetadata `json:"metadata"`              // Numeric sums
}

// RenderableNodes is a set of RenderableNodes
type RenderableNodes map[string]RenderableNode

// DetailedNode is the data type that's yielded to the JavaScript layer when
// we want deep information about an individual node.
type DetailedNode struct {
	ID         string  `json:"id"`
	LabelMajor string  `json:"label_major"`
	LabelMinor string  `json:"label_minor,omitempty"`
	Pseudo     bool    `json:"pseudo,omitempty"`
	Tables     []Table `json:"tables"`
}

// Table is a dataset associated with a node. It will be displayed in the
// detail panel when a user clicks on a node.
type Table struct {
	Title   string `json:"title"`   // e.g. Bandwidth
	Numeric bool   `json:"numeric"` // should the major column be right-aligned?
	Rows    []Row  `json:"rows"`
}

// Row is a single entry in a Table dataset.
type Row struct {
	Key        string `json:"key"`                   // e.g. Ingress
	ValueMajor string `json:"value_major"`           // e.g. 25
	ValueMinor string `json:"value_minor,omitempty"` // e.g. KB/s
}

// MakeReport makes a clean report, ready to Merge() other reports into.
func MakeReport() Report {
	return Report{
		Endpoint:  NewTopology(),
		Address:   NewTopology(),
		Process:   NewTopology(),
		Container: NewTopology(),
		Host:      NewTopology(),
		Overlay:   NewTopology(),
	}
}

// Squash squashes all non-local nodes in the report to a super-node called
// the Internet.
func (r Report) Squash() Report {
	localNetworks := r.LocalNetworks()
	r.Endpoint = r.Endpoint.Squash(EndpointIDAddresser, localNetworks)
	r.Address = r.Address.Squash(AddressIDAddresser, localNetworks)
	r.Process = r.Process.Squash(PanicIDAddresser, localNetworks)
	r.Container = r.Container.Squash(PanicIDAddresser, localNetworks)
	r.Host = r.Host.Squash(PanicIDAddresser, localNetworks)
	r.Overlay = r.Overlay.Squash(PanicIDAddresser, localNetworks)
	return r
}

// LocalNetworks returns a superset of the networks (think: CIDRs) that are
// "local" from the perspective of each host represented in the report. It's
// used to determine which nodes in the report are "remote", i.e. outside of
// our infrastructure.
func (r Report) LocalNetworks() []*net.IPNet {
	var ipNets []*net.IPNet
	for _, md := range r.Host.NodeMetadatas {
		val, ok := md["local_networks"]
		if !ok {
			continue
		}
	outer:
		for _, s := range strings.Fields(val) {
			_, ipNet, err := net.ParseCIDR(s)
			if err != nil {
				continue
			}
			for _, existing := range ipNets {
				if ipNet.String() == existing.String() {
					continue outer
				}
			}
			ipNets = append(ipNets, ipNet)
		}
	}
	return ipNets
}

// Topologies returns a slice of Topologies in this report
func (r Report) Topologies() []Topology {
	return []Topology{r.Endpoint, r.Address, r.Process, r.Container, r.Host}
}

// Validate checks the report for various inconsistencies.
func (r Report) Validate() error {
	for _, topology := range r.Topologies() {
		if err := topology.Validate(); err != nil {
			return err
		}
	}
	return nil
}
