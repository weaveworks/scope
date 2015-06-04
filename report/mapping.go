package report

import (
	"fmt"
	"net"
	"strings"
)

// MapFunc deterministically maps nodes in a report to some other domain.
type MapFunc func(r Report, ts TopologySelector, nodeID string) (MappedNode, bool)

// PseudoFunc creates MappedNodes representing pseudo nodes given the node ID.
// The node ID is prior to mapping.
type PseudoFunc func(nodeID string) MappedNode

// TopologySelector chooses a topology from a report.
type TopologySelector func(Report) Topology

// SelectEndpoint selects the endpoint topology.
func SelectEndpoint(r Report) Topology { return r.Endpoint }

// SelectAddress selects the address topology.
func SelectAddress(r Report) Topology { return r.Address }

// SelectProcess selects the process topology.
func SelectProcess(r Report) Topology { return r.Process }

// SelectHost selects the host topology.
func SelectHost(r Report) Topology { return r.Host }

// MappedNode is an intermediate form, produced by MapFunc and PseudoFunc. It
// represents a node from a report, after that node's been passed through a
// mapping transformation. Multiple report nodes may map to the same mapped
// node.
type MappedNode struct {
	ID    string
	Major string
	Minor string
	Rank  string
}

// ProcessPID is a MapFunc that maps all nodes to their origin process PID.
// That is, all nodes with the same process PID (on the same host) will be
// mapped together. Nodes without processes are not mapped.
func ProcessPID(r Report, ts TopologySelector, nodeID string) (MappedNode, bool) {
	md, ok := ts(r).NodeMetadatas[nodeID]
	if !ok {
		return MappedNode{}, false // programmer error
	}
	processNodeID, ok := md["process_node_id"]
	if !ok {
		return MappedNode{}, false // process not available
	}
	md, ok = r.Process.NodeMetadatas[processNodeID]
	if !ok {
		return MappedNode{}, false // programmer error
	}
	var (
		processName = md.GetDefault("process_name", "(unknown)")
		hostName    = md.GetDefault("host_name", "(unknown)")
		processPID  = md.GetDefault("pid", "?")
	)
	return MappedNode{
		ID:    processNodeID,
		Major: processName,
		Minor: fmt.Sprintf("%s (%s)", hostName, processPID),
		Rank:  hostName,
	}, true
}

// ProcessName is a MapFunc that maps all nodes to their origin process name.
// That is, all nodes with the same process name (independent of host) will be
// mapped together. Nodes without processes are not mapped.
func ProcessName(r Report, ts TopologySelector, nodeID string) (MappedNode, bool) {
	md, ok := ts(r).NodeMetadatas[nodeID]
	if !ok {
		return MappedNode{}, false // programmer error
	}
	processNodeID, ok := md["process_node_id"]
	if !ok {
		return MappedNode{}, false // process not available
	}
	md, ok = r.Process.NodeMetadatas[processNodeID]
	if !ok {
		return MappedNode{}, false // programmer error
	}
	processName, ok := md["process_name"]
	if !ok {
		return MappedNode{}, false
	}
	return MappedNode{
		ID:    processName,
		Major: processName,
		Minor: md.GetDefault("host_name", "(unknown)"),
		Rank:  processName,
	}, true
}

// ProcessContainer is a MapFunc that maps all nodes to their origin container
// ID. That is, all nodes running in the same Docker container ID (independent
// of host) will be mapped together. Nodes without containers are not mapped.
func ProcessContainer(r Report, ts TopologySelector, nodeID string) (MappedNode, bool) {
	md, ok := ts(r).NodeMetadatas[nodeID]
	if !ok {
		return MappedNode{}, false // programmer error
	}
	processNodeID, ok := md["process_node_id"]
	if !ok {
		return MappedNode{}, false // process not available
	}
	md, ok = r.Process.NodeMetadatas[processNodeID]
	if !ok {
		return MappedNode{}, false // programmer error
	}
	dockerContainerID, ok := md["docker_container_id"]
	if !ok {
		return MappedNode{
			ID:    "uncontained",
			Major: "Uncontained",
			Minor: "",
			Rank:  "uncontained",
		}, true
	}
	var (
		dockerContainerName = md.GetDefault("docker_container_name", "(unknown)")
		hostName            = md.GetDefault("host_name", "(unknown)")
		dockerImageID       = md.GetDefault("docker_image_id", "unknown")
	)
	return MappedNode{
		ID:    dockerContainerID,
		Major: dockerContainerName,
		Minor: hostName,
		Rank:  dockerImageID,
	}, true
}

