package report

import (
	"fmt"
	"log"
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

// RenderBy transforms a given Topology into a set of RenderableNodes, which
// the UI will render collectively as a graph. Note that a RenderableNode will
// always be rendered with other nodes, and therefore contains limited detail.
//
// RenderBy takes a a MapFunc, which defines how to group and label nodes. Npdes
// with the same mapped IDs will be merged.
func (t Topology) RenderBy(mapFunc MapFunc, pseudoFunc PseudoFunc) RenderableNodes {
	nodes := RenderableNodes{}

	// Build a set of RenderableNodes for all non-pseudo probes, and an
	// addressID to nodeID lookup map. Multiple addressIDs can map to the same
	// RenderableNodes.
	var (
		source2mapped = map[string]string{} // source node ID -> mapped node ID
		source2host   = map[string]string{} // source node ID -> origin host ID
	)
	for nodeID, metadata := range t.NodeMetadatas {
		mapped, ok := mapFunc(metadata)
		if !ok {
			continue
		}

		// mapped.ID needs not be unique over all addressIDs. If not, we merge with
		// the existing data, on the assumption that the MapFunc returns the same
		// data.
		existing, ok := nodes[mapped.ID]
		if ok {
			mapped.Merge(existing)
		}

		mapped.Origins = mapped.Origins.Add(nodeID)
		nodes[mapped.ID] = mapped
		source2mapped[nodeID] = mapped.ID
		source2host[nodeID] = metadata[HostNodeID]
	}

	// Walk the graph and make connections.
	for src, dsts := range t.Adjacency {
		var (
			srcNodeID, ok = ParseAdjacencyID(src)
			//srcOriginHostID, _, ok2 = ParseNodeID(srcNodeID)
			srcHostNodeID     = source2host[srcNodeID]
			srcRenderableID   = source2mapped[srcNodeID] // must exist
			srcRenderableNode = nodes[srcRenderableID]   // must exist
		)
		if !ok {
			log.Printf("bad adjacency ID %q", src)
			continue
		}

		for _, dstNodeID := range dsts {
			dstRenderableID, ok := source2mapped[dstNodeID]
			if !ok {
				pseudoNode, ok := pseudoFunc(srcNodeID, srcRenderableNode, dstNodeID)
				if !ok {
					continue
				}
				dstRenderableID = pseudoNode.ID
				nodes[dstRenderableID] = pseudoNode
				source2mapped[dstNodeID] = dstRenderableID
			}

			srcRenderableNode.Adjacency = srcRenderableNode.Adjacency.Add(dstRenderableID)
			srcRenderableNode.Origins = srcRenderableNode.Origins.Add(srcHostNodeID)
			srcRenderableNode.Origins = srcRenderableNode.Origins.Add(srcNodeID)
			edgeID := MakeEdgeID(srcNodeID, dstNodeID)
			if md, ok := t.EdgeMetadatas[edgeID]; ok {
				srcRenderableNode.Metadata.Merge(md.Transform())
			}
		}

		nodes[srcRenderableID] = srcRenderableNode
	}

	return nodes
}

// EdgeMetadata gives the metadata of an edge from the perspective of the
// srcRenderableID. Since an edgeID can have multiple edges on the address
// level, it uses the supplied mapping function to translate address IDs to
// renderable node (mapped) IDs.
func (t Topology) EdgeMetadata(mapFunc MapFunc, srcRenderableID, dstRenderableID string) EdgeMetadata {
	metadata := EdgeMetadata{}
	for edgeID, edgeMeta := range t.EdgeMetadatas {
		src, dst, ok := ParseEdgeID(edgeID)
		if !ok {
			log.Printf("bad edge ID %q", edgeID)
			continue
		}
		if src != TheInternet {
			mapped, _ := mapFunc(t.NodeMetadatas[src])
			src = mapped.ID
		}
		if dst != TheInternet {
			mapped, _ := mapFunc(t.NodeMetadatas[dst])
			dst = mapped.ID
		}
		if src == srcRenderableID && dst == dstRenderableID {
			metadata.Flatten(edgeMeta)
		}
	}
	return metadata
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
