package app

import (
	"net/http"
	"sort"
	"sync"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

const apiTopologyURL = "/api/topology/"

var (
	topologyRegistry = &registry{
		items: map[string]APITopologyDesc{},
	}
)

func init() {
	serviceFilters := []APITopologyOptionGroup{
		{
			ID:      "system",
			Default: "application",
			Options: []APITopologyOption{
				{"system", "System services", render.IsSystem},
				{"application", "Application services", render.IsApplication},
				{"both", "Both", render.Noop},
			},
		},
	}

	podFilters := []APITopologyOptionGroup{
		{
			ID:      "system",
			Default: "application",
			Options: []APITopologyOption{
				{"system", "System pods", render.IsSystem},
				{"application", "Application pods", render.IsApplication},
				{"both", "Both", render.Noop},
			},
		},
	}

	containerFilters := []APITopologyOptionGroup{
		{
			ID:      "system",
			Default: "application",
			Options: []APITopologyOption{
				{"system", "System containers", render.IsSystem},
				{"application", "Application containers", render.IsApplication},
				{"both", "Both", render.Noop},
			},
		},
		{
			ID:      "stopped",
			Default: "running",
			Options: []APITopologyOption{
				{"stopped", "Stopped containers", render.IsStopped},
				{"running", "Running containers", render.IsRunning},
				{"both", "Both", render.Noop},
			},
		},
	}

	unconnectedFilter := []APITopologyOptionGroup{
		{
			ID:      "unconnected",
			Default: "hide",
			Options: []APITopologyOption{
				// Show the user why there are filtered nodes in this view.
				// Don't give them the option to show those nodes.
				{"hide", "Unconnected nodes hidden", render.Noop},
			},
		},
	}

	// Topology option labels should tell the current state. The first item must
	// be the verb to get to that state
	topologyRegistry.add(
		APITopologyDesc{
			id:       "processes",
			renderer: render.FilterUnconnected(render.ProcessWithContainerNameRenderer),
			Name:     "Processes",
			Rank:     1,
			Options:  unconnectedFilter,
		},
		APITopologyDesc{
			id:       "processes-by-name",
			parent:   "processes",
			renderer: render.FilterUnconnected(render.ProcessNameRenderer),
			Name:     "by name",
			Options:  unconnectedFilter,
		},
		APITopologyDesc{
			id:       "containers",
			renderer: render.ContainerWithImageNameRenderer,
			Name:     "Containers",
			Rank:     2,
			Options:  containerFilters,
		},
		APITopologyDesc{
			id:       "containers-by-image",
			parent:   "containers",
			renderer: render.ContainerImageRenderer,
			Name:     "by image",
			Options:  containerFilters,
		},
		APITopologyDesc{
			id:       "containers-by-hostname",
			parent:   "containers",
			renderer: render.ContainerHostnameRenderer,
			Name:     "by DNS name",
			Options:  containerFilters,
		},
		APITopologyDesc{
			id:       "hosts",
			renderer: render.HostRenderer,
			Name:     "Hosts",
			Rank:     4,
		},
		APITopologyDesc{
			id:          "pods",
			renderer:    render.PodRenderer,
			Name:        "Pods",
			Rank:        3,
			HideIfEmpty: true,
			Options:     podFilters,
		},
		APITopologyDesc{
			id:          "pods-by-service",
			parent:      "pods",
			renderer:    render.PodServiceRenderer,
			Name:        "by service",
			HideIfEmpty: true,
			Options:     serviceFilters,
		},
	)
}

// registry is a threadsafe store of the available topologies
type registry struct {
	sync.RWMutex
	items map[string]APITopologyDesc
}

// APITopologyDesc is returned in a list by the /api/topology handler.
type APITopologyDesc struct {
	id       string
	parent   string
	renderer render.Renderer

	Name        string                   `json:"name"`
	Rank        int                      `json:"rank"`
	HideIfEmpty bool                     `json:"hide_if_empty"`
	Options     []APITopologyOptionGroup `json:"options"`

	URL           string            `json:"url"`
	SubTopologies []APITopologyDesc `json:"sub_topologies,omitempty"`
	Stats         topologyStats     `json:"stats,omitempty"`
}

type byName []APITopologyDesc

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// APITopologyOptionGroup describes a group of APITopologyOptions
type APITopologyOptionGroup struct {
	ID      string              `json:"id"`
	Default string              `json:"defaultValue,omitempty"`
	Options []APITopologyOption `json:"options,omitempty"`
}

// APITopologyOption describes a &param=value to a given topology.
type APITopologyOption struct {
	Value string `json:"value"`
	Label string `json:"label"`

	filter render.FilterFunc
}

type topologyStats struct {
	NodeCount          int `json:"node_count"`
	NonpseudoNodeCount int `json:"nonpseudo_node_count"`
	EdgeCount          int `json:"edge_count"`
	FilteredNodes      int `json:"filtered_nodes"`
}

func (r *registry) add(ts ...APITopologyDesc) {
	r.Lock()
	defer r.Unlock()
	for _, t := range ts {
		t.URL = apiTopologyURL + t.id

		if t.parent != "" {
			parent := r.items[t.parent]
			parent.SubTopologies = append(parent.SubTopologies, t)
			sort.Sort(byName(parent.SubTopologies))
			r.items[t.parent] = parent
		}

		r.items[t.id] = t
	}
}

func (r *registry) get(name string) (APITopologyDesc, bool) {
	r.RLock()
	defer r.RUnlock()
	t, ok := r.items[name]
	return t, ok
}

func (r *registry) walk(f func(APITopologyDesc)) {
	r.RLock()
	defer r.RUnlock()
	descs := []APITopologyDesc{}
	for _, desc := range r.items {
		if desc.parent != "" {
			continue
		}
		descs = append(descs, desc)
	}
	sort.Sort(byName(descs))
	for _, desc := range descs {
		f(desc)
	}
}

// makeTopologyList returns a handler that yields an APITopologyList.
func (r *registry) makeTopologyList(rep Reporter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		report, err := rep.Report(ctx)
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err.Error())
			return
		}
		topologies := r.renderTopologies(report, req)
		respondWith(w, http.StatusOK, topologies)
	}
}

