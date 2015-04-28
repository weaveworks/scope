package report

import (
	"log"
	"net"
	"reflect"
	"strings"
)

const (
	// ScopeDelim separates the scope portion of an address from the address
	// string itself.
	ScopeDelim = ";"

	// IDDelim separates fields in a node ID.
	IDDelim = "|"
)

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
	WithBytes        bool
	BytesIngress     uint
	BytesEgress      uint
	WithConnCountTCP bool
	MaxConnCountTCP  uint
}

// NodeMetadata describes a superset of the metadata that probes can collect
// about a given node in a given topology. Right now it's a weakly-typed map,
// which should probably change (see comment on type MapFunc).
type NodeMetadata map[string]string

// NewTopology gives you a Topology.
func NewTopology() Topology {
	return Topology{
		Adjacency:     map[string]IDList{},
		EdgeMetadatas: map[string]EdgeMetadata{},
		NodeMetadatas: map[string]NodeMetadata{},
	}
}

// RenderBy translates a given Topology into something consumable by the
// JavaScript layer and renderable to a user. It takes as an argument a
// MapFunc, which defines how to group and label nodes in the output.
//
// If the result will be given to an /api/topology/:name, it should first
// be Downcast to map[string]RenderableNode.
func (t Topology) RenderBy(
	f MapFunc,
	classView bool,
	tps ThirdPartyTemplates,
) map[string]DetailedRenderableNode {
	nodes := map[string]DetailedRenderableNode{}

	// Build RenderableNodes for all non-pseudo probes, and an addressID to
	// nodeID lookup map. Multiple addressIDs can map to the same
	// RenderableNodes.
	nodeAddresses := map[string]string{}
	for addressID, meta := range t.NodeMetadatas {
		mapped, ok := f(addressID, meta, classView)
		if !ok {
			continue
		}

		tp, err := tps.Execute(mapped)
		if err != nil {
			log.Printf("thirdparty template error: %s", err)
		}
		// ID needs not be unique.
		nodes[mapped.ID] = DetailedRenderableNode{
			RenderableNode: RenderableNode{
				ID:         mapped.ID,
				LabelMajor: mapped.Major,
				LabelMinor: mapped.Minor,
				Rank:       mapped.Rank,
				Pseudo:     false,
			},
			Aggregate:  RenderableMetadata{},
			ThirdParty: tp,
		}

		nodeAddresses[addressID] = mapped.ID
	}

	for local, remotes := range t.Adjacency {
		var (
			fields       = strings.SplitN(local, IDDelim, 2) // "<host>|<address>"
			origin       = fields[0]
			localAddress = fields[1]
			localID      = nodeAddresses[localAddress] // must exist
			localNode    = nodes[localID]              // must exist
		)

		for _, remoteAddress := range remotes {
			remoteID, ok := nodeAddresses[remoteAddress]
			if !ok {
				// No node, make a pseudo-node for this address.
				remoteID = remoteAddress
				if remoteID != TheInternet {
					remoteID = "pseudo:" + remoteID
					if classView {
						remoteID = "localunknown"
					}
				}
				if classView {
					nodes[remoteID] = DetailedRenderableNode{
						RenderableNode: RenderableNode{
							ID:         remoteID,
							LabelMajor: "",
							LabelMinor: "",
							Pseudo:     true,
						},
						Aggregate: RenderableMetadata{},
						// no Third Party for pseudo nodes
					}
				} else {
					remoteLabelMajor, remoteLabelMinor := formatLabel(remoteAddress)
					nodes[remoteID] = DetailedRenderableNode{
						RenderableNode: RenderableNode{
							ID:         remoteID,
							LabelMajor: remoteLabelMajor,
							LabelMinor: remoteLabelMinor,
							// No rank for pseudo nodes.
							Pseudo: true,
						},
						Aggregate: RenderableMetadata{},
						// no Third Party for pseudo nodes
					}
				}
				nodeAddresses[remoteAddress] = remoteID
			}
			localNode.RenderableNode.Origin = localNode.RenderableNode.Origin.Add(origin)
			localNode.RenderableNode.Adjacency = localNode.RenderableNode.Adjacency.Add(remoteID)

			edgeID := localAddress + IDDelim + remoteAddress
			if md, ok := t.EdgeMetadatas[edgeID]; ok {
				localNode.Aggregate.Merge(md.Render())
			}
		}

		nodes[localID] = localNode
	}

	return nodes
}

// Downcast converts a map[string]DetailedRenderableNode (the output of
// RenderBy) to a map[string]RenderableNode, which is what should be returned
// by a plain /api/topology/:name handler.
func Downcast(in map[string]DetailedRenderableNode) map[string]RenderableNode {
	out := make(map[string]RenderableNode, len(in))
	for k, v := range in {
		out[k] = v.RenderableNode
	}
	return out
}

// EdgeMetadata gives the metadata of an edge from the perspective of the
// localMappedID. Since an edgeID can have multiple edges on the address
// level, it uses the supplied mapping function to translate addressIDs to
// mappedIDs.
func (t Topology) EdgeMetadata(f MapFunc, classView bool, localMappedID, remoteMappedID string) EdgeMetadata {
	metadata := EdgeMetadata{}
	for edgeID, edgeMeta := range t.EdgeMetadatas {
		edgeParts := strings.SplitN(edgeID, IDDelim, 2)
		localID := edgeParts[0]
		if localID != TheInternet {
			mapped, _ := f(localID, t.NodeMetadatas[localID], classView)
			localID = mapped.ID
		}
		remoteID := edgeParts[1]
		if remoteID != TheInternet {
			mapped, _ := f(remoteID, t.NodeMetadatas[remoteID], classView)
			remoteID = mapped.ID
		}
		if localID == localMappedID && remoteID == remoteMappedID {
			metadata.Flatten(edgeMeta)
		}
	}
	return metadata
}

// formatLabel is an opportunistic helper to format any addressID into
// something we can show on screen.
func formatLabel(s string) (string, string) {
	if s == TheInternet {
		return "the Internet", ""
	}

	// Format is either "scope;ip;port", "scope;ip", or some process id.
	parts := strings.SplitN(s, ScopeDelim, 3)
	if len(parts) < 2 {
		return s, ""
	}

	if len(parts) == 2 {
		return parts[1], ""
	}

	return net.JoinHostPort(parts[1], parts[2]), ""
}

// Diff is returned by TopoDiff. It represents the changes between two
// RenderableNode maps.
type Diff struct {
	Add    []DetailedRenderableNode `json:"add"`
	Update []DetailedRenderableNode `json:"update"`
	Remove []string                 `json:"remove"`
}

// TopoDiff gives you the diff to get from A to B.
func TopoDiff(a, b map[string]DetailedRenderableNode) Diff {
	diff := Diff{}

	notSeen := map[string]struct{}{}
	for k := range a {
		notSeen[k] = struct{}{}
	}

	for k, node := range b {
		if _, ok := a[k]; !ok {
			diff.Add = append(diff.Add, node)
		} else {
			if !reflect.DeepEqual(node, a[k]) {
				diff.Update = append(diff.Update, node)
			}
		}
		delete(notSeen, k)
	}

	// leftover keys
	for k := range notSeen {
		diff.Remove = append(diff.Remove, k)
	}

	return diff
}

// ByID is a sort interface for a DetailedRenderableNode slice.
type ByID []DetailedRenderableNode

func (r ByID) Len() int           { return len(r) }
func (r ByID) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ByID) Less(i, j int) bool { return r[i].RenderableNode.ID < r[j].RenderableNode.ID }
