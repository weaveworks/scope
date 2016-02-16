package app

import (
	"encoding/gob"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/ghost/handlers"
	"github.com/gorilla/mux"

	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/scope/common/hostname"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/report"
)

var (
	// Version - set at buildtime.
	Version = "dev"

	// UniqueID - set at runtime.
	UniqueID = "0"
)

// URLMatcher uses request.RequestURI (the raw, unparsed request) to attempt
// to match pattern.  It does this as go's URL.Parse method is broken, and
// mistakenly unescapes the Path before parsing it.  This breaks %2F (encoded
// forward slashes) in the paths.
func URLMatcher(pattern string) mux.MatcherFunc {
	return func(r *http.Request, rm *mux.RouteMatch) bool {
		vars, match := matchURL(r, pattern)
		if match {
			rm.Vars = vars
		}
		return match
	}
}

func matchURL(r *http.Request, pattern string) (map[string]string, bool) {
	matchParts := strings.Split(pattern, "/")
	path := strings.SplitN(r.RequestURI, "?", 2)[0]
	parts := strings.Split(path, "/")
	if len(parts) != len(matchParts) {
		return nil, false
	}

	vars := map[string]string{}
	for i, part := range parts {
		unescaped, err := url.QueryUnescape(part)
		if err != nil {
			return nil, false
		}
		match := matchParts[i]
		if strings.HasPrefix(match, "{") && strings.HasSuffix(match, "}") {
			vars[strings.Trim(match, "{}")] = unescaped
		} else if matchParts[i] != unescaped {
			return nil, false
		}
	}
	return vars, true
}

func gzipHandler(h http.HandlerFunc) http.HandlerFunc {
	return handlers.GZIPHandlerFunc(h, nil)
}

// TopologyHandler registers the various topology routes with a http mux.
//
// The returned http.Handler has to be passed directly to http.ListenAndServe,
// and cannot be nested inside another gorrilla.mux.
//
// Routes which should be matched before the topology routes should be added
// to a router and passed in preRoutes.  Routes to be matches after topology
// routes should be added to a router and passed to postRoutes.
func TopologyHandler(c Reporter, preRoutes *mux.Router, postRoutes http.Handler) http.Handler {
	get := preRoutes.Methods("GET").Subrouter()
	get.HandleFunc("/api", gzipHandler(apiHandler))
	get.HandleFunc("/api/topology", gzipHandler(topologyRegistry.makeTopologyList(c)))
	get.HandleFunc("/api/topology/{topology}",
		gzipHandler(topologyRegistry.captureRenderer(c, handleTopology)))
	get.HandleFunc("/api/topology/{topology}/ws",
		topologyRegistry.captureRenderer(c, handleWs)) // NB not gzip!

	// We pull in the http.DefaultServeMux to get the pprof routes
	preRoutes.PathPrefix("/debug/pprof").Handler(http.DefaultServeMux)

	if postRoutes != nil {
		preRoutes.PathPrefix("/").Handler(postRoutes)
	}

	// We have to handle the node details path manually due to
	// various bugs in gorilla.mux and Go URL parsing.  See #804.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars, match := matchURL(r, "/api/topology/{topology}/{id}")
		if !match {
			preRoutes.ServeHTTP(w, r)
			return
		}

		topologyID := vars["topology"]
		nodeID := vars["id"]
		if nodeID == "ws" {
			preRoutes.ServeHTTP(w, r)
			return
		}

		handler := gzipHandler(topologyRegistry.captureRendererWithoutFilters(
			c, topologyID, handleNode(nodeID),
		))
		handler.ServeHTTP(w, r)
	})
}

// RegisterReportPostHandler registers the handler for report submission
func RegisterReportPostHandler(a Adder, router *mux.Router) {
	post := router.Methods("GET").Subrouter()
	post.HandleFunc("/api/report", func(w http.ResponseWriter, r *http.Request) {
		var rpt report.Report

		log.Debug("Entering report endpoint")

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Errorf("Error upgrading report websocket: %s", err)
			return
		}
		defer conn.Close()

		rawConn := conn.UnderlyingConn()
		// no gzip for now
		decoder := gob.NewDecoder(rawConn)
		for {
			if err = decoder.Decode(&rpt); err != nil {
				log.Errorf("Error decoding report, probe died?: %s", err)
				return
			}
			a.Add(rpt)
			if len(rpt.Pod.Nodes) > 0 {
				topologyRegistry.enableKubernetesTopologies()
			}

		}
	})
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	respondWith(w, http.StatusOK, xfer.Details{
		ID:       UniqueID,
		Version:  Version,
		Hostname: hostname.Get(),
	})
}
