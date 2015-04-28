package main

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/alicebob/cello/report"
)

// Test /api/topology/processpid
func TestAPITopologyProcesspid(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}, nil))
	defer ts.Close()

	is404(t, ts, "/api/topology/processpid/foobar")

	{
		body := getRawJSON(t, ts, "/api/topology/processpid")
		var topo APITopology
		if err := json.Unmarshal(body, &topo); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		equals(t, 4, len(topo.Nodes))
		node, ok := topo.Nodes["pid:node-a.local:23128"]
		if !ok {
			t.Errorf("missing PID 23128 node")
		}
		equals(t, report.NewIDList("pid:node-b.local:215"), node.Adjacency)
		equals(t, report.NewIDList("hostA"), node.Origin)
		equals(t, "23128", node.LabelMajor)
		equals(t, "node-a.local (curl)", node.LabelMinor)
		equals(t, "23128", node.Rank)
		equals(t, false, node.Pseudo)
	}

	{
		// Node detail
		body := getRawJSON(t, ts, "/api/topology/processpid/pid:node-a.local:23128")
		var node APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		equals(t, node.Node.Adjacency, report.NewIDList("pid:node-b.local:215"))
		equals(t, node.Node.Aggregate, report.RenderableMetadata{
			"egress_bytes":       24,
			"ingress_bytes":      0,
			"max_conn_count_tcp": 401,
		})
	}

	{
		// Edge detail
		body := getRawJSON(t, ts, "/api/topology/processpid/pid:node-a.local:23128/pid:node-b.local:215")
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := report.RenderableMetadata{
			"egress_bytes":       24,
			"ingress_bytes":      0,
			"max_conn_count_tcp": 401,
		}
		if have := edge.Metadata; !reflect.DeepEqual(want, have) {
			t.Errorf("Edge metadata error. Want %#v, have %#v", want, have)
		}
	}
}

// Test /api/topology/processname
func TestAPITopologyProcessname(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}, nil))
	defer ts.Close()

	is404(t, ts, "/api/topology/processname/foobar")

	{
		body := getRawJSON(t, ts, "/api/topology/processname")
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
		body := getRawJSON(t, ts, "/api/topology/processname/proc:node-a.local:curl")
		var node APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		equals(t, node.Node.Adjacency, report.NewIDList("proc:node-b.local:apache"))
	}

	{
		// Edge detail
		body := getRawJSON(t, ts, "/api/topology/processname/proc:node-a.local:curl/proc:node-b.local:apache")
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := report.RenderableMetadata{
			"egress_bytes":       24,
			"ingress_bytes":      0,
			"max_conn_count_tcp": 401,
		}
		if !reflect.DeepEqual(want, edge.Metadata) {
			t.Errorf("Edge metadata error. Want %v, have %v", want, edge)
		}
	}
}

// Test /api/topology/networkip
func TestAPITopologyIP(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}, nil))
	defer ts.Close()

	is404(t, ts, "/api/topology/networkip/foobar")

	{
		body := getRawJSON(t, ts, "/api/topology/networkip")
		var topo APITopology
		if err := json.Unmarshal(body, &topo); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}

		equals(t, 3, len(topo.Nodes))
		node, ok := topo.Nodes["addr:;192.168.1.2"]
		if !ok {
			t.Errorf("missing 192.168.1.2 node")
		}
		equals(t, report.NewIDList("addr:;192.168.1.1"), node.Adjacency)
		equals(t, report.NewIDList("hostB"), node.Origin)
		equals(t, "192.168.1.2", node.LabelMajor)
		equals(t, "host-b", node.LabelMinor)
		equals(t, "192.168.1.2", node.Rank)
		equals(t, false, node.Pseudo)
	}

	{
		// Node detail
		body := getRawJSON(t, ts, "/api/topology/networkip/addr:;192.168.1.2")
		var node APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		equals(t, node.Node.Adjacency, report.NewIDList("addr:;192.168.1.1"))
	}

	{
		// Edge detail
		body := getRawJSON(t, ts, "/api/topology/networkip/addr:;192.168.1.2/addr:;192.168.1.1")
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := report.RenderableMetadata{
			"egress_bytes":       0,
			"ingress_bytes":      12,
			"max_conn_count_tcp": 16,
		}
		if !reflect.DeepEqual(want, edge.Metadata) {
			t.Errorf("Edge metadata error. Want %v, have %v", want, edge)
		}
	}

	{
		// Edge detail for the internet.
		body := getRawJSON(t, ts, "/api/topology/networkip/addr:;192.168.1.1/theinternet")
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := report.RenderableMetadata{
			"egress_bytes":       200,
			"ingress_bytes":      0,
			"max_conn_count_tcp": 15,
		}
		if !reflect.DeepEqual(want, edge.Metadata) {
			t.Errorf("Edge metadata error. Want %v, have %v", want, edge)
		}
	}
}

// Test /api/topology/networkhost
func TestAPITopologyNetwork(t *testing.T) {
	tpl, err := report.ThirdParty{
		Topology: "networkhost",
		Label:    "Exhibit A",
		URL:      "http://local.dev/showme.cgi?id={{ .Major }}",
	}.Compile("template_1")
	ok(t, err)
	ts := httptest.NewServer(
		Router(StaticReport{},
			report.ThirdPartyTemplates{tpl},
		))
	defer ts.Close()

	is404(t, ts, "/api/topology/networkhost/foobar")

	{
		body := getRawJSON(t, ts, "/api/topology/networkhost")
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
		body := getRawJSON(t, ts, "/api/topology/networkhost/host:host-b")
		var node APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		equals(t,
			report.NewIDList("host:host-a"),
			node.Node.Adjacency,
		)
		equals(t,
			[]report.ThirdParty{
				{Topology: "",
					Label: "Exhibit A",
					URL:   "http://local.dev/showme.cgi?id=host-b"},
			},
			node.Node.ThirdParty)
	}

	{
		// Edge detail
		body := getRawJSON(t, ts, "/api/topology/networkhost/host:host-b/host:host-a")
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := report.RenderableMetadata{
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
	ts := httptest.NewServer(Router(StaticReport{}, nil))
	defer ts.Close()

	url := "/api/topology/processpid/ws"

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
