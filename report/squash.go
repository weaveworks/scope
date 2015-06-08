package report

import (
	"log"
	"net"
)

// Squash takes a Topology, and folds all remote nodes into a supernode.
func Squash(t Topology, f IDAddresser, localNets []*net.IPNet) Topology {
	newTopo := NewTopology()
	isRemote := func(ip net.IP) bool { return !netsContain(localNets, ip) }

	// If any node ID on the right-hand (destination) side of an adjacency
	// list is remote, rename it to TheInternet. (We'll never have remote
	// nodes on the left-hand (source) side of an adjacency list, by
	// definition.)
	for nodeID, adjacent := range t.Adjacency {
		var newAdjacency IDList
		for _, adjacentID := range adjacent {
			if isRemote(f(adjacentID)) {
				adjacentID = TheInternet
			}
			newAdjacency = newAdjacency.Add(adjacentID)
		}
		newTopo.Adjacency[nodeID] = newAdjacency
	}

	// Edge metadata keys are "<src node ID>|<dst node ID>". If the dst node
	// ID is remote, rename it to TheInternet.
	for key, metadata := range t.EdgeMetadatas {
		srcNodeID, dstNodeID, ok := ParseEdgeID(key)
		if !ok {
			log.Printf("bad edge ID %q", key)
			continue
		}
		if ip := f(dstNodeID); ip != nil && isRemote(ip) {
			key = MakeEdgeID(srcNodeID, TheInternet)
		}

		// Could be we're merging two keys into one now.
		summedMetadata := newTopo.EdgeMetadatas[key]
		summedMetadata.Flatten(metadata)
		newTopo.EdgeMetadatas[key] = summedMetadata
	}

	newTopo.NodeMetadatas = t.NodeMetadatas
	return newTopo
}

func netsContain(nets []*net.IPNet, ip net.IP) bool {
	for _, net := range nets {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}
