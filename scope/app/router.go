// HTTP server routing table
package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/weaveworks/scope/scope/report"
)

// Router gives of the HTTP dispatcher. It will always use the embedded HTML
// resources.
func Router(c Reporter) *mux.Router {
	router := mux.NewRouter()
	get := router.Methods("GET").Subrouter()
	get.HandleFunc("/api/topology", makeTopologyList(c))
	for name, def := range topologyRegistry {
		makeTopologyHandlers(
			c,
			def.topologySelecter,
			def.MapFunc,
			false, // not grouped
			get,
			"/api/topology/"+name,
		)
		if def.hasGrouped {
			makeTopologyHandlers(
				c,
				def.topologySelecter,
				def.MapFunc,
				true, // grouped
				get,
				"/api/topology/"+name+"grouped",
			)
		}
	}
	get.HandleFunc("/api/origin/{id}", makeOriginHandler(c))
	get.HandleFunc("/api/report", makeRawReportHandler(c))
	get.PathPrefix("/").Handler(http.FileServer(FS(false))) // everything else is static
	return router
}

var topologyRegistry = map[string]struct {
	human string
	topologySelecter
	report.MapFunc
	hasGrouped bool
	typ        string
}{
	"processpid":    {"Process ID", selectProcess, report.ProcessPID, true, "Process"},
	"processname":   {"Process name", selectProcess, report.ProcessName, true, "Process"},
	"processcgroup": {"Process cgroup", selectProcess, report.ProcessCgroup, true, "Process"},
	"networkip":     {"Network IP", selectNetwork, report.NetworkIP, false, "Network"},
	"networkhost":   {"Network Host", selectNetwork, report.NetworkHostname, false, "Network"},
}
