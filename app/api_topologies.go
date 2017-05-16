package app

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

const (
	apiTopologyURL         = "/api/topology/"
	processesID            = "processes"
	processesByNameID      = "processes-by-name"
	systemGroupID          = "system"
	containersID           = "containers"
	containersByHostnameID = "containers-by-hostname"
	containersByImageID    = "containers-by-image"
	podsID                 = "pods"
	replicaSetsID          = "replica-sets"
	deploymentsID          = "deployments"
	daemonsetsID           = "daemonsets"
	servicesID             = "services"
	hostsID                = "hosts"
	weaveID                = "weave"
	ecsTasksID             = "ecs-tasks"
	ecsServicesID          = "ecs-services"
	swarmServicesID        = "swarm-services"
)

var (
	topologyRegistry = MakeRegistry()
	unmanagedFilter  = APITopologyOptionGroup{
		ID:      "pseudo",
		Default: "hide",
		Options: []APITopologyOption{
			{Value: "show", Label: "Show Unmanaged", filter: nil, filterPseudo: false},
			{Value: "hide", Label: "Hide Unmanaged", filter: render.IsNotPseudo, filterPseudo: true},
		},
	}
)

// namespaceFilters generates a namespace selector option group based on the given namespaces
func namespaceFilters(namespaces []string, noneLabel string) APITopologyOptionGroup {
	options := APITopologyOptionGroup{ID: "namespace", Default: "none", SelectType: "union", NoneLabel: noneLabel}
	for _, namespace := range namespaces {
		options.Options = append(options.Options, APITopologyOption{
			Value: namespace, Label: namespace, filter: render.IsNamespace(namespace), filterPseudo: false,
		})
	}
	return options
}

// updateFilters updates the available filters based on the current report.
func updateFilters(rpt report.Report, topologies []APITopologyDesc) []APITopologyDesc {
	topologies = updateKubeFilters(rpt, topologies)
	topologies = updateSwarmFilters(rpt, topologies)
	return topologies
}

func updateSwarmFilters(rpt report.Report, topologies []APITopologyDesc) []APITopologyDesc {
	namespaces := map[string]struct{}{}
	for _, n := range rpt.SwarmService.Nodes {
		if namespace, ok := n.Latest.Lookup(docker.StackNamespace); ok {
			namespaces[namespace] = struct{}{}
		}
	}
	if len(namespaces) == 0 {
		// We only want to apply filters when we have swarm-related nodes,
		// so if we don't then return early
		return topologies
	}
	ns := []string{}
	for namespace := range namespaces {
		ns = append(ns, namespace)
	}
	topologies = append([]APITopologyDesc{}, topologies...) // Make a copy so we can make changes safely
	for i, t := range topologies {
		if t.id == containersID || t.id == swarmServicesID {
			topologies[i] = mergeTopologyFilters(t, []APITopologyOptionGroup{
				namespaceFilters(ns, "All Stacks"),
			})
		}
	}
	return topologies
}

func updateKubeFilters(rpt report.Report, topologies []APITopologyDesc) []APITopologyDesc {
	namespaces := map[string]struct{}{}
	for _, t := range []report.Topology{rpt.Pod, rpt.Service, rpt.Deployment, rpt.ReplicaSet} {
		for _, n := range t.Nodes {
			if state, ok := n.Latest.Lookup(kubernetes.State); ok && state == kubernetes.StateDeleted {
				continue
			}
			if namespace, ok := n.Latest.Lookup(kubernetes.Namespace); ok {
				namespaces[namespace] = struct{}{}
			}
		}
	}
	if len(namespaces) == 0 {
		// We only want to apply k8s filters when we have k8s-related nodes,
		// so if we don't then return early
		return topologies
	}
	ns := []string{}
	for namespace := range namespaces {
		ns = append(ns, namespace)
	}
	sort.Strings(ns)
	topologies = append([]APITopologyDesc{}, topologies...) // Make a copy so we can make changes safely
	for i, t := range topologies {
		if t.id == containersID || t.id == podsID || t.id == servicesID || t.id == deploymentsID || t.id == replicaSetsID || t.id == daemonsetsID {
			topologies[i] = mergeTopologyFilters(t, []APITopologyOptionGroup{
				namespaceFilters(ns, "All Namespaces"),
			})
		}
	}
	return topologies
}

