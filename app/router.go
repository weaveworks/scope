package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/render"
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

// Router gives of the HTTP dispatcher. It will always use the embedded HTML
// resources.
func Router(c Reporter) *mux.Router {
	router := mux.NewRouter()
	get := router.Methods("GET").Subrouter()
	get.HandleFunc("/api", apiHandler)
	get.HandleFunc("/api/topology", makeTopologyList(c))
	get.HandleFunc("/api/topology/{topology}", captureTopology(c, handleTopology))
	get.HandleFunc("/api/topology/{topology}/ws", captureTopology(c, handleWs))
	get.MatcherFunc(URLMatcher("/api/topology/{topology}/{id}")).HandlerFunc(captureTopology(c, handleNode))
	get.MatcherFunc(URLMatcher("/api/topology/{topology}/{local}/{remote}")).HandlerFunc(captureTopology(c, handleEdge))
	get.MatcherFunc(URLMatcher("/api/origin/host/{id}")).HandlerFunc(makeOriginHostHandler(c))
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

var topologyRegistry = map[string]topologyView{
	"applications": {
		human:    "Applications",
		parent:   "",
		renderer: render.FilterUnconnected{Renderer: render.ProcessRenderer},
	},
	"applications-by-name": {
		human:    "by name",
		parent:   "applications",
		renderer: render.FilterUnconnected{Renderer: render.ProcessNameRenderer},
	},
	"containers": {
		human:    "Containers",
		parent:   "",
		renderer: render.ContainerRenderer,
	},
	"containers-by-image": {
		human:    "by image",
		parent:   "containers",
		renderer: render.ContainerImageRenderer,
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
}
