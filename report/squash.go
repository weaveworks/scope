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
func Squash(t Topology, c IDAddresser, localNets []*net.IPNet) Topology {
	nt := NewTopology()
	for id, a := range t.Adjacency {
		var newAdj IDList

		for _, addrID := range a {
			ip := c(addrID)

			local := false
			for _, localNet := range localNets {
				if localNet.Contains(ip) {
					local = true
					break
				}
			}

			if !local {
				addrID = TheInternet
			}

			newAdj = newAdj.Add(addrID)
		}

		nt.Adjacency[id] = newAdj
	}

	for key, md := range t.EdgeMetadatas {
		parts := strings.SplitN(key, IDDelim, 2)
		if ip := c(parts[1]); ip != nil {
			local := false
			for _, localNet := range localNets {
				if localNet.Contains(ip) {
					local = true
					break
				}
			}
			if !local {
				key = parts[0] + IDDelim + TheInternet
			}
		}

		// Could be we're merging two keys into one now.
		summedMeta := nt.EdgeMetadatas[key]
		summedMeta.Flatten(md)
		nt.EdgeMetadatas[key] = summedMeta
	}

	nt.NodeMetadatas = t.NodeMetadatas
	return nt
}
