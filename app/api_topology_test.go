package app_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
)

func TestAll(t *testing.T) {
	ts := topologyServer()
	defer ts.Close()

	body := getRawJSON(t, ts, "/api/topology")
	var topologies []app.APITopologyDesc
	if err := json.Unmarshal(body, &topologies); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}

	getTopology := func(topologyURL string) {
		body := getRawJSON(t, ts, topologyURL)
		var topology app.APITopology
		if err := json.Unmarshal(body, &topology); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}

		for _, node := range topology.Nodes {
			body := getRawJSON(t, ts, fmt.Sprintf("%s/%s", topologyURL, url.QueryEscape(node.ID)))
			var node app.APINode
			if err := json.Unmarshal(body, &node); err != nil {
				t.Fatalf("JSON parse error: %s", err)
			}
		}
	}

	for _, topology := range topologies {
		getTopology(topology.URL)

		for _, subTopology := range topology.SubTopologies {
			getTopology(subTopology.URL)
		}
	}
}

func TestAPITopologyContainers(t *testing.T) {
	ts := topologyServer()
	{
		body := getRawJSON(t, ts, "/api/topology/containers")
		var topo app.APITopology
		if err := json.Unmarshal(body, &topo); err != nil {
			t.Fatal(err)
		}
		want := expected.RenderedContainers.Copy()
		for id, node := range want {
			node.ControlNode = ""
			want[id] = node
		}

		if have := topo.Nodes.Prune(); !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}

func TestAPITopologyApplications(t *testing.T) {
	ts := topologyServer()
	defer ts.Close()
	is404(t, ts, "/api/topology/applications/foobar")
	{
		body := getRawJSON(t, ts, "/api/topology/applications/"+expected.ServerProcessID)
		var node app.APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatal(err)
		}
		equals(t, expected.ServerProcessID, node.Node.ID)
		equals(t, "apache", node.Node.LabelMajor)
		equals(t, fmt.Sprintf("%s (server:%s)", fixture.ServerHostID, fixture.ServerPID), node.Node.LabelMinor)
		equals(t, false, node.Node.Pseudo)
		// Let's not unit-test the specific content of the detail tables
	}

	{
		body := getRawJSON(t, ts, "/api/topology/applications-by-name/"+
			url.QueryEscape(fixture.Client1Name))
		var node app.APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatal(err)
		}
		equals(t, fixture.Client1Name, node.Node.ID)
		equals(t, fixture.Client1Name, node.Node.LabelMajor)
		equals(t, "2 processes", node.Node.LabelMinor)
		equals(t, false, node.Node.Pseudo)
		// Let's not unit-test the specific content of the detail tables
	}
}

func TestAPITopologyHosts(t *testing.T) {
	ts := topologyServer()
	defer ts.Close()
	is404(t, ts, "/api/topology/hosts/foobar")
	{
		body := getRawJSON(t, ts, "/api/topology/hosts")
		var topo app.APITopology
		if err := json.Unmarshal(body, &topo); err != nil {
			t.Fatal(err)
		}

		if want, have := expected.RenderedHosts, topo.Nodes.Prune(); !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
	{
		body := getRawJSON(t, ts, "/api/topology/hosts/"+expected.ServerHostRenderedID)
		var node app.APINode
		if err := json.Unmarshal(body, &node); err != nil {
			t.Fatal(err)
		}
		equals(t, expected.ServerHostRenderedID, node.Node.ID)
		equals(t, "server", node.Node.LabelMajor)
		equals(t, "hostname.com", node.Node.LabelMinor)
		equals(t, false, node.Node.Pseudo)
		// Let's not unit-test the specific content of the detail tables
	}
}

// Basic websocket test
func TestAPITopologyWebsocket(t *testing.T) {
	ts := topologyServer()
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
	equals(t, 7, len(d.Add))
	equals(t, 0, len(d.Update))
	equals(t, 0, len(d.Remove))
}

func newu64(value uint64) *uint64 { return &value }
