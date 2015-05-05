package main

import (
	"net/http"
	"time"

	"github.com/weaveworks/scope/scope/report"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	websocketLoop    = 1 * time.Second
	websocketTimeout = 10 * time.Second
)

// APITopology is returned by the /api/topology/* handlers.
type APITopology struct {
	Nodes map[string]report.RenderableNode `json:"nodes"`
}

// APINode is returned by the /api/topology/*/* handlers.
type APINode struct {
	Node report.DetailedRenderableNode `json:"node"`
}

// APIEdge is returned by the /api/topology/*/*/* handlers.
type APIEdge struct {
	Metadata report.RenderableMetadata `json:"metadata"`
}

// topologySelecter selects a single topology from a report.
type topologySelecter func(r report.Report) report.Topology

func selectProcess(r report.Report) report.Topology {
	return r.Process
}

func selectNetwork(r report.Report) report.Topology {
	return r.Network
}

// makeTopologyHandlers make /api/topology/* handlers.
func makeTopologyHandlers(
	rep Reporter,
	topo topologySelecter,
	mapping report.MapFunc,
	grouped bool,
	get *mux.Router,
	base string,
	tps report.ThirdPartyTemplates,
) {
	// Full topology:
	get.HandleFunc(base, func(w http.ResponseWriter, r *http.Request) {
		rpt := rep.Report()
		rendered := topo(rpt).RenderBy(mapping, grouped, nil)
		t := APITopology{
			Nodes: report.Downcast(rendered),
		}
		respondWith(w, http.StatusOK, t)
	})

	// Websocket for the full topology:
	// This route overlaps with the next one.
	get.HandleFunc(base+"/ws", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		loop := websocketLoop
		if t := r.Form.Get("t"); t != "" {
			var err error
			if loop, err = time.ParseDuration(t); err != nil {
				respondWith(w, http.StatusBadRequest, t)
				return
			}
		}
		handleWebsocket(w, r, rep, topo, mapping, grouped, loop)
	})

	// Individual nodes:
	get.HandleFunc(base+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		var (
			vars     = mux.Vars(r)
			nodeID   = vars["id"]
			rpt      = rep.Report()
			rendered = topo(rpt).RenderBy(mapping, grouped, tps)
			node, ok = rendered[nodeID]
		)
		if !ok {
			http.NotFound(w, r)
			return
		}

		respondWith(w, http.StatusOK, APINode{
			Node: node,
		})
	})

	// Individual edges:
	get.HandleFunc(base+"/{local}/{remote}", func(w http.ResponseWriter, r *http.Request) {
		var (
			vars     = mux.Vars(r)
			localID  = vars["local"]
			remoteID = vars["remote"]
			rpt      = rep.Report()
			metadata = topo(rpt).EdgeMetadata(mapping, grouped, localID, remoteID).Render()
		)
		respondWith(w, http.StatusOK, APIEdge{
			Metadata: metadata,
		})
	})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWebsocket(
	w http.ResponseWriter,
	r *http.Request,
	rep Reporter,
	topo topologySelecter,
	mapping report.MapFunc,
	grouped bool,
	loop time.Duration,
) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// log.Println("Upgrade:", err)
		return
	}
	defer conn.Close()

	quit := make(chan struct{})
	go func(c *websocket.Conn) {
		// Discard all the browser sends us.
		for {
			if _, _, err := c.NextReader(); err != nil {
				close(quit)
				break
			}
		}
	}(conn)

	var previousTopo map[string]report.DetailedRenderableNode
	for {
		newTopo := topo(rep.Report()).RenderBy(mapping, grouped, nil)
		diff := report.TopoDiff(previousTopo, newTopo)
		previousTopo = newTopo

		conn.SetWriteDeadline(time.Now().Add(websocketTimeout))
		if err := conn.WriteJSON(diff); err != nil {
			return
		}

		select {
		case <-quit:
			return
		case <-time.After(loop):
		}
	}
}
