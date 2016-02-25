package app

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/detailed"
)

const (
	websocketLoop = 1 * time.Second
)

// APITopology is returned by the /api/topology/{name} handler.
type APITopology struct {
	Nodes render.RenderableNodes `json:"nodes"`
}

// APINode is returned by the /api/topology/{name}/{id} handler.
type APINode struct {
	Node detailed.Node `json:"node"`
}

// Full topology.
func handleTopology(ctx context.Context, rep Reporter, renderer render.Renderer, w http.ResponseWriter, r *http.Request) {
	respondWith(w, http.StatusOK, APITopology{
		Nodes: renderer.Render(rep.Report(ctx)).Prune(),
	})
}

// Websocket for the full topology. This route overlaps with the next.
func handleWs(ctx context.Context, rep Reporter, renderer render.Renderer, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		respondWith(w, http.StatusInternalServerError, err.Error())
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
	handleWebsocket(ctx, w, r, rep, renderer, loop)
}

// Individual nodes.
func handleNode(nodeID string) func(context.Context, Reporter, render.Renderer, http.ResponseWriter, *http.Request) {
	return func(ctx context.Context, rep Reporter, renderer render.Renderer, w http.ResponseWriter, r *http.Request) {
		var (
			rpt      = rep.Report(ctx)
			node, ok = renderer.Render(rep.Report(ctx))[nodeID]
		)
		if !ok {
			http.NotFound(w, r)
			return
		}
		respondWith(w, http.StatusOK, APINode{Node: detailed.MakeNode(rpt, node)})
	}
}

func handleWebsocket(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	rep Reporter,
	renderer render.Renderer,
	loop time.Duration,
) {
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
		previousTopo render.RenderableNodes
		tick         = time.Tick(loop)
		wait         = make(chan struct{}, 1)
	)
	rep.WaitOn(ctx, wait)
	defer rep.UnWait(ctx, wait)

	for {
		newTopo := renderer.Render(rep.Report(ctx)).Prune()
		diff := render.TopoDiff(previousTopo, newTopo)
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
