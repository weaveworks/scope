package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/report"
)

var topologyRegistry = map[string]topologyDefinition{
	"applications":         {"Applications", report.SelectEndpoint, report.ProcessPID, report.BasicPseudoNode, "applications-grouped"},
	"applications-grouped": {"Applications", report.SelectEndpoint, report.ProcessName, report.GroupedPseudoNode, ""},
	"containers":           {"Containers", report.SelectEndpoint, report.ProcessContainer, report.NoPseudoNode, "containers-grouped"},
	"containers-grouped":   {"Containers", report.SelectEndpoint, report.ProcessContainerImage, report.NoPseudoNode, ""},
	"hosts":                {"Hosts", report.SelectAddress, report.AddressHostname, report.BasicPseudoNode, ""},
}

func newRouter(r Reporter) http.Handler {
	router := mux.NewRouter()
	getRouter := router.Methods("GET").Subrouter()
	getRouter.HandleFunc("/api", handleAPI)
	getRouter.HandleFunc("/api/topology", handleListTopologies(r))
	getRouter.HandleFunc("/api/topology/{topology}", captureTopology(r, handleTopology))
	getRouter.HandleFunc("/api/topology/{topology}/ws", captureTopology(r, handleTopologyWebsocket))
	getRouter.HandleFunc("/api/topology/{topology}/{id}", captureTopology(r, handleTopologyNode))
	getRouter.HandleFunc("/api/topology/{topology}/{src}/{dst}", captureTopology(r, handleTopologyEdge))
	getRouter.HandleFunc("/api/report", handleReport(r))
	getRouter.PathPrefix("/").Handler(http.FileServer(FS(false))) // static
	return router
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	respondWith(w, http.StatusOK, APIDetails{Version: version})
}

func handleListTopologies(reporter Reporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rpt := reporter.Report()
		var descriptors []APITopologyDescriptor
		for name, definition := range topologyRegistry {
			if strings.HasSuffix(name, "-grouped") {
				continue
			}
			var (
				path        = "/api/topology/" + name
				groupedPath = ""
			)
			if definition.grouped != "" {
				groupedPath = "/api/topology/" + definition.grouped
			}
			descriptors = append(descriptors, APITopologyDescriptor{
				Name:        definition.human,
				Path:        path,
				GroupedPath: groupedPath,
				Stats:       makeTopologyStats(report.Render(rpt, definition.selector, definition.mapper, definition.pseudo)),
			})
		}
		respondWith(w, http.StatusOK, descriptors)
	}
}

func captureTopology(reporter Reporter, f func(Reporter, topologyDefinition, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topology, ok := mux.Vars(r)["topology"]
		if !ok {
			respondWith(w, http.StatusBadRequest, errorResponse("topology not provided"))
			return
		}
		definition, ok := topologyRegistry[topology]
		if !ok {
			respondWith(w, http.StatusNotFound, errorResponse(fmt.Sprintf("%q not found", topology)))
			return
		}
		f(reporter, definition, w, r)
	}
}

func handleTopology(reporter Reporter, definition topologyDefinition, w http.ResponseWriter, r *http.Request) {
	respondWith(w, http.StatusOK, APITopology{
		Nodes: report.Render(reporter.Report(), definition.selector, definition.mapper, definition.pseudo),
	})
}

func handleTopologyWebsocket(reporter Reporter, definition topologyDefinition, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		respondWith(w, http.StatusInternalServerError, err.Error())
		return
	}
	interval := websocketInterval
	if t := r.Form.Get("t"); t != "" {
		var err error
		if interval, err = time.ParseDuration(t); err != nil {
			respondWith(w, http.StatusBadRequest, errorResponse(err.Error()))
			return
		}
	}
	handleWebsocket(w, r, reporter, definition, interval)
}

func handleTopologyNode(reporter Reporter, definition topologyDefinition, w http.ResponseWriter, r *http.Request) {
	var (
		nodeID   = mux.Vars(r)["id"]
		rpt      = reporter.Report()
		node, ok = report.Render(rpt, definition.selector, definition.mapper, definition.pseudo)[nodeID]
	)
	if !ok {
		respondWith(w, http.StatusNotFound, errorResponse(fmt.Sprintf("%q not found", nodeID)))
		return
	}
	respondWith(w, http.StatusOK, APINode{Node: report.MakeDetailedNode(rpt, node)})
}

func handleTopologyEdge(reporter Reporter, definition topologyDefinition, w http.ResponseWriter, r *http.Request) {
	var (
		vars        = mux.Vars(r)
		srcMappedID = vars["src"]
		dstMappedID = vars["dst"]
		rpt         = reporter.Report()
		metadata    = rpt.EdgeMetadata(definition.selector, definition.mapper, srcMappedID, dstMappedID).Export()
	)
	respondWith(w, http.StatusOK, APIEdge{Metadata: metadata})
}

func handleReport(reporter Reporter) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		respondWith(w, http.StatusOK, reporter.Report())
	}
}

func makeTopologyStats(m map[string]report.RenderableNode) topologyStats {
	var (
		nodes     int
		realNodes int
		edges     int
	)
	for _, node := range m {
		nodes++
		if !node.Pseudo {
			realNodes++
		}
		edges += len(node.Adjacency)
	}
	return topologyStats{
		NodeCount:          nodes,
		NonPseudoNodeCount: realNodes,
		EdgeCount:          edges,
	}
}

func respondWith(w http.ResponseWriter, code int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Print(err)
	}
}

func errorResponse(err string) interface{} {
	return map[string]string{"err": err}
}

type topologyDefinition struct {
	human    string
	selector report.TopologySelector
	mapper   report.MapFunc
	pseudo   report.PseudoFunc
	grouped  string
}

// APIDetails are some generic details that can be fetched from /api.
type APIDetails struct {
	Version string `json:"version"`
}

// APITopologyDescriptor is returned by the /api/topology handler.
type APITopologyDescriptor struct {
	Name        string        `json:"name"`
	Path        string        `json:"url"`                   // TODO rename to path
	GroupedPath string        `json:"grouped_url,omitempty"` // TODO rename to grouped_path
	Stats       topologyStats `json:"stats"`
}

// APITopology is returned by the /api/topology/{topology} handler.
type APITopology struct {
	Nodes map[string]report.RenderableNode `json:"nodes"`
}

// APINode is returned by the /api/topology/{topology}/{id} handler.
type APINode struct {
	Node report.DetailedNode `json:"node"`
}

// APIEdge is returned by the /api/topology/{topology}/{src}/{dst} handler.
type APIEdge struct {
	Metadata report.AggregateMetadata `json:"metadata"`
}

type topologyStats struct {
	NodeCount          int `json:"node_count"`
	NonPseudoNodeCount int `json:"non_pseudo_node_count"`
	EdgeCount          int `json:"edge_count"`
}
