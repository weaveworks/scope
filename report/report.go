package report

import (
	"fmt"
	"net"
	"strings"
)

// Report is the core data type. It's produced by probes, and consumed and
// stored by apps. It's composed of multiple topologies, each representing
// a different (related, but not equivalent) view of the network.
type Report struct {
	Endpoint Topology
	Address  Topology
	Process  Topology
	Host     Topology
}

// MakeReport produces a new report, ready for use. It's the only correct way
// to produce reports for general use, so please use it.
func MakeReport() Report {
	return Report{
		Endpoint: MakeTopology(),
		Address:  MakeTopology(),
		Process:  MakeTopology(),
		Host:     MakeTopology(),
	}
}

// Copy returns a value copy, useful for tests.
func (r Report) Copy() Report {
	return Report{
		Endpoint: r.Endpoint.Copy(),
		Address:  r.Address.Copy(),
		Process:  r.Process.Copy(),
		Host:     r.Host.Copy(),
	}
}

// Merge merges two reports together, returning the result. Always reassign
// the result of merge to the destination report. Merge is defined on report
// as a value-type, but report contains reference fields, so if you want to
// maintain immutable reports, use copy.
func (r Report) Merge(other Report) Report {
	r.Endpoint = r.Endpoint.Merge(other.Endpoint)
	r.Address = r.Address.Merge(other.Address)
	r.Process = r.Process.Merge(other.Process)
	r.Host = r.Host.Merge(other.Host)
	return r
}

// Squash squashes all non-local nodes in the report to a super-node called
// the Internet.
func (r Report) Squash() Report {
	localNetworks := r.LocalNetworks()
	r.Endpoint = r.Endpoint.Squash(EndpointIDAddresser, localNetworks)
	r.Address = r.Address.Squash(AddressIDAddresser, localNetworks)
	r.Process = r.Process.Squash(PanicIDAddresser, localNetworks)
	r.Host = r.Host.Squash(PanicIDAddresser, localNetworks)
	return r
}

// LocalNetworks returns a superset of the networks (think: CIDR) that are
// "local" from the perspective of each host represented in the report. It's
// used to determine which nodes in the report are "remote", i.e. outside of
// our domain of awareness.
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

// EdgeMetadata gives the metadata of an edge from the perspective of the
// srcMappedID. Since an edgeID can have multiple edges on the address level,
// it uses the supplied mapping function to translate core node IDs to
// renderable mapped IDs.
func (r Report) EdgeMetadata(ts TopologySelector, mapper MapFunc, srcMappedID, dstMappedID string) EdgeMetadata {
	t := ts(r)
	result := EdgeMetadata{}
	for edgeID, edgeMetadata := range t.EdgeMetadatas {
		srcNodeID, dstNodeID, ok := ParseEdgeID(edgeID)
		if !ok {
			panic(fmt.Sprintf("invalid edge ID %q", edgeID))
		}
		src, showSrc := mapper(r, ts, srcNodeID) // TODO srcNodeID == TheInternet checking?
		dst, showDst := mapper(r, ts, dstNodeID) // TODO dstNodeID == TheInternet checking?
		if showSrc && showDst && src.ID == srcMappedID && dst.ID == dstMappedID {
			result = result.Flatten(edgeMetadata)
		}
	}
	return result
}

// OriginTable produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func (r Report) OriginTable(originID string) (Table, bool) {
	for nodeID, nodeMetadata := range r.Endpoint.NodeMetadatas {
		if originID == nodeID {
			return endpointOriginTable(nodeMetadata)
		}
	}
	for nodeID, nodeMetadata := range r.Address.NodeMetadatas {
		if originID == nodeID {
			return addressOriginTable(nodeMetadata)
		}
	}
	for nodeID, nodeMetadata := range r.Process.NodeMetadatas {
		if originID == nodeID {
			return processOriginTable(nodeMetadata)
		}
	}
	for nodeID, nodeMetadata := range r.Host.NodeMetadatas {
		if originID == nodeID {
			return hostOriginTable(nodeMetadata)
		}
	}
	return Table{}, false
}

func endpointOriginTable(nmd NodeMetadata) (Table, bool) {
	rows := []Row{}
	if val, ok := nmd["endpoint"]; ok {
		rows = append(rows, Row{"Endpoint", val, ""})
	}
	if val, ok := nmd["host_name"]; ok {
		rows = append(rows, Row{"Host name", val, ""})
	}
	return Table{
		Title:   "Origin Endpoint",
		Numeric: false,
		Rows:    rows,
	}, len(rows) > 0
}

func addressOriginTable(nmd NodeMetadata) (Table, bool) {
	rows := []Row{}
	if val, ok := nmd["address"]; ok {
		rows = append(rows, Row{"Address", val, ""})
	}
	if val, ok := nmd["host_name"]; ok {
		rows = append(rows, Row{"Host name", val, ""})
	}
	return Table{
		Title:   "Origin Address",
		Numeric: false,
		Rows:    rows,
	}, len(rows) > 0
}

func processOriginTable(nmd NodeMetadata) (Table, bool) {
	rows := []Row{}
	if val, ok := nmd["process_name"]; ok {
		rows = append(rows, Row{"Process name", val, ""})
	}
	if val, ok := nmd["pid"]; ok {
		rows = append(rows, Row{"PID", val, ""})
	}
	if val, ok := nmd["docker_id"]; ok {
		rows = append(rows, Row{"Docker container ID", val, ""})
	}
	if val, ok := nmd["docker_name"]; ok {
		rows = append(rows, Row{"Docker container name", val, ""})
	}
	if val, ok := nmd["docker_image_id"]; ok {
		rows = append(rows, Row{"Docker image ID", val, ""})
	}
	if val, ok := nmd["docker_image_name"]; ok {
		rows = append(rows, Row{"Docker image name", val, ""})
	}
	return Table{
		Title:   "Origin Process",
		Numeric: false,
		Rows:    rows,
	}, len(rows) > 0
}

func hostOriginTable(nmd NodeMetadata) (Table, bool) {
	rows := []Row{}
	if val, ok := nmd["host_name"]; ok {
		rows = append(rows, Row{"Host name", val, ""})
	}
	if val, ok := nmd["load"]; ok {
		rows = append(rows, Row{"Load", val, ""})
	}
	if val, ok := nmd["os"]; ok {
		rows = append(rows, Row{"Operating system", val, ""})
	}
	return Table{
		Title:   "Origin Host",
		Numeric: false,
		Rows:    rows,
	}, len(rows) > 0
}
