package app

import (
        "os"
        "fmt"
        "github.com/weaveworks/scope/render"
)

type Filter struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Label string `json:"label"`
}

func getContainerTopologyOptions() ([]APITopologyOption, error) {
        var toptions []APITopologyOption

        // get JSON string from environment variable
        s := os.Getenv("CONTAINER_FILTERS")

	var filters []Filter
	json.Unmarshal([]byte(s), &filters)
	
	for _, f := range filters {
		v := APITopologyOption{f.ID, f.Title, render.IsDesired(f.Label), false}
		toptions = append(toptions,v)
	}
	
	// Add option to view all
	all := APITopologyOption{"all", "All", nil, false}
	toptions = append(toptions,all)

        return toptions, nil
}

