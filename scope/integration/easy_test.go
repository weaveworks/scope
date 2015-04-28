package integration

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/cello/report"
)

func TestComponentsAreAvailable(t *testing.T) {
	var pause = 1 * time.Millisecond

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

func TestProcessPID(t *testing.T) {
	withContext(t, oneProbe, func() {
		topo := parseTopology(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/processpid", appPort)))
		assertAdjacent(t, topo["pid:node-1.2.3.4:4000"], "theinternet", "pid:node-192.168.1.1:4000")
	})
}

func TestProcessName(t *testing.T) {
	withContext(t, oneProbe, func() {
		topo := parseTopology(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/processname", appPort)))
		assertAdjacent(t, topo["proc:node-1.2.3.4:apache"], "theinternet", "proc:node-192.168.1.1:wget")

		have := parseEdge(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/processname/%s/%s", appPort, "proc:node-192.168.1.1:wget", "theinternet")))
		want := map[string]interface{}{
			"max_conn_count_tcp": float64(19),
		}
		if !reflect.DeepEqual(have, want) {
			t.Errorf("have: %#v, want %#v", have, want)
		}
	})
}

func TestNetworkIP(t *testing.T) {
	withContext(t, oneProbe, func() {
		topo := parseTopology(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/networkip", appPort)))
		assertAdjacent(t, topo["addr:;1.2.3.4"], "theinternet", "addr:;192.168.1.1")
	})
}

func TestNetworkHost(t *testing.T) {
	withContext(t, oneProbe, func() {
		topo := parseTopology(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/networkhost", appPort)))
		assertAdjacent(t, topo["host:1_2_3_4"], "theinternet", "host:192_168_1_1")

		have := parseEdge(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/networkhost/%s/%s", appPort, "host:192_168_1_1", "theinternet")))
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
		topo := parseTopology(t, httpGet(t, fmt.Sprintf("http://localhost:%d/api/topology/processname", appPort)))
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
	want := report.NewIDList(ids...)

	if have := n.Adjacency; !reflect.DeepEqual(want, have) {
		t.Fatalf("want adjacency list %v, have %v", want, have)
	}
}
