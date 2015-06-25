package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test"
)

func TestAPITopologyApplications(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}))
	defer ts.Close()
	is404(t, ts, "/api/topology/applications/foobar")
	{
		body := getRawJSON(t, ts, "/api/topology/applications")
		var topo APITopology
		if err := json.Unmarshal(body, &topo); err != nil {
			t.Fatal(err)
		}

		want := render.OnlyConnected(expected.RenderedProcesses)
		if !reflect.DeepEqual(want, topo.Nodes) {
			t.Error("\n" + test.Diff(want, topo.Nodes))
		}
	}
	{
		body := getRawJSON(t, ts, "/api/topology/applications/"+expected.ServerProcessID)
		var node APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatal(err)
		}
		equals(t, expected.ServerProcessID, node.Node.ID)
		equals(t, "apache", node.Node.LabelMajor)
		equals(t, fmt.Sprintf("%s (%s)", test.ServerHostID, test.ServerPID), node.Node.LabelMinor)
		equals(t, false, node.Node.Pseudo)
		// Let's not unit-test the specific content of the detail tables
	}
	{
		body := getRawJSON(t, ts, fmt.Sprintf("/api/topology/applications/%s/%s", expected.ClientProcess1ID, expected.ServerProcessID))
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := render.AggregateMetadata{
			"egress_bytes":  10,
			"ingress_bytes": 100,
		}
		if !reflect.DeepEqual(want, edge.Metadata) {
			t.Error("\n" + test.Diff(want, edge.Metadata))
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
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected.RenderedHosts, topo.Nodes) {
			t.Error("\n" + test.Diff(expected.RenderedHosts, topo.Nodes))
		}
	}
	{
		body := getRawJSON(t, ts, "/api/topology/hosts/"+expected.ServerHostRenderedID)
		var node APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatal(err)
		}
		equals(t, expected.ServerHostRenderedID, node.Node.ID)
		equals(t, "server", node.Node.LabelMajor)
		equals(t, "hostname.com", node.Node.LabelMinor)
		equals(t, false, node.Node.Pseudo)
		// Let's not unit-test the specific content of the detail tables
	}
	{
		body := getRawJSON(t, ts, fmt.Sprintf("/api/topology/hosts/%s/%s", expected.ServerHostRenderedID, expected.ClientHostRenderedID))
		var edge APIEdge
		if err := json.Unmarshal(body, &edge); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		want := render.AggregateMetadata{
			"max_conn_count_tcp": 3,
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

	// Not a websocket request
	res, _ := checkGet(t, ts, url)
	if have := res.StatusCode; have != 400 {
		t.Fatalf("Expected status %d, got %d.", 400, have)
	}

	// Proper websocket request
	ts.URL = "ws" + ts.URL[len("http"):]
	dialer := &websocket.Dialer{}
	ws, res, err := dialer.Dial(ts.URL+url, nil)
	ok(t, err)
	defer ws.Close()

	if want, have := 101, res.StatusCode; want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	_, p, err := ws.ReadMessage()
	ok(t, err)
	var d render.Diff
	if err := json.Unmarshal(p, &d); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
	equals(t, 6, len(d.Add))
	equals(t, 0, len(d.Update))
	equals(t, 0, len(d.Remove))
}
