package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/report"
)

// Router gives of the HTTP dispatcher. It will always use the embedded HTML
// resources.
func Router(c Reporter) *mux.Router {
	router := mux.NewRouter()
	get := router.Methods("GET").Subrouter()
	get.HandleFunc("/api", apiHandler)
	get.HandleFunc("/api/topology", makeTopologyList(c))
	get.HandleFunc("/api/topology/{topology}", captureTopology(c, handleTopology))
	get.HandleFunc("/api/topology/{topology}/ws", captureTopology(c, handleWs))
	get.HandleFunc("/api/topology/{topology}/{id}", captureTopology(c, handleNode))
	get.HandleFunc("/api/topology/{topology}/{local}/{remote}", captureTopology(c, handleEdge))
	get.HandleFunc("/api/origin/host/{id}", makeOriginHostHandler(c))
	get.HandleFunc("/api/report", makeRawReportHandler(c))
	get.PathPrefix("/").Handler(http.FileServer(FS(false))) // everything else is static
	return router
}

func captureTopology(rep Reporter, f func(Reporter, topologyView, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		topology, ok := topologyRegistry[mux.Vars(r)["topology"]]
		if !ok {
			http.NotFound(w, r)
			return
		}
		f(rep, topology, w, r)
	}
}

// APIDetails are some generic details that can be fetched from /api
type APIDetails struct {
	Version string `json:"version"`
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	respondWith(w, http.StatusOK, APIDetails{Version: version})
}

type topologyView struct {
	human           string
	selector        report.TopologySelector
	mapper          report.MapFunc
	pseudo          report.PseudoFunc
	groupedTopology string
}

var topologyRegistry = map[string]topologyView{
	"applications":         {"Applications", report.SelectProcess, report.ProcessPID, report.GenericPseudoNode, "applications-grouped"},
	"applications-grouped": {"Applications", report.SelectProcess, report.ProcessName, report.GenericGroupedPseudoNode, ""},
	"containers":           {"Containers", report.SelectProcess, report.ProcessContainer, report.InternetOnlyPseudoNode, "containers-grouped"},
	"containers-grouped":   {"Containers", report.SelectProcess, report.ProcessContainerImage, report.InternetOnlyPseudoNode, ""},
	"hosts":                {"Hosts", report.SelectNetwork, report.NetworkHostname, report.GenericPseudoNode, ""},
}
