package main

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/weaveworks/scope/report"
)

func TestAPITopologyApplications(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}))
	defer ts.Close()

	is404(t, ts, "/api/topology/applications/foobar")

	{
		body := getRawJSON(t, ts, "/api/topology/applications")
		var topo APITopology
		if err := json.Unmarshal(body, &topo); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		equals(t, 4, len(topo.Nodes))
		node, ok := topo.Nodes["proc:node-a.local:curl"]
		if !ok {
			t.Errorf("missing curl node")
		}
		equals(t, report.NewIDList("proc:node-b.local:apache"), node.Adjacency)
		equals(t, report.NewIDList("hostA"), node.Origin)
		equals(t, "curl", node.LabelMajor)
		equals(t, "node-a.local", node.LabelMinor)
		equals(t, "curl", node.Rank)
		equals(t, false, node.Pseudo)
	}

	{
		// Node detail
		body := getRawJSON(t, ts, "/api/topology/applications/proc:node-a.local:curl")
		var node APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		// TODO(pb): replace
	}

	{
		// Edge detail
		body := getRawJSON(t, ts, "/api/topology/applications/proc:node-a.local:curl/proc:node-b.local:apache")
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := report.AggregateMetadata{
			"egress_bytes":       24,
			"ingress_bytes":      0,
			"max_conn_count_tcp": 401,
		}
		if !reflect.DeepEqual(want, edge.Metadata) {
			t.Errorf("Edge metadata error. Want %v, have %v", want, edge)
		}
	}
}

func TestAPITopologyHosts(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}))
	defer ts.Close()

	is404(t, ts, "/api/topology/hosts/foobar")

	{
		body := getRawJSON(t, ts, "/api/topology/hosts")
		var topo APITopology
		if err := json.Unmarshal(body, &topo); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}

		equals(t, 3, len(topo.Nodes))
		node, ok := topo.Nodes["host:host-b"]
		if !ok {
			t.Errorf("missing host:host-b node")
		}
		equals(t, report.NewIDList("host:host-a"), node.Adjacency)
		equals(t, report.NewIDList("hostB"), node.Origin)
		equals(t, "host-b", node.LabelMajor)
		equals(t, "", node.LabelMinor)
		equals(t, "host-b", node.Rank)
		equals(t, false, node.Pseudo)
	}

	{
		// Node detail
		body := getRawJSON(t, ts, "/api/topology/hosts/host:host-b")
		var node APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		// TODO(pb): replace
	}

	{
		// Edge detail
		body := getRawJSON(t, ts, "/api/topology/hosts/host:host-b/host:host-a")
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := report.AggregateMetadata{
			"egress_bytes":       0,
			"ingress_bytes":      12,
			"max_conn_count_tcp": 16,
		}
		if !reflect.DeepEqual(want, edge.Metadata) {
			t.Errorf("Edge metadata error. Want %v, have %v", want, edge)
		}
	}
}

// Basic websocket test
func TestAPITopologyWebsocket(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}))
	defer ts.Close()

	url := "/api/topology/applications/ws"

	// Not a websocket request:
	res, _ := checkGet(t, ts, url)
	if have := res.StatusCode; have != 400 {
		t.Fatalf("Expected status %d, got %d.", 400, have)
	}

	// Proper websocket request:
	ts.URL = "ws" + ts.URL[len("http"):]
	dialer := &websocket.Dialer{}
	ws, res, err := dialer.Dial(ts.URL+url, nil)
	ok(t, err)
	defer ws.Close()

	if have := res.StatusCode; have != 101 {
		t.Fatalf("Expected status %d, got %d.", 101, have)
	}

	_, p, err := ws.ReadMessage()
	ok(t, err)
	var d report.Diff
	if err := json.Unmarshal(p, &d); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
	equals(t, 4, len(d.Add))
	equals(t, 0, len(d.Update))
	equals(t, 0, len(d.Remove))
}
