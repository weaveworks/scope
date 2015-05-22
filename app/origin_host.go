package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/report"
)

// OriginHost represents a host that runs a probe, i.e. the origin host of
// some data in the system. The struct is returned by the /api/origin/{id}
// handler.
type OriginHost struct {
	Hostname    string   `json:"hostname"`
	OS          string   `json:"os"`
	Addresses   []string `json:"addresses"`
	LoadOne     float64  `json:"load_one"`
	LoadFive    float64  `json:"load_five"`
	LoadFifteen float64  `json:"load_fifteen"`
}

func getOriginHost(mds report.HostMetadatas, nodeID string) (OriginHost, bool) {
	host, ok := mds[nodeID]
	if !ok {
		return OriginHost{}, false
	}

	var addrs []string
	for _, l := range host.LocalNets {
		addrs = append(addrs, l.String())
	}

	return OriginHost{
		Hostname:    host.Hostname,
		OS:          host.OS,
		Addresses:   addrs,
		LoadOne:     host.LoadOne,
		LoadFive:    host.LoadFive,
		LoadFifteen: host.LoadFifteen,
	}, true
}

// makeOriginHostHandler makes the /api/origin/* handler.
func makeOriginHostHandler(rep Reporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars   = mux.Vars(r)
			nodeID = vars["id"]
		)
		origin, ok := getOriginHost(rep.Report().HostMetadatas, nodeID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		respondWith(w, http.StatusOK, origin)
	}
}
