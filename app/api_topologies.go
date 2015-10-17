package main

import (
	"net/http"
	"sort"
	"sync"

	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

const apiTopologyURL = "/api/topology/"

var (
	topologyRegistry = &registry{
		items: map[string]APITopologyDesc{},
	}
	kubernetesTopologies = []APITopologyDesc{
		{
			id:       "pods",
			renderer: render.PodRenderer,
			Name:     "Pods",
			Options: map[string][]APITopologyOption{"system": {
				{"show", "System containers shown", false, nop},
				{"hide", "System containers hidden", true, render.FilterSystem},
			}},
		},
		{
			id:       "pods-by-service",
			parent:   "pods",
			renderer: render.PodServiceRenderer,
			Name:     "by service",
			Options: map[string][]APITopologyOption{"system": {
				{"show", "System containers shown", false, nop},
				{"hide", "System containers hidden", true, render.FilterSystem},
			}},
		},
	}
)

func init() {
	// Topology option labels should tell the current state. The first item must
	// be the verb to get to that state
	topologyRegistry.add(
		APITopologyDesc{
			id:       "applications",
			renderer: render.FilterUnconnected(render.ProcessWithContainerNameRenderer),
			Name:     "Applications",
			Options: map[string][]APITopologyOption{"unconnected": {
				// Show the user why there are filtered nodes in this view.
				// Don't give them the option to show those nodes.
				{"hide", "Unconnected nodes hidden", true, nop},
			}},
		},
		APITopologyDesc{
			id:       "applications-by-name",
			parent:   "applications",
			renderer: render.FilterUnconnected(render.ProcessNameRenderer),
			Name:     "by name",
			Options: map[string][]APITopologyOption{"unconnected": {
				// Ditto above.
				{"hide", "Unconnected nodes hidden", true, nop},
			}},
		},
		APITopologyDesc{
			id:       "containers",
			renderer: render.ContainerWithImageNameRenderer,
			Name:     "Containers",
			Options: map[string][]APITopologyOption{"system": {
				{"show", "System containers shown", false, nop},
				{"hide", "System containers hidden", true, render.FilterSystem},
			}},
		},
		APITopologyDesc{
			id:       "containers-by-image",
			parent:   "containers",
			renderer: render.ContainerImageRenderer,
			Name:     "by image",
			Options: map[string][]APITopologyOption{"system": {
				{"show", "System containers shown", false, nop},
				{"hide", "System containers hidden", true, render.FilterSystem},
			}},
		},
		APITopologyDesc{
			id:       "containers-by-hostname",
			parent:   "containers",
			renderer: render.ContainerHostnameRenderer,
			Name:     "by hostname",
			Options: map[string][]APITopologyOption{"system": {
				{"show", "System containers shown", false, nop},
				{"hide", "System containers hidden", true, render.FilterSystem},
			}},
		},
		APITopologyDesc{
			id:       "hosts",
			renderer: render.HostRenderer,
			Name:     "Hosts",
			Options:  map[string][]APITopologyOption{},
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

	Name    string                         `json:"name"`
	Options map[string][]APITopologyOption `json:"options"`

	URL           string            `json:"url"`
	SubTopologies []APITopologyDesc `json:"sub_topologies,omitempty"`
	Stats         *topologyStats    `json:"stats,omitempty"`
}

// APITopologyOption describes a &param=value to a given topology.
type APITopologyOption struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Default bool   `json:"default,omitempty"`

	decorator func(render.Renderer) render.Renderer
}

type topologyStats struct {
	NodeCount          int `json:"node_count"`
	NonpseudoNodeCount int `json:"nonpseudo_node_count"`
	EdgeCount          int `json:"edge_count"`
	FilteredNodes      int `json:"filtered_nodes"`
}

type byName []APITopologyDesc

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (r *registry) add(ts ...APITopologyDesc) {
	r.Lock()
	defer r.Unlock()
	for _, t := range ts {
		t.URL = apiTopologyURL + t.id

		if t.parent != "" {
			parent := r.items[t.parent]
			parent.SubTopologies = append(r.items[t.parent].SubTopologies, t)
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
func makeTopologyList(rep xfer.Reporter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			rpt        = rep.Report()
			topologies = []APITopologyDesc{}
		)
		topologyRegistry.walk(func(desc APITopologyDesc) {
			decorateTopologyForRequest(r, &desc)
			decorateWithStats(&desc, rpt)
			for i := range desc.SubTopologies {
				decorateTopologyForRequest(r, &desc.SubTopologies[i])
				decorateWithStats(&desc.SubTopologies[i], rpt)
			}
			topologies = append(topologies, desc)
		})
		respondWith(w, http.StatusOK, topologies)
	}
}

func decorateWithStats(desc *APITopologyDesc, rpt report.Report) {
	var (
		nodes     int
		realNodes int
		edges     int
	)
	for _, n := range desc.renderer.Render(rpt) {
		nodes++
		if !n.Pseudo {
			realNodes++
		}
		edges += len(n.Adjacency)
	}
	renderStats := desc.renderer.Stats(rpt)
	desc.Stats = &topologyStats{
		NodeCount:          nodes,
		NonpseudoNodeCount: realNodes,
		EdgeCount:          edges,
		FilteredNodes:      renderStats.FilteredNodes,
	}
}

func nop(r render.Renderer) render.Renderer { return r }

func enableKubernetesTopologies() {
	topologyRegistry.add(kubernetesTopologies...)
}

func decorateTopologyForRequest(r *http.Request, topology *APITopologyDesc) {
	for param, opts := range topology.Options {
		value := r.FormValue(param)
		for _, opt := range opts {
			if (value == "" && opt.Default) || (opt.Value != "" && opt.Value == value) {
				topology.renderer = opt.decorator(topology.renderer)
			}
		}
	}
}

func captureTopology(rep xfer.Reporter, f func(xfer.Reporter, APITopologyDesc, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topology, ok := topologyRegistry.get(mux.Vars(r)["topology"])
		if !ok {
			http.NotFound(w, r)
			return
		}
		decorateTopologyForRequest(r, &topology)
		f(rep, topology, w, r)
	}
}
