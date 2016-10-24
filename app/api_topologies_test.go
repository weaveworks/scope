package app_test

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/ugorji/go/codec"
	"k8s.io/kubernetes/pkg/api"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

func TestAPITopology(t *testing.T) {
	ts := topologyServer()
	defer ts.Close()

	body := getRawJSON(t, ts, "/api/topology")

	var topologies []app.APITopologyDesc
	decoder := codec.NewDecoderBytes(body, &codec.JsonHandle{})
	if err := decoder.Decode(&topologies); err != nil {
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

func TestAPITopologyAddsKubernetes(t *testing.T) {
	router := mux.NewRouter()
	c := app.NewCollector(1 * time.Minute)
	app.RegisterReportPostHandler(c, router)
	app.RegisterTopologyRoutes(router, c)
	ts := httptest.NewServer(router)
	defer ts.Close()

	body := getRawJSON(t, ts, "/api/topology")

	var topologies []app.APITopologyDesc
	decoder := codec.NewDecoderBytes(body, &codec.JsonHandle{})
	if err := decoder.Decode(&topologies); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
	equals(t, 4, len(topologies))

	// Enable the kubernetes topologies
	rpt := report.MakeReport()
	rpt.Pod = report.MakeTopology()
	rpt.Pod.Nodes[fixture.ClientPodNodeID] = kubernetes.NewPod(&api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:      "pong-a",
			Namespace: "ping",
			Labels:    map[string]string{"ponger": "true"},
		},
		Status: api.PodStatus{
			HostIP: "1.2.3.4",
			ContainerStatuses: []api.ContainerStatus{
				{ContainerID: "container1"},
				{ContainerID: "container2"},
			},
		},
		Spec: api.PodSpec{
			SecurityContext: &api.PodSecurityContext{},
		},
	}).GetNode("")
	buf := &bytes.Buffer{}
	encoder := codec.NewEncoder(buf, &codec.MsgpackHandle{})
	if err := encoder.Encode(rpt); err != nil {
		t.Fatalf("GOB encoding error: %s", err)
	}
	checkRequest(t, ts, "POST", "/api/report", buf.Bytes())

	body = getRawJSON(t, ts, "/api/topology")
	decoder = codec.NewDecoderBytes(body, &codec.JsonHandle{})
	if err := decoder.Decode(&topologies); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
	equals(t, 4, len(topologies))

	found := false
	for _, topology := range topologies {
		if topology.Name == "Pods" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Could not find pods topology")
	}
}
