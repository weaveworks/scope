package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestAPITopology(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}))
	defer ts.Close()

	body := getRawJSON(t, ts, "/api/topology")

	var topologies []APITopologyDesc
	if err := json.Unmarshal(body, &topologies); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
	equals(t, 4, len(topologies))

	for _, topology := range topologies {
		is200(t, ts, topology.URL)

		for _, subTopology := range topology.SubTopologies {
			is200(t, ts, subTopology.URL)
		}

		if have := topology.Stats.EdgeCount; have <= 0 {
			t.Errorf("EdgeCount isn't positive for %s: %d", topology.Name, have)
		}

		if have := topology.Stats.NodeCount; have <= 0 {
			t.Errorf("NodeCount isn't positive for %s: %d", topology.Name, have)
		}

		if have := topology.Stats.NonpseudoNodeCount; have <= 0 {
			t.Errorf("NonpseudoNodeCount isn't positive for %s: %d", topology.Name, have)
		}
	}
}
