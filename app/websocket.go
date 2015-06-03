package main

import (
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/websocket"

	"github.com/weaveworks/scope/report"
)

const (
	websocketInterval = 1 * time.Second
	websocketTimeout  = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWebsocket(w http.ResponseWriter, r *http.Request, reporter Reporter, definition topologyDefinition, d time.Duration) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// log.Println("Upgrade:", err)
		return
	}
	defer conn.Close()

	quit := make(chan struct{})
	go func(c *websocket.Conn) {
		for {
			// just discard everything the browser sends
			if _, _, err := c.NextReader(); err != nil {
				close(quit)
				break
			}
		}
	}(conn)

	var (
		prevTopology map[string]report.RenderableNode
		tick         = time.Tick(d)
	)
	for {
		currTopology := report.Render(reporter.Report(), definition.selector, definition.mapper, definition.pseudo)
		diff := topologyDiff(prevTopology, currTopology)
		prevTopology = currTopology

		if err := conn.SetWriteDeadline(time.Now().Add(websocketTimeout)); err != nil {
			return
		}
		if err := conn.WriteJSON(diff); err != nil {
			return
		}

		select {
		case <-tick:
		case <-quit:
			return
		}
	}
}

// diff is returned by topologyDiff. It represents the changes between two
// renderable topologies.
type diff struct {
	Add    []report.RenderableNode `json:"add"`
	Update []report.RenderableNode `json:"update"`
	Remove []string                `json:"remove"`
}

// topologyDiff gives you the diff to get from A to B.
func topologyDiff(a, b map[string]report.RenderableNode) diff {
	d := diff{}
	notSeen := map[string]struct{}{}
	for k := range a {
		notSeen[k] = struct{}{}
	}

	for k, node := range b {
		if _, ok := a[k]; !ok {
			d.Add = append(d.Add, node)
		} else if !reflect.DeepEqual(node, a[k]) {
			d.Update = append(d.Update, node)
		}
		delete(notSeen, k)
	}

	for k := range notSeen {
		d.Remove = append(d.Remove, k)
	}

	return d
}
