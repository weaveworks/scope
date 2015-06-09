package report

import (
	"encoding/json"
	"net"
	"time"
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

	HostMetadatas
}

// HostMetadatas contains metadata about the host(s) represented in the Report.
type HostMetadatas map[string]HostMetadata

// HostMetadata describes metadata that probes can collect about the host that
// they run on. It has a timestamp when the measurement was made.
type HostMetadata struct {
	Timestamp                      time.Time
	Hostname                       string
	LocalNets                      []*net.IPNet
	OS                             string
	LoadOne, LoadFive, LoadFifteen float64
}

// RenderableNode is the data type that's yielded to the JavaScript layer as
// an element of a topology. It should contain information that's relevant
// to rendering a node when there are many nodes visible at once.
type RenderableNode struct {
	ID          string            `json:"id"`                     //
	LabelMajor  string            `json:"label_major"`            // e.g. "process", human-readable
	LabelMinor  string            `json:"label_minor,omitempty"`  // e.g. "hostname", human-readable, optional
	Rank        string            `json:"rank"`                   // to help the layout engine
	Pseudo      bool              `json:"pseudo,omitempty"`       // sort-of a placeholder node, for rendering purposes
	Adjacency   IDList            `json:"adjacency,omitempty"`    // Node IDs (in the same topology domain)
	OriginHosts IDList            `json:"origin_hosts,omitempty"` // Which hosts contributed information to this node
	OriginNodes IDList            `json:"origin_nodes,omitempty"` // Which origin nodes (depends on topology) contributed
	Metadata    AggregateMetadata `json:"metadata"`               // Numeric sums
}

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
		Endpoint:      NewTopology(),
		Address:       NewTopology(),
		HostMetadatas: map[string]HostMetadata{},
	}
}

// SquashRemote folds all remote nodes into a special supernode. It uses the
// LocalNets of the hosts in HostMetadata to determine which addresses are
// local.
func (r Report) SquashRemote() Report {
	localNets := r.HostMetadatas.LocalNets()
	return Report{
		Endpoint:      Squash(r.Endpoint, EndpointIDAddresser, localNets),
		Address:       Squash(r.Address, AddressIDAddresser, localNets),
		HostMetadatas: r.HostMetadatas,
	}
}

// LocalNets gives the union of all local network IPNets for all hosts
// represented in the HostMetadatas.
func (m HostMetadatas) LocalNets() []*net.IPNet {
	var nets []*net.IPNet
	for _, node := range m {
	OUTER:
		for _, local := range node.LocalNets {
			for _, existing := range nets {
				if existing == local {
					continue OUTER
				}
			}
			nets = append(nets, local)
		}
	}
	return nets
}

// UnmarshalJSON is a custom JSON deserializer for HostMetadata to deal with
// the Localnets.
func (m *HostMetadata) UnmarshalJSON(data []byte) error {
	type netmask struct {
		IP   net.IP
		Mask []byte
	}
	tmpHMD := struct {
		Timestamp                      time.Time
		Hostname                       string
		LocalNets                      []*netmask
		OS                             string
		LoadOne, LoadFive, LoadFifteen float64
	}{}
	err := json.Unmarshal(data, &tmpHMD)
	if err != nil {
		return err
	}

	m.Timestamp = tmpHMD.Timestamp
	m.Hostname = tmpHMD.Hostname
	m.OS = tmpHMD.OS
	m.LoadOne = tmpHMD.LoadOne
	m.LoadFive = tmpHMD.LoadFive
	m.LoadFifteen = tmpHMD.LoadFifteen
	for _, ln := range tmpHMD.LocalNets {
		m.LocalNets = append(m.LocalNets, &net.IPNet{IP: ln.IP, Mask: ln.Mask})
	}
	return nil
}