// mergeTopologyFilters recursively merges in new options on a topology description
func mergeTopologyFilters(t APITopologyDesc, options []APITopologyOptionGroup) APITopologyDesc {
	t.Options = append(append([]APITopologyOptionGroup{}, t.Options...), options...)
	newSubTopologies := make([]APITopologyDesc, len(t.SubTopologies))
	for i, sub := range t.SubTopologies {
		newSubTopologies[i] = mergeTopologyFilters(sub, options)
	}
	t.SubTopologies = newSubTopologies
	return t
}

// MakeAPITopologyOption provides an external interface to the package for creating an APITopologyOption.
func MakeAPITopologyOption(value string, label string, filterFunc render.FilterFunc, pseudo bool) APITopologyOption {
	return APITopologyOption{Value: value, Label: label, filter: filterFunc, filterPseudo: pseudo}
}

// Registry is a threadsafe store of the available topologies
type Registry struct {
	sync.RWMutex
	items map[string]APITopologyDesc
}

// MakeRegistry returns a new Registry
func MakeRegistry() *Registry {
	registry := &Registry{
		items: map[string]APITopologyDesc{},
	}
	containerFilters := []APITopologyOptionGroup{
		{
			ID:      systemGroupID,
			Default: "application",
			Options: []APITopologyOption{
				{Value: "all", Label: "All", filter: nil, filterPseudo: false},
				{Value: "system", Label: "System Containers", filter: render.IsSystem, filterPseudo: false},
				{Value: "application", Label: "Application Containers", filter: render.IsApplication, filterPseudo: false}},
		},
		{
			ID:      "stopped",
			Default: "running",
			Options: []APITopologyOption{
				{Value: "stopped", Label: "Stopped containers", filter: render.IsStopped, filterPseudo: false},
				{Value: "running", Label: "Running containers", filter: render.IsRunning, filterPseudo: false},
				{Value: "both", Label: "Both", filter: nil, filterPseudo: false},
			},
		},
		{
			ID:      "pseudo",
			Default: "hide",
			Options: []APITopologyOption{
				{Value: "show", Label: "Show Uncontained", filter: nil, filterPseudo: false},
				{Value: "hide", Label: "Hide Uncontained", filter: render.IsNotPseudo, filterPseudo: true},
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
				{Value: "hide", Label: "Unconnected nodes hidden", filter: nil, filterPseudo: false},
			},
		},
	}

	// Topology option labels should tell the current state. The first item must
	// be the verb to get to that state
	registry.Add(
		APITopologyDesc{
			id:          processesID,
			renderer:    render.FilterUnconnected(render.ProcessWithContainerNameRenderer),
			Name:        "Processes",
			Rank:        1,
			Options:     unconnectedFilter,
			HideIfEmpty: true,
		},
		APITopologyDesc{
			id:          processesByNameID,
			parent:      processesID,
			renderer:    render.FilterUnconnected(render.ProcessNameRenderer),
			Name:        "by name",
			Options:     unconnectedFilter,
			HideIfEmpty: true,
		},
		APITopologyDesc{
			id:       containersID,
			renderer: render.ContainerWithImageNameRenderer,
			Name:     "Containers",
			Rank:     2,
			Options:  containerFilters,
		},
		APITopologyDesc{
			id:       containersByHostnameID,
			parent:   containersID,
			renderer: render.ContainerHostnameRenderer,
			Name:     "by DNS name",
			Options:  containerFilters,
		},
		APITopologyDesc{
			id:       containersByImageID,
			parent:   containersID,
			renderer: render.ContainerImageRenderer,
			Name:     "by image",
			Options:  containerFilters,
		},
		APITopologyDesc{
			id:          podsID,
			renderer:    render.PodRenderer,
			Name:        "Pods",
			Rank:        3,
			Options:     []APITopologyOptionGroup{unmanagedFilter},
			HideIfEmpty: true,
		},
		APITopologyDesc{
			id:          replicaSetsID,
			parent:      podsID,
			renderer:    render.ReplicaSetRenderer,
			Name:        "replica sets",
			Options:     []APITopologyOptionGroup{unmanagedFilter},
			HideIfEmpty: true,
		},
		APITopologyDesc{
			id:          deploymentsID,
			parent:      podsID,
			renderer:    render.DeploymentRenderer,
			Name:        "deployments",
			Options:     []APITopologyOptionGroup{unmanagedFilter},
			HideIfEmpty: true,
		},
		APITopologyDesc{
			id:          daemonsetsID,
			parent:      podsID,
			renderer:    render.DaemonSetRenderer,
			Name:        "daemonsets",
			Options:     []APITopologyOptionGroup{unmanagedFilter},
			HideIfEmpty: true,
		},
		APITopologyDesc{
			id:          servicesID,
			parent:      podsID,
			renderer:    render.PodServiceRenderer,
			Name:        "services",
			Options:     []APITopologyOptionGroup{unmanagedFilter},
			HideIfEmpty: true,
		},
		APITopologyDesc{
			id:          ecsTasksID,
			renderer:    render.ECSTaskRenderer,
			Name:        "Tasks",
			Rank:        3,
			Options:     []APITopologyOptionGroup{unmanagedFilter},
			HideIfEmpty: true,
		},
		APITopologyDesc{
			id:          ecsServicesID,
			parent:      ecsTasksID,
			renderer:    render.ECSServiceRenderer,
			Name:        "services",
			Options:     []APITopologyOptionGroup{unmanagedFilter},
			HideIfEmpty: true,
		},
		APITopologyDesc{
			id:          swarmServicesID,
			renderer:    render.SwarmServiceRenderer,
			Name:        "services",
			Rank:        3,
			Options:     []APITopologyOptionGroup{unmanagedFilter},
			HideIfEmpty: true,
		},
		APITopologyDesc{
			id:       hostsID,
			renderer: render.HostRenderer,
			Name:     "Hosts",
			Rank:     4,
		},
		APITopologyDesc{
			id:       weaveID,
			parent:   hostsID,
			renderer: render.WeaveRenderer,
			Name:     "Weave Net",
		},
	)

	return registry
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
	ID string `json:"id"`
	// Default value for the UI to adopt. NOT used as the default if the value is omitted, allowing "" as a distinct value.
	Default string              `json:"defaultValue,omitempty"`
	Options []APITopologyOption `json:"options,omitempty"`
	// SelectType describes how options can be picked. Currently defined values:
	//   "one": Default if empty. Exactly one option may be picked from the list.
	//   "union": Any number of options may be picked. Nodes matching any option filter selected are displayed.
	//           Value and Default should be a ","-separated list.
	SelectType string `json:"selectType,omitempty"`
	// For "union" type, this is the label the UI should use to represent the case where nothing is selected
	NoneLabel string `json:"noneLabel,omitempty"`
}

