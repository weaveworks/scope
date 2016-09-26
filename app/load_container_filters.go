package app

import (
	"encoding/json"
	"github.com/weaveworks/scope/render"
	"os"
)

type filter struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Label string `json:"label"`
}

func getContainerTopologyOptions() ([]APITopologyOption, error) {
	var toptions []APITopologyOption

	// get JSON string from environment variable
	s := os.Getenv("CONTAINER_FILTERS")

	var filters []filter
	json.Unmarshal([]byte(s), &filters)

	for _, f := range filters {
		v := APITopologyOption{Value:f.ID, Label:f.Title, filter:render.IsDesired(f.Label), filterPseudo:false}
		toptions = append(toptions, v)
	}

	// Add option to not view weave system containers
	notSystem := APITopologyOption{Value:"notsystem", Label:"Application Containers", filter:render.IsApplication, filterPseudo:false}
	toptions = append(toptions, notSystem)

	// Add option to view all
	all := APITopologyOption{Value:"all", Label:"All", filter:nil, filterPseudo:false}
	toptions = append(toptions, all)

	return toptions, nil
}
