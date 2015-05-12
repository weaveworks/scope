package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/weaveworks/scope/report"
)

const (
	websocketLoop    = 1 * time.Second
	websocketTimeout = 10 * time.Second
)

// APITopology is returned by the /api/topology/{name} handler.
type APITopology struct {
	Nodes map[string]report.RenderableNode `json:"nodes"`
}

// APINode is returned by the /api/topology/{name}/{id} handler.
type APINode struct {
	Node report.DetailedNode `json:"node"`
}

// APIEdge is returned by the /api/topology/*/*/* handlers.
type APIEdge struct {
	Metadata report.AggregateMetadata `json:"metadata"`
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
) {
	// Full topology.
	get.HandleFunc(base, func(w http.ResponseWriter, r *http.Request) {
		rpt := rep.Report()
		rendered := topo(rpt).RenderBy(mapping, grouped)
		t := APITopology{
			Nodes: rendered,
		}
		respondWith(w, http.StatusOK, t)
	})

	// Websocket for the full topology. This route overlaps with the next.
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

	// Individual nodes.
	get.HandleFunc(base+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		var (
			vars     = mux.Vars(r)
			nodeID   = vars["id"]
			rpt      = rep.Report()
			node, ok = topo(rpt).RenderBy(mapping, grouped)[nodeID]
		)
		if !ok {
			http.NotFound(w, r)
			return
		}
		of := func(nodeID string) (Origin, bool) { return getOrigin(rep, nodeID) }
		respondWith(w, http.StatusOK, APINode{Node: makeDetailed(node, of)})
	})

	// Individual edges.
	get.HandleFunc(base+"/{local}/{remote}", func(w http.ResponseWriter, r *http.Request) {
		var (
			vars     = mux.Vars(r)
			localID  = vars["local"]
			remoteID = vars["remote"]
			rpt      = rep.Report()
			metadata = topo(rpt).EdgeMetadata(mapping, grouped, localID, remoteID).Transform()
		)
		respondWith(w, http.StatusOK, APIEdge{Metadata: metadata})
	})
}

// TODO(pb): temporary hack
func makeDetailed(n report.RenderableNode, of func(string) (Origin, bool)) report.DetailedNode {
	// A RenderableNode may be the result of merge operation(s), and so may
	// have multiple origins.
	origins := []report.Table{}
	for _, originID := range n.Origin {
		origin, ok := of(originID)
		if !ok {
			origin = unknownOrigin
		}
		origins = append(origins, report.Table{
			Title:   "Origin",
			Numeric: false,
			Rows: []report.Row{
				{"Hostname", origin.Hostname, ""},
				{"Load", fmt.Sprintf("%.2f %.2f %.2f", origin.LoadOne, origin.LoadFive, origin.LoadFifteen), ""},
				{"OS", origin.OS, ""},
				//{"Addresses", strings.Join(origin.Addresses, ", "), ""},
				{"ID", originID, ""},
			},
		})
	}

	tables := []report.Table{}
	tables = append(tables, report.Table{
		Title:   "Connections",
		Numeric: true,
		Rows: []report.Row{
			{"TCP (max)", strconv.FormatInt(int64(n.Metadata[report.KeyMaxConnCountTCP]), 10), ""},
		},
	})
	tables = append(tables, origins...)

	return report.DetailedNode{
		ID:         n.ID,
		LabelMajor: n.LabelMajor,
		LabelMinor: n.LabelMinor,
		Pseudo:     n.Pseudo,
		Tables:     tables,
	}
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

	var previousTopo map[string]report.RenderableNode
	for {
		newTopo := topo(rep.Report()).RenderBy(mapping, grouped)
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

var unknownOrigin = Origin{
	Hostname:    "unknown",
	OS:          "unknown",
	Addresses:   []string{},
	LoadOne:     0.0,
	LoadFive:    0.0,
	LoadFifteen: 0.0,
}
