package main

import (
	"compress/gzip"
	"encoding/gob"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/ghost/handlers"
	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

// URLMatcher uses request.RequestURI (the raw, unparsed request) to attempt
// to match pattern.  It does this as go's URL.Parse method is broken, and
// mistakenly unescapes the Path before parsing it.  This breaks %2F (encoded
// forward slashes) in the paths.
func URLMatcher(pattern string) mux.MatcherFunc {
	matchParts := strings.Split(pattern, "/")

	return func(r *http.Request, rm *mux.RouteMatch) bool {
		path := strings.SplitN(r.RequestURI, "?", 2)[0]
		parts := strings.Split(path, "/")
		if len(parts) != len(matchParts) {
			return false
		}

		rm.Vars = map[string]string{}
		for i, part := range parts {
			unescaped, err := url.QueryUnescape(part)
			if err != nil {
				return false
			}
			match := matchParts[i]
			if strings.HasPrefix(match, "{") && strings.HasSuffix(match, "}") {
				rm.Vars[strings.Trim(match, "{}")] = unescaped
			} else if matchParts[i] != unescaped {
				return false
			}
		}
		return true
	}
}

type collector interface {
	Reporter
	Adder
}

func gzipHandler(h http.HandlerFunc) http.HandlerFunc {
	return handlers.GZIPHandlerFunc(h, nil)
}

func registerTopologyRoutes(c collector, router *mux.Router) {
	get := router.Methods("GET").Subrouter()
	get.HandleFunc("/api", gzipHandler(apiHandler))
	get.HandleFunc("/api/topology", gzipHandler(topologyRegistry.makeTopologyList(c)))
	get.HandleFunc("/api/topology/{topology}",
		gzipHandler(topologyRegistry.captureRenderer(c, handleTopology)))
	get.HandleFunc("/api/topology/{topology}/ws",
		topologyRegistry.captureRenderer(c, handleWs)) // NB not gzip!
	get.MatcherFunc(URLMatcher("/api/topology/{topology}/{id}")).HandlerFunc(
		gzipHandler(topologyRegistry.captureRenderer(c, handleNode)))
	get.MatcherFunc(URLMatcher("/api/topology/{topology}/{local}/{remote}")).HandlerFunc(
		gzipHandler(topologyRegistry.captureRenderer(c, handleEdge)))
	get.HandleFunc("/api/report", gzipHandler(makeRawReportHandler(c)))

	post := router.Methods("POST").Subrouter()
	post.HandleFunc("/api/report", makeReportPostHandler(c)).Methods("POST")
}

func makeReportPostHandler(a Adder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			rpt    report.Report
			reader = r.Body
			err    error
		)
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			reader, err = gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		if err := gob.NewDecoder(reader).Decode(&rpt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		a.Add(rpt)
		if len(rpt.Pod.Nodes) > 0 {
			topologyRegistry.enableKubernetesTopologies()
		}
		w.WriteHeader(http.StatusOK)
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	respondWith(w, http.StatusOK, xfer.Details{ID: uniqueID, Version: version})
}
