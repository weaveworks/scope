package report

import (
	"fmt"
	"net"
	"reflect"
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
// about a given node in a given topology. Right now it's a weakly-typed map,
// which should probably change (see comment on type MapFunc).
type NodeMetadata map[string]string

// Copy returns a value copy, useful for tests.
func (nm NodeMetadata) Copy() NodeMetadata {
	cp := make(NodeMetadata, len(nm))
	for k, v := range nm {
		cp[k] = v
	}
	return cp
}

// Merge merges two node metadata maps together. In case of conflict, the
// other (right-hand) side wins. Always reassign the result of merge to the
// destination. Merge is defined on the value-type, but node metadata map is
// itself a reference type, so if you want to maintain immutability, use copy.
func (nm NodeMetadata) Merge(other NodeMetadata) NodeMetadata {
	for k, v := range other {
		nm[k] = v // other takes precedence
	}
	return nm
}

// NewTopology gives you a Topology.
func NewTopology() Topology {
	return Topology{
		Adjacency:     map[string]IDList{},
		EdgeMetadatas: map[string]EdgeMetadata{},
		NodeMetadatas: map[string]NodeMetadata{},
	}
}

// Squash squashes all non-local nodes in the topology to a super-node called
// the Internet.
// We rely on the values in the t.Adjacency lists being valid keys in
// t.NodeMetadata (or t.Adjacency).
func (t Topology) Squash(f IDAddresser, localNets []*net.IPNet) Topology {
	isRemote := func(id string) bool {
		if _, ok := t.NodeMetadatas[id]; ok {
			return false // it is a node, cannot possibly be remote
		}

		if _, ok := t.Adjacency[MakeAdjacencyID(id)]; ok {
			return false // it is in our adjacency list, cannot possibly be remote
		}

		if ip := f(id); ip != nil && netsContain(localNets, ip) {
			return false // it is in our local nets, so it is not remote
		}

		return true
	}

	for srcID, dstIDs := range t.Adjacency {
		newDstIDs := make(IDList, 0, len(dstIDs))
		for _, dstID := range dstIDs {
			if isRemote(dstID) {
				dstID = TheInternet
			}
			newDstIDs = newDstIDs.Add(dstID)
		}
		t.Adjacency[srcID] = newDstIDs
	}
	return t
}

func netsContain(nets []*net.IPNet, ip net.IP) bool {
	for _, net := range nets {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

// Diff is returned by TopoDiff. It represents the changes between two
// RenderableNode maps.
type Diff struct {
	Add    []RenderableNode `json:"add"`
	Update []RenderableNode `json:"update"`
	Remove []string         `json:"remove"`
}

// TopoDiff gives you the diff to get from A to B.
func TopoDiff(a, b RenderableNodes) Diff {
	diff := Diff{}

	notSeen := map[string]struct{}{}
	for k := range a {
		notSeen[k] = struct{}{}
	}

	for k, node := range b {
		if _, ok := a[k]; !ok {
			diff.Add = append(diff.Add, node)
		} else if !reflect.DeepEqual(node, a[k]) {
			diff.Update = append(diff.Update, node)
		}
		delete(notSeen, k)
	}

	// leftover keys
	for k := range notSeen {
		diff.Remove = append(diff.Remove, k)
	}

	return diff
}

// ByID is a sort interface for a RenderableNode slice.
type ByID []RenderableNode

func (r ByID) Len() int           { return len(r) }
func (r ByID) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ByID) Less(i, j int) bool { return r[i].ID < r[j].ID }

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