// Get the render filters to use for this option group as a Decorator, if any.
// If second arg is false, no decorator was needed.
func (g APITopologyOptionGroup) getFilterDecorator(value string) (render.Decorator, bool) {
	selectType := g.SelectType
	if selectType == "" {
		selectType = "one"
	}
	var values []string
	switch selectType {
	case "one":
		values = []string{value}
	case "union":
		values = strings.Split(value, ",")
	default:
		log.Errorf("Invalid select type %s for option group %s, ignoring option", selectType, g.ID)
		return nil, false
	}
	filters := []render.FilterFunc{}
	for _, opt := range g.Options {
		for _, v := range values {
			if v != opt.Value {
				continue
			}
			var filter render.FilterFunc
			if opt.filter == nil {
				// No filter means match everything (pseudo doesn't matter)
				filter = func(n report.Node) bool { return true }
			} else if opt.filterPseudo {
				// Apply filter to pseudo topologies also
				filter = opt.filter
			} else {
				// Allow all pseudo topology nodes, only apply filter to non-pseudo
				filter = render.AnyFilterFunc(render.IsPseudoTopology, opt.filter)
			}
			filters = append(filters, filter)
		}
	}
	if len(filters) == 0 {
		return nil, false
	}
	// Since we've encoded whether to ignore pseudo topologies into each subfilter,
	// we want no special behaviour for pseudo topologies here, which corresponds to MakePseudo
	return render.MakeFilterPseudoDecorator(render.AnyFilterFunc(filters...)), true
}

// APITopologyOption describes a &param=value to a given topology.
type APITopologyOption struct {
	Value string `json:"value"`
	Label string `json:"label"`

	filter       render.FilterFunc
	filterPseudo bool
}

type topologyStats struct {
	NodeCount          int `json:"node_count"`
	NonpseudoNodeCount int `json:"nonpseudo_node_count"`
	EdgeCount          int `json:"edge_count"`
	FilteredNodes      int `json:"filtered_nodes"`
}

