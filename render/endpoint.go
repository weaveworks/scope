package render

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/weaveworks/scope/report"
)

// Pseudo is the topology for nodes that aren't "real" nodes inside a
// cluster, such as nodes representing the internet, external
// services, and artificial grouping such as "uncontained processes"
// and "unmanaged containers".
const Pseudo = "pseudo"

// EndpointRenderer is a Renderer which produces a renderable endpoint graph.
var EndpointRenderer = SelectEndpoint

type endpointMapFunc func(report.Node) string

type mapEndpoints struct {
	f        endpointMapFunc
	topology string
}

// MapEndpoints creates a renderer for the endpoint topology. Each
// endpoint is either turned into a pseudo node, or mapped to a node
// in the specified topology by the supplied function.
func MapEndpoints(f endpointMapFunc, topology string) Renderer {
	return mapEndpoints{f: f, topology: topology}
}

func (e mapEndpoints) Render(ctx context.Context, rpt report.Report) Nodes {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mapEndpoints.Render")
	defer span.Finish()
	local := LocalNetworks(rpt)
	endpoints := SelectEndpoint.Render(ctx, rpt)
	ret := newJoinResults(TopologySelector(e.topology).Render(ctx, rpt).Nodes)

	for _, n := range endpoints.Nodes {
		// Nodes without a hostid are mapped to pseudo nodes, if
		// possible.
		if _, ok := n.Latest.Lookup(report.HostNodeID); !ok {
			if id, ok := pseudoNodeID(rpt, n, local); ok {
				ret.addChild(n, id, Pseudo)
				continue
			}
		}
		if id := e.f(n); id != "" {
			ret.addChild(n, id, e.topology)
		}
	}
	span.LogFields(otlog.Int("input.nodes", len(endpoints.Nodes)),
		otlog.Int("ouput.nodes", len(ret.nodes)))
	return ret.result(ctx, endpoints)
}
