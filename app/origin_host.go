package main

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/report"
)

// OriginHost represents a host that runs a probe, i.e. the origin host of
// some data in the system. The struct is returned by the /api/origin/{id}
// handler.
type OriginHost struct {
	Hostname string   `json:"hostname"`
	OS       string   `json:"os"`
	Networks []string `json:"networks"`
	Load     string   `json:"load"`
}

func getOriginHost(t report.Topology, nodeID string) (OriginHost, bool) {
	host, ok := t.NodeMetadatas[nodeID]
	if !ok {
		return OriginHost{}, false
	}

	return OriginHost{
		Hostname: host["host_name"],
		OS:       host["os"],
		Networks: strings.Split(host["local_networks"], " "),
		Load:     host["load"],
	}, true
}

// makeOriginHostHandler makes the /api/origin/* handler.
func makeOriginHostHandler(rep Reporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars   = mux.Vars(r)
			nodeID = vars["id"]
		)
		origin, ok := getOriginHost(rep.Report().Host, nodeID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		respondWith(w, http.StatusOK, origin)
	}
}
