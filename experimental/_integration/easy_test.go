package integration_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/scope/report"
)

func TestComponentsAreAvailable(t *testing.T) {
	pause := time.Millisecond
	for _, c := range []string{
		fmt.Sprintf(`app -http.address=:%d`, appPort),
		fmt.Sprintf(`bridge -listen=:%d`, bridgePort),
		fmt.Sprintf(`fixprobe -listen=:%d`, probePort1),
		fmt.Sprintf(`demoprobe -listen=:%d`, probePort1),
	} {
		cmd := start(t, c)
		time.Sleep(pause)
		stop(t, cmd)
		t.Logf("%s: OK", filepath.Base(cmd.Path))
	}
}

func TestApplications(t *testing.T) {
	withContext(t, oneProbe, func() {
		topo := parseTopology(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/applications", appPort)))
		assertAdjacent(t, topo["proc:node-1.2.3.4:apache"], "theinternet", "proc:node-192.168.1.1:wget")
		want := map[string]interface{}{"max_conn_count_tcp": float64(19)}
		have := parseEdge(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/applications/%s/%s", appPort, "proc:node-192.168.1.1:wget", "theinternet")))
		if !reflect.DeepEqual(have, want) {
			t.Errorf("have: %#v, want %#v", have, want)
		}
	})
}

func TestHosts(t *testing.T) {
	withContext(t, oneProbe, func() {
		topo := parseTopology(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/hosts", appPort)))
		assertAdjacent(t, topo["host:1_2_3_4"], "theinternet", "host:192_168_1_1")

		have := parseEdge(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/hosts/%s/%s", appPort, "host:192_168_1_1", "theinternet")))
		want := map[string]interface{}{
			// "window":             "15s",
			"max_conn_count_tcp": float64(12),
		}
		if !reflect.DeepEqual(have, want) {
			t.Errorf("have: %#v, want %#v", have, want)
		}
	})
}

func TestMultipleProbes(t *testing.T) {
	withContext(t, twoProbes, func() {
		topo := parseTopology(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/applications", appPort)))
		assertAdjacent(t, topo["proc:node-1.2.3.4:apache"], "theinternet", "proc:node-192.168.1.1:wget", "proc:node-192.168.1.1:curl")
	})
}

func parseTopology(t *testing.T, p []byte) map[string]report.RenderableNode {
	var r struct {
		Nodes map[string]report.RenderableNode `json:"nodes"`
	}

	if err := json.Unmarshal(p, &r); err != nil {
		t.Fatalf("parseTopology: %s", err)
	}

	return r.Nodes
}

func parseEdge(t *testing.T, p []byte) map[string]interface{} {
	var edge struct {
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal(p, &edge); err != nil {
		t.Fatalf("Err: %v", err)
	}

	return edge.Metadata
}

func assertAdjacent(t *testing.T, n report.RenderableNode, ids ...string) {
	want := report.MakeIDList(ids...)

	if have := n.Adjacency; !reflect.DeepEqual(want, have) {
		t.Fatalf("want adjacency list %v, have %v", want, have)
	}
}
