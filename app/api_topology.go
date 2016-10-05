package app

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

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

// Full topology.
func handleTopology(ctx context.Context, renderer render.Renderer, decorator render.Decorator, report report.Report, w http.ResponseWriter, r *http.Request) {
	respondWith(w, http.StatusOK, APITopology{
		Nodes: detailed.Summaries(report, renderer.Render(report, decorator)),
	})
}

// Individual nodes.
func handleNode(ctx context.Context, renderer render.Renderer, decorator render.Decorator, report report.Report, w http.ResponseWriter, r *http.Request) {
	var (
		vars             = mux.Vars(r)
		topologyID       = vars["topology"]
		nodeID           = vars["id"]
		preciousRenderer = render.PreciousNodeRenderer{nodeID, renderer}
		rendered         = preciousRenderer.Render(report, decorator)
		node, ok         = rendered[nodeID]
	)
	if !ok {
		http.NotFound(w, r)
		return
	}
	respondWith(w, http.StatusOK, APINode{Node: detailed.MakeNode(topologyID, report, rendered, node)})
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
					log.Println("err:", err)
				}
				close(quit)
				break
			}
		}
	}(conn)

	var (
		previousTopo detailed.NodeSummaries
		tick         = time.Tick(loop)
		wait         = make(chan struct{}, 1)
		topologyID   = mux.Vars(r)["topology"]
	)
	rep.WaitOn(ctx, wait)
	defer rep.UnWait(ctx, wait)

	for {
		report, err := rep.Report(ctx)
		if err != nil {
			log.Errorf("Error generating report: %v", err)
			return
		}
		renderer, decorator, err := topologyRegistry.rendererForTopology(topologyID, r.Form, report)
		if err != nil {
			log.Errorf("Error generating report: %v", err)
			return
		}
		newTopo := detailed.Summaries(report, renderer.Render(report, decorator))
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
