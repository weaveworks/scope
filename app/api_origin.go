package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Origin is returned by the /api/origin/* handler. It represents a machine
// that runs a probe, i.e. the origin of some data in the system.
type Origin struct {
	Hostname    string   `json:"hostname"`
	OS          string   `json:"os"`
	Addresses   []string `json:"addresses"`
	LoadOne     float64  `json:"load_one"`
	LoadFive    float64  `json:"load_five"`
	LoadFifteen float64  `json:"load_fifteen"`
}

// makeOriginHandler makes the /api/origin/* handler.
func makeOriginHandler(rep Reporter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars   = mux.Vars(r)
			nodeID = vars["id"]
		)
		origin, ok := getOrigin(rep, nodeID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		respondWith(w, http.StatusOK, origin)
	}
}

func getOrigin(rep Reporter, nodeID string) (Origin, bool) {
	host, ok := rep.Report().HostMetadatas[nodeID]
	if !ok {
		return Origin{}, false
	}

	var addrs []string
	for _, l := range host.LocalNets {
		addrs = append(addrs, l.String())
	}

	return Origin{
		Hostname:    host.Hostname,
		OS:          host.OS,
		Addresses:   addrs,
		LoadOne:     host.LoadOne,
		LoadFive:    host.LoadFive,
		LoadFifteen: host.LoadFifteen,
	}, true
}
