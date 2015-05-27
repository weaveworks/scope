package main

import (
	"net/http"
	"strings"

	"github.com/weaveworks/scope/report"
)

// APITopologyDesc is returned in a list by the /api/topology handler.
type APITopologyDesc struct {
	Name       string        `json:"name"`
	URL        string        `json:"url"`
	GroupedURL string        `json:"grouped_url,omitempty"`
	Stats      topologyStats `json:"stats"`
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

		var a []APITopologyDesc
		for name, def := range topologyRegistry {
			if strings.HasSuffix(name, "grouped") {
				continue
			}

			url := "/api/topology/" + name
			var groupedURL string
			if def.groupedTopology != "" {
				groupedURL = "/api/topology/" + def.groupedTopology
			}

			a = append(a, APITopologyDesc{
				Name:       def.human,
				URL:        url,
				GroupedURL: groupedURL,
				Stats:      stats(def.selector(rpt).RenderBy(def.mapper, def.pseudo)),
			})
		}
		respondWith(w, http.StatusOK, a)
	}
}

func stats(r map[string]report.RenderableNode) topologyStats {
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

	return topologyStats{
		NodeCount:          nodes,
		NonpseudoNodeCount: realNodes,
		EdgeCount:          edges,
	}
}
