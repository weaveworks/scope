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
	var topos []APITopologyDesc
	if err := json.Unmarshal(body, &topos); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
	equals(t, 3, len(topos))
	for _, topo := range topos {
		is200(t, ts, topo.URL)
		if topo.GroupedURL != "" {
			is200(t, ts, topo.GroupedURL)
		}
		if have := topo.Stats.EdgeCount; have <= 0 {
			t.Errorf("EdgeCount isn't positive: %d", have)
		}
		if have := topo.Stats.NodeCount; have <= 0 {
			t.Errorf("NodeCount isn't positive: %d", have)
		}
		if have := topo.Stats.NonpseudoNodeCount; have <= 0 {
			t.Errorf("NonpseudoNodeCount isn't positive: %d", have)
		}
	}
}
