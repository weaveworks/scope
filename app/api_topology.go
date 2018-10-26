package app

import (
	"net/http"
	"time"

	"context"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
)

const (
	websocketLoop = 1 * time.Second
)

// APITopology is returned by the /api/topology/{name} handler.
type APITopology struct {
	Nodes detailed.NodeSummaries `json:"nodes"`
}

// APINode is returned by the /api/topology/{name}/{id} handler.
type APINode struct {
	Node detailed.Node `json:"node"`
}

// RenderContextForReporter creates the rendering context for the given reporter.
func RenderContextForReporter(rep Reporter, r report.Report) detailed.RenderContext {
	rc := detailed.RenderContext{Report: r}
	if wrep, ok := rep.(WebReporter); ok {
		rc.MetricsGraphURL = wrep.MetricsGraphURL
	}
	return rc
}

type rendererHandler func(context.Context, render.Renderer, render.Transformer, detailed.RenderContext, http.ResponseWriter, *http.Request)

// Full topology.
func handleTopology(ctx context.Context, renderer render.Renderer, transformer render.Transformer, rc detailed.RenderContext, w http.ResponseWriter, r *http.Request) {
	respondWith(w, http.StatusOK, APITopology{
		Nodes: detailed.Summaries(ctx, rc, render.Render(ctx, rc.Report, renderer, transformer).Nodes),
	})
}

// Individual nodes.
func handleNode(ctx context.Context, renderer render.Renderer, transformer render.Transformer, rc detailed.RenderContext, w http.ResponseWriter, r *http.Request) {
	var (
		vars       = mux.Vars(r)
		topologyID = vars["topology"]
		nodeID     = vars["id"]
	)
	// We must not lose the node during filtering. We achieve that by
	// (1) rendering the report with the base renderer, without
	// filtering, which gives us the node (if it exists at all), and
	// then (2) applying the filter separately to that result.  If the
	// node is lost in the second step, we simply put it back.
	nodes := renderer.Render(ctx, rc.Report)
	node, ok := nodes.Nodes[nodeID]
	if !ok {
		http.NotFound(w, r)
		return
	}
	nodes = transformer.Transform(nodes)
	if filteredNode, ok := nodes.Nodes[nodeID]; ok {
		node = filteredNode
	} else { // we've lost the node during filtering; put it back
		nodes.Nodes[nodeID] = node
		nodes.Filtered--
	}
	respondWith(w, http.StatusOK, APINode{Node: detailed.MakeNode(topologyID, rc, nodes.Nodes, node)})
}

// Websocket for the full topology.
func handleWebsocket(
	ctx context.Context,
	rep Reporter,
	w http.ResponseWriter,
	r *http.Request,
) {
	if err := r.ParseForm(); err != nil {
		respondWith(w, http.StatusInternalServerError, err)
		return
	}
	loop := websocketLoop
	if t := r.Form.Get("t"); t != "" {
		var err error
		if loop, err = time.ParseDuration(t); err != nil {
			respondWith(w, http.StatusBadRequest, t)
			return
		}
	}

	conn, err := xfer.Upgrade(w, r, nil)
	if err != nil {
		// log.Info("Upgrade:", err)
		return
	}
	defer conn.Close()

	quit := make(chan struct{})
	go func(c xfer.Websocket) {
		for { // just discard everything the browser sends
			if _, _, err := c.ReadMessage(); err != nil {
				if !xfer.IsExpectedWSCloseError(err) {
					log.Error("err:", err)
				}
				close(quit)
				break
			}
		}
	}(conn)

	var (
		previousTopo     detailed.NodeSummaries
		tick             = time.Tick(loop)
		wait             = make(chan struct{}, 1)
		topologyID       = mux.Vars(r)["topology"]
		startReportingAt = deserializeTimestamp(r.Form.Get("timestamp"))
		channelOpenedAt  = time.Now()
	)

	rep.WaitOn(ctx, wait)
	defer rep.UnWait(ctx, wait)

	for {
		// We measure how much time has passed since the channel was opened
		// and add it to the initial report timestamp to get the timestamp
		// of the snapshot we want to report right now.
		// NOTE: Multiplying `timestampDelta` by a constant factor here
		// would have an effect of fast-forward, which is something we
		// might be interested in implementing in the future.
		timestampDelta := time.Since(channelOpenedAt)
		reportTimestamp := startReportingAt.Add(timestampDelta)
		re, err := rep.Report(ctx, reportTimestamp)
		if err != nil {
			log.Errorf("Error generating report: %v", err)
			return
		}
		renderer, filter, err := topologyRegistry.RendererForTopology(topologyID, r.Form, re)
		if err != nil {
			log.Errorf("Error generating report: %v", err)
			return
		}
		newTopo := detailed.Summaries(ctx, RenderContextForReporter(rep, re), render.Render(ctx, re, renderer, filter).Nodes)
		diff := detailed.TopoDiff(previousTopo, newTopo)
		previousTopo = newTopo

		if err := conn.WriteJSON(diff); err != nil {
			if !xfer.IsExpectedWSCloseError(err) {
				log.Errorf("cannot serialize topology diff: %s", err)
			}
			return
		}

		select {
		case <-wait:
		case <-tick:
		case <-quit:
			return
		}
	}
}