// ProcessContainerImage is a MapFunc that maps all nodes to their origin
// container image ID. That is, all nodes running from the same Docker image
// ID (independent of host) will be mapped together. Nodes without containers
// are not mapped.
func ProcessContainerImage(r Report, ts TopologySelector, nodeID string) (MappedNode, bool) {
	md, ok := ts(r).NodeMetadatas[nodeID]
	if !ok {
		return MappedNode{}, false // programmer error
	}
	processNodeID, ok := md["process_node_id"]
	if !ok {
		return MappedNode{}, false // process not available
	}
	md, ok = r.Process.NodeMetadatas[processNodeID]
	if !ok {
		return MappedNode{}, false // programmer error
	}
	if _, ok := md["docker_container_id"]; !ok {
		return MappedNode{
			ID:    "uncontained",
			Major: "Uncontained",
			Minor: "",
			Rank:  "uncontained",
		}, true
	}
	var (
		dockerImageID   = md.GetDefault("docker_image_id", "unknown")
		dockerImageName = md.GetDefault("docker_image_name", "unknown")
	)
	return MappedNode{
		ID:    dockerImageID,
		Major: dockerImageName,
		Minor: "",
		Rank:  dockerImageID,
	}, true
}

// AddressHostname is a MapFunc that maps all nodes to their origin host
// (hostname). That is, all nodes pulled from the same host will be mapped
// together. Nodes without information about origin host (via the address
// topology) are not mapped.
func AddressHostname(r Report, ts TopologySelector, nodeID string) (MappedNode, bool) {
	md, ok := ts(r).NodeMetadatas[nodeID]
	if !ok {
		return MappedNode{}, false // programmer error
	}
	addressNodeID, ok := md["address_node_id"]
	if !ok {
		return MappedNode{}, false // process not available
	}
	md, ok = r.Address.NodeMetadatas[addressNodeID]
	if !ok {
		return MappedNode{}, false // programmer error
	}
	hostName := md.GetDefault("host_name", "(unknown)")
	major, minor := hostName, ""
	if fields := strings.SplitN(hostName, ".", 2); len(fields) == 2 {
		major, minor = fields[0], fields[1]
	}
	return MappedNode{
		ID:    addressNodeID,
		Major: major,
		Minor: minor,
		Rank:  addressNodeID,
	}, true
}

// BasicPseudoNode is a PseudoFunc that grants each node ID its own pseudo
// node. It's effectively an identity, or one-to-one, mapping.
func BasicPseudoNode(nodeID string) MappedNode {
	if nodeID == TheInternet {
		return MappedNode{
			ID:    TheInternet,
			Major: formatLabel(TheInternet),
			Minor: "",
			Rank:  TheInternet,
		}
	}
	return MappedNode{
		ID:    MakePseudoNodeID(nodeID),
		Major: formatLabel(nodeID),
		Minor: "",
		Rank:  MakePseudoNodeID(nodeID),
	}
}

// GroupedPseudoNode is a PseudoFunc that maps every node ID to the same
// pseudo node. It's effectively a many-to-one mapping.
func GroupedPseudoNode(nodeID string) MappedNode {
	if nodeID == TheInternet {
		return MappedNode{
			ID:    TheInternet,
			Major: formatLabel(TheInternet),
			Minor: "",
			Rank:  TheInternet,
		}
	}
	return MappedNode{
		ID:    MakePseudoNodeID("unknown"),
		Major: "Unknown",
		Minor: "",
		Rank:  MakePseudoNodeID("unknown"),
	}
}

// NoPseudoNode is (effectively) a PseudoFunc that suppresses the generation
// of all pseudo nodes.
var NoPseudoNode PseudoFunc // = nil

func formatLabel(nodeID string) string {
	if nodeID == TheInternet {
		return "the Internet"
	}
	switch fields := strings.SplitN(nodeID, ScopeDelim, 3); {
	case len(fields) < 2:
		return nodeID
	case len(fields) == 2:
		return fields[1]
	default:
		return net.JoinHostPort(fields[1], fields[2])
	}
}
