package app_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test/fixture"
)

func TestAll(t *testing.T) {
	ts := topologyServer()
	defer ts.Close()

	body := getRawJSON(t, ts, "/api/topology")
	var topologies []app.APITopologyDesc
	decoder := codec.NewDecoderBytes(body, &codec.JsonHandle{})
	if err := decoder.Decode(&topologies); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}

	getTopology := func(topologyURL string) {
		body := getRawJSON(t, ts, topologyURL)
		var topology app.APITopology
		decoder := codec.NewDecoderBytes(body, &codec.JsonHandle{})
		if err := decoder.Decode(&topology); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}

		for _, node := range topology.Nodes {
			body := getRawJSON(t, ts, fmt.Sprintf("%s/%s", topologyURL, url.QueryEscape(node.ID)))
			var node app.APINode
			decoder = codec.NewDecoderBytes(body, &codec.JsonHandle{})
			if err := decoder.Decode(&node); err != nil {
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

func TestAPITopologyProcesses(t *testing.T) {
	ts := topologyServer()
	defer ts.Close()
	is404(t, ts, "/api/topology/processes/foobar")
	{
		body := getRawJSON(t, ts, "/api/topology/processes/"+fixture.ServerProcessNodeID)
		var node app.APINode
		decoder := codec.NewDecoderBytes(body, &codec.JsonHandle{})
		if err := decoder.Decode(&node); err != nil {
			t.Fatal(err)
		}
		equals(t, fixture.ServerProcessNodeID, node.Node.ID)
		equals(t, "apache", node.Node.Label)
		equals(t, false, node.Node.Pseudo)
		// Let's not unit-test the specific content of the detail tables
	}

	{
		body := getRawJSON(t, ts, "/api/topology/processes-by-name/"+
			url.QueryEscape(fixture.Client1Name))
		var node app.APINode
		decoder := codec.NewDecoderBytes(body, &codec.JsonHandle{})
		if err := decoder.Decode(&node); err != nil {
			t.Fatal(err)
		}
		equals(t, fixture.Client1Name, node.Node.ID)
		equals(t, fixture.Client1Name, node.Node.Label)
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
		decoder := codec.NewDecoderBytes(body, &codec.JsonHandle{})
		if err := decoder.Decode(&topo); err != nil {
			t.Fatal(err)
		}

		// Should have the rendered host nodes
		for id := range expected.RenderedHosts {
			if _, ok := topo.Nodes[id]; !ok {
				t.Errorf("Expected output to include node: %s, but wasn't found", id)
			}
		}
	}
	{
		body := getRawJSON(t, ts, "/api/topology/hosts/"+fixture.ServerHostNodeID)
		var node app.APINode
		decoder := codec.NewDecoderBytes(body, &codec.JsonHandle{})
		if err := decoder.Decode(&node); err != nil {
			t.Fatal(err)
		}
		equals(t, fixture.ServerHostNodeID, node.Node.ID)
		equals(t, "server", node.Node.Label)
		equals(t, false, node.Node.Pseudo)
		// Let's not unit-test the specific content of the detail tables
	}
}

// Basic websocket test
func TestAPITopologyWebsocket(t *testing.T) {
	ts := topologyServer()
	defer ts.Close()
	url := "/api/topology/processes/ws"

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
	var d detailed.Diff
	decoder := codec.NewDecoderBytes(p, &codec.JsonHandle{})
	if err := decoder.Decode(&d); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
	equals(t, 6, len(d.Add))
	equals(t, 0, len(d.Update))
	equals(t, 0, len(d.Remove))
}

func newu64(value uint64) *uint64 { return &value }
