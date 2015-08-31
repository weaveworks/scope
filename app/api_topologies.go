package main

import (
	"net/http"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/xfer"
)

// APITopologyDesc is returned in a list by the /api/topology handler.
type APITopologyDesc struct {
	Name          string                         `json:"name"`
	URL           string                         `json:"url"`
	SubTopologies []APITopologyDesc              `json:"sub_topologies,omitempty"`
	Options       map[string][]APITopologyOption `json:"options"`
	Stats         *topologyStats                 `json:"stats,omitempty"`
}

// APITopologyOption describes a &param=value to a given topology.
type APITopologyOption struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Default bool   `json:"default,omitempty"`
}

type topologyStats struct {
	NodeCount          int `json:"node_count"`
	NonpseudoNodeCount int `json:"nonpseudo_node_count"`
	EdgeCount          int `json:"edge_count"`
}

// makeTopologyList returns a handler that yields an APITopologyList.
func makeTopologyList(rep xfer.Reporter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			rpt        = rep.Report()
			topologies = []APITopologyDesc{}
		)
		for name, def := range topologyRegistry {
			// Don't show sub-topologies at the top level.
			if def.parent != "" {
				continue
			}

			// Collect all sub-topologies of this one, depth=1 only.
			subTopologies := []APITopologyDesc{}
			for subName, subDef := range topologyRegistry {
				if subDef.parent == name {
					subTopologies = append(subTopologies, APITopologyDesc{
						Name:    subDef.human,
						URL:     "/api/topology/" + subName,
						Options: makeTopologyOptions(subDef),
						Stats:   stats(subDef.renderer.Render(rpt)),
					})
				}
			}

			// Append.
			topologies = append(topologies, APITopologyDesc{
				Name:          def.human,
				URL:           "/api/topology/" + name,
				SubTopologies: subTopologies,
				Options:       makeTopologyOptions(def),
				Stats:         stats(def.renderer.Render(rpt)),
			})
		}
		respondWith(w, http.StatusOK, topologies)
	}
}

func makeTopologyOptions(view topologyView) map[string][]APITopologyOption {
	options := map[string][]APITopologyOption{}
	for param, optionVals := range view.options {
		for _, optionVal := range optionVals {
			options[param] = append(options[param], APITopologyOption{
				Value:   optionVal.value,
				Display: optionVal.human,
				Default: optionVal.def,
			})
		}
	}
	return options
}

func stats(r render.RenderableNodes) *topologyStats {
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
