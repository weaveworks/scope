package report

import (
	"net"
	"strings"
)

const (
	// TheInternet is the ID that we assign to the super-node composed of all
	// remote nodes that have been squashed together.
	TheInternet = "theinternet"
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
		parts := strings.SplitN(key, IDDelim, 2)
		if ip := f(parts[1]); ip != nil && isRemote(ip) {
			key = parts[0] + IDDelim + TheInternet
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
