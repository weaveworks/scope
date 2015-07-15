package report

import (
	"fmt"
	"strings"
)

const localUnknown = "localUnknown"

// Topology describes a specific view of a network. It consists of nodes and
// edges, represented by Adjacency, and metadata about those nodes and edges,
// represented by EdgeMetadatas and NodeMetadatas respectively.
type Topology struct {
	Adjacency
	EdgeMetadatas
	NodeMetadatas
}

// Adjacency is an adjacency-list encoding of the topology. Keys are node IDs,
// as produced by the relevant MappingFunc for the topology.
type Adjacency map[string]IDList

// EdgeMetadatas collect metadata about each edge in a topology. Keys are a
// concatenation of node IDs.
type EdgeMetadatas map[string]EdgeMetadata

// NodeMetadatas collect metadata about each node in a topology. Keys are node
// IDs.
type NodeMetadatas map[string]NodeMetadata

// EdgeMetadata describes a superset of the metadata that probes can
// conceivably (and usefully) collect about an edge between two nodes in any
// topology.
type EdgeMetadata struct {
	WithBytes    bool `json:"with_bytes,omitempty"`
	BytesIngress uint `json:"bytes_ingress,omitempty"` // dst -> src
	BytesEgress  uint `json:"bytes_egress,omitempty"`  // src -> dst

	WithConnCountTCP bool `json:"with_conn_count_tcp,omitempty"`
	MaxConnCountTCP  uint `json:"max_conn_count_tcp,omitempty"`
}

// NodeMetadata describes a superset of the metadata that probes can collect
// about a given node in a given topology.
type NodeMetadata struct {
	Metadata
}

// Metadata is shorthand for a string:string map.
type Metadata map[string]string

// NewNodeMetadata creates a new NodeMetadata with the supplied Metadata.
func NewNodeMetadata(m Metadata) NodeMetadata {
	return NodeMetadata{
		Metadata: m,
	}
}

// Copy returns a value copy, useful for tests.
func (nm NodeMetadata) Copy() NodeMetadata {
	cp := NewNodeMetadata(Metadata{})
	for k, v := range nm.Metadata {
		cp.Metadata[k] = v
	}
	return cp
}

// NewTopology gives you a Topology.
func NewTopology() Topology {
	return Topology{
		Adjacency:     map[string]IDList{},
		EdgeMetadatas: map[string]EdgeMetadata{},
		NodeMetadatas: map[string]NodeMetadata{},
	}
}

// Validate checks the topology for various inconsistencies.
func (t Topology) Validate() error {
	// Check all edge metadata keys must have the appropriate entries in
	// adjacencies & node metadata.
	var errs []string
	for edgeID := range t.EdgeMetadatas {
		srcNodeID, dstNodeID, ok := ParseEdgeID(edgeID)
		if !ok {
			errs = append(errs, fmt.Sprintf("invalid edge ID %q", edgeID))
			continue
		}
		if _, ok := t.NodeMetadatas[srcNodeID]; !ok {
			errs = append(errs, fmt.Sprintf("node metadata missing for source node ID %q (from edge %q)", srcNodeID, edgeID))
			continue
		}
		dstNodeIDs, ok := t.Adjacency[MakeAdjacencyID(srcNodeID)]
		if !ok {
			errs = append(errs, fmt.Sprintf("adjacency entries missing for source node ID %q (from edge %q)", srcNodeID, edgeID))
			continue
		}
		if !dstNodeIDs.Contains(dstNodeID) {
			errs = append(errs, fmt.Sprintf("adjacency destination missing for destination node ID %q (from edge %q)", dstNodeID, edgeID))
			continue
		}
	}

	// Check all adjancency keys has entries in NodeMetadata.
	for adjacencyID := range t.Adjacency {
		nodeID, ok := ParseAdjacencyID(adjacencyID)
		if !ok {
			errs = append(errs, fmt.Sprintf("invalid adjacency ID %q", adjacencyID))
			continue
		}
		if _, ok := t.NodeMetadatas[nodeID]; !ok {
			errs = append(errs, fmt.Sprintf("node metadata missing for source node %q (from adjacency %q)", nodeID, adjacencyID))
			continue
		}
	}

	// Check all node metadata keys are parse-able (i.e. contain a scope)
	for nodeID := range t.NodeMetadatas {
		if _, _, ok := ParseNodeID(nodeID); !ok {
			errs = append(errs, fmt.Sprintf("invalid node ID %q", nodeID))
			continue
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "; "))
	}

	return nil
}
