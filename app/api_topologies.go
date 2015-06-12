package main

import (
	"net/http"

	"github.com/weaveworks/scope/report"
)

// APITopologyDesc is returned in a list by the /api/topology handler.
type APITopologyDesc struct {
	Name          string            `json:"name"`
	URL           string            `json:"url"`
	SubTopologies []APITopologyDesc `json:"sub_topologies,omitempty"`
	Stats         *topologyStats    `json:"stats,omitempty"`
}

type topologyStats struct {
	NodeCount          int `json:"node_count"`
	NonpseudoNodeCount int `json:"nonpseudo_node_count"`
	EdgeCount          int `json:"edge_count"`
}

// makeTopologyList returns a handler that yields an APITopologyList.
func makeTopologyList(rep Reporter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		rpt := rep.Report()
		topologies := []APITopologyDesc{}
		for name, def := range topologyRegistry {
			subTopologies := []APITopologyDesc{}
			for subName, subDef := range topologyRegistry {
				if subDef.parent == name {
					subTopologies = append(subTopologies, APITopologyDesc{
						Name: subDef.human,
						URL:  "/api/topology/" + subName,
					})
				}
			}
			topologies = append(topologies, APITopologyDesc{
				Name:          def.human,
				URL:           "/api/topology/" + name,
				SubTopologies: subTopologies,
				Stats:         stats(render(rpt, def.maps)),
			})
		}
		respondWith(w, http.StatusOK, topologies)
	}
}

func stats(r report.RenderableNodes) *topologyStats {
	var (
		nodes     int
		realNodes int
		edges     int
	)

	for _, n := range r {
		nodes++
		if !n.Pseudo {
			realNodes++
		}
		edges += len(n.Adjacency)
	}

	return &topologyStats{
		NodeCount:          nodes,
		NonpseudoNodeCount: realNodes,
		EdgeCount:          edges,
	}
}