// AddContainerFilters adds to the default Registry (topologyRegistry)'s containerFilters
func AddContainerFilters(newFilters ...APITopologyOption) {
	topologyRegistry.AddContainerFilters(newFilters...)
}

// AddContainerFilters adds container filters to this Registry
func (r *Registry) AddContainerFilters(newFilters ...APITopologyOption) {
	r.Lock()
	defer r.Unlock()
	for _, key := range []string{containersID, containersByHostnameID, containersByImageID} {
		for i := range r.items[key].Options {
			if r.items[key].Options[i].ID == systemGroupID {
				r.items[key].Options[i].Options = append(r.items[key].Options[i].Options, newFilters...)
				break
			}
		}
	}
}

// Add inserts a topologyDesc to the Registry's items map
func (r *Registry) Add(ts ...APITopologyDesc) {
	r.Lock()
	defer r.Unlock()
	for _, t := range ts {
		t.URL = apiTopologyURL + t.id

		if t.parent != "" {
			parent := r.items[t.parent]
			parent.SubTopologies = append(parent.SubTopologies, t)
			r.items[t.parent] = parent
		}

		r.items[t.id] = t
	}
}

func (r *Registry) get(name string) (APITopologyDesc, bool) {
	r.RLock()
	defer r.RUnlock()
	t, ok := r.items[name]
	return t, ok
}

func (r *Registry) walk(f func(APITopologyDesc)) {
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
func (r *Registry) makeTopologyList(rep Reporter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		report, err := rep.Report(ctx, time.Now())
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err)
			return
		}
		respondWith(w, http.StatusOK, r.renderTopologies(report, req))
	}
}

func (r *Registry) renderTopologies(rpt report.Report, req *http.Request) []APITopologyDesc {
	topologies := []APITopologyDesc{}
	req.ParseForm()
	r.walk(func(desc APITopologyDesc) {
		renderer, decorator, _ := r.RendererForTopology(desc.id, req.Form, rpt)
		desc.Stats = decorateWithStats(rpt, renderer, decorator)
		for i, sub := range desc.SubTopologies {
			renderer, decorator, _ := r.RendererForTopology(sub.id, req.Form, rpt)
			desc.SubTopologies[i].Stats = decorateWithStats(rpt, renderer, decorator)
		}
		topologies = append(topologies, desc)
	})
	return updateFilters(rpt, topologies)
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

// RendererForTopology ..
func (r *Registry) RendererForTopology(topologyID string, values url.Values, rpt report.Report) (render.Renderer, render.Decorator, error) {
	topology, ok := r.get(topologyID)
	if !ok {
		return nil, nil, fmt.Errorf("topology not found: %s", topologyID)
	}
	topology = updateFilters(rpt, []APITopologyDesc{topology})[0]

	if len(values) == 0 {
		// Do not apply filtering if no options where provided
		return topology.renderer, nil, nil
	}

	var decorators []render.Decorator
	for _, group := range topology.Options {
		value := values.Get(group.ID)
		if decorator, ok := group.getFilterDecorator(value); ok {
			decorators = append(decorators, decorator)
		}
	}
	if len(decorators) > 0 {
		// Here we tell the topology renderer to apply the filtering decorator
		// that we construct as a composition of all the selected filters.
		composedFilterDecorator := render.ComposeDecorators(decorators...)
		return render.ApplyDecorator(topology.renderer), composedFilterDecorator, nil
	}
	return topology.renderer, nil, nil
}

type reporterHandler func(context.Context, Reporter, http.ResponseWriter, *http.Request)

func captureReporter(rep Reporter, f reporterHandler) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		f(ctx, rep, w, r)
	}
}

type rendererHandler func(context.Context, render.Renderer, render.Decorator, report.Report, http.ResponseWriter, *http.Request)

func (r *Registry) captureRenderer(rep Reporter, f rendererHandler) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		topologyID := mux.Vars(req)["topology"]
		if _, ok := r.get(topologyID); !ok {
			http.NotFound(w, req)
			return
		}
		rpt, err := rep.Report(ctx, time.Now())
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err)
			return
		}
		req.ParseForm()
		renderer, decorator, err := r.RendererForTopology(topologyID, req.Form, rpt)
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err)
			return
		}
		f(ctx, renderer, decorator, rpt, w, req)
	}
}