func (r *registry) renderTopologies(rpt report.Report, req *http.Request) []APITopologyDesc {
	topologies := []APITopologyDesc{}
	r.walk(func(desc APITopologyDesc) {
		renderer, decorator := renderedForRequest(req, desc)
		desc.Stats = decorateWithStats(rpt, renderer, decorator)
		for i := range desc.SubTopologies {
			renderer, decorator := renderedForRequest(req, desc.SubTopologies[i])
			desc.SubTopologies[i].Stats = decorateWithStats(rpt, renderer, decorator)
		}
		topologies = append(topologies, desc)
	})
	return topologies
}

func decorateWithStats(rpt report.Report, renderer render.Renderer, decorator render.Decorator) topologyStats {
	var (
		nodes     int
		realNodes int
		edges     int
	)
	for _, n := range renderer.Render(rpt, decorator) {
		nodes++
		if n.Topology != render.Pseudo {
			realNodes++
		}
		edges += len(n.Adjacency)
	}
	renderStats := renderer.Stats(rpt, decorator)
	return topologyStats{
		NodeCount:          nodes,
		NonpseudoNodeCount: realNodes,
		EdgeCount:          edges,
		FilteredNodes:      renderStats.FilteredNodes,
	}
}

func renderedForRequest(r *http.Request, topology APITopologyDesc) (render.Renderer, render.Decorator) {
	var filters []render.FilterFunc
	for _, group := range topology.Options {
		value := r.FormValue(group.ID)
		for _, opt := range group.Options {
			if (value == "" && group.Default == opt.Value) || (opt.Value != "" && opt.Value == value) {
				filters = append(filters, opt.filter)
			}
		}
	}
	return topology.renderer, func(renderer render.Renderer) render.Renderer {
		return render.MakeFilter(render.ComposeFilterFuncs(filters...), renderer)
	}
}

type reportRenderHandler func(
	context.Context,
	Reporter, render.Renderer, render.Decorator,
	http.ResponseWriter, *http.Request,
)

func (r *registry) captureRenderer(rep Reporter, f reportRenderHandler) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		topology, ok := r.get(mux.Vars(req)["topology"])
		if !ok {
			http.NotFound(w, req)
			return
		}
		renderer, decorator := renderedForRequest(req, topology)
		f(ctx, rep, renderer, decorator, w, req)
	}
}
