package main

import (
	"compress/gzip"
	"encoding/gob"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/ghost/handlers"
	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/render"
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
	xfer.Reporter
	xfer.Adder
}

func gzipHandler(h http.HandlerFunc) http.HandlerFunc {
	return handlers.GZIPHandlerFunc(h, nil)
}

// Router returns the HTTP dispatcher, managing API and UI requests, and
// accepting reports from probes.. It will always use the embedded HTML
// resources for the UI.
func Router(c collector) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/report", makeReportPostHandler(c)).Methods("POST")

	get := router.Methods("GET").Subrouter()
	get.HandleFunc("/api", gzipHandler(apiHandler))
	get.HandleFunc("/api/topology", gzipHandler(makeTopologyList(c)))
	get.HandleFunc("/api/topology/{topology}", gzipHandler(captureTopology(c, handleTopology)))
	get.HandleFunc("/api/topology/{topology}/ws", captureTopology(c, handleWs)) // NB not gzip!
	get.MatcherFunc(URLMatcher("/api/topology/{topology}/{id}")).HandlerFunc(gzipHandler(captureTopology(c, handleNode)))
	get.MatcherFunc(URLMatcher("/api/topology/{topology}/{local}/{remote}")).HandlerFunc(gzipHandler(captureTopology(c, handleEdge)))
	get.MatcherFunc(URLMatcher("/api/origin/host/{id}")).HandlerFunc(gzipHandler(makeOriginHostHandler(c)))
	get.HandleFunc("/api/report", gzipHandler(makeRawReportHandler(c)))
	get.PathPrefix("/").Handler(http.FileServer(FS(false))) // everything else is static

	return router
}

func makeReportPostHandler(a xfer.Adder) http.HandlerFunc {
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
		w.WriteHeader(http.StatusOK)
	}
}

func captureTopology(rep xfer.Reporter, f func(xfer.Reporter, topologyView, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topology, ok := topologyRegistry[mux.Vars(r)["topology"]]
		if !ok {
			http.NotFound(w, r)
			return
		}
		for param, opts := range topology.options {
			value := r.FormValue(param)
			for _, opt := range opts {
				if (value == "" && opt.def) || (opt.value != "" && opt.value == value) {
					topology.renderer = opt.decorator(topology.renderer)
				}
			}
		}
		f(rep, topology, w, r)
	}
}

// APIDetails are some generic details that can be fetched from /api
type APIDetails struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	respondWith(w, http.StatusOK, APIDetails{ID: uniqueID, Version: version})
}

// Topology option labels should tell the current state. The first item must
// be the verb to get to that state
var topologyRegistry = map[string]topologyView{
	"applications": {
		human:    "Applications",
		parent:   "",
		renderer: render.FilterUnconnected(render.ProcessWithContainerNameRenderer{}),
	},
	"applications-by-name": {
		human:    "by name",
		parent:   "applications",
		renderer: render.FilterUnconnected(render.ProcessNameRenderer),
	},
	"containers": {
		human:    "Containers",
		parent:   "",
		renderer: render.ContainerWithImageNameRenderer{},
		options: optionParams{"system": {
			{"show", "System containers shown", false, nop},
			{"hide", "System containers hidden", true, render.FilterSystem},
		}},
	},
	"containers-by-image": {
		human:    "by image",
		parent:   "containers",
		renderer: render.ContainerImageRenderer,
		options: optionParams{"system": {
			{"show", "System containers shown", false, nop},
			{"hide", "System containers hidden", true, render.FilterSystem},
		}},
	},
	"hosts": {
		human:    "Hosts",
		parent:   "",
		renderer: render.HostRenderer,
	},
}

type topologyView struct {
	human    string
	parent   string
	renderer render.Renderer
	options  optionParams
}

type optionParams map[string][]optionValue // param: values

type optionValue struct {
	value     string // "hide"
	human     string // "Hide system containers"
	def       bool
	decorator func(render.Renderer) render.Renderer
}

func nop(r render.Renderer) render.Renderer { return r }
