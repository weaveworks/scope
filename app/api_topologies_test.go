package app_test

import (
	"bytes"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/ugorji/go/codec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
	"github.com/weaveworks/scope/test/utils"
)

const (
	systemGroupID                   = "system"
	customAPITopologyOptionFilterID = "containerLabelFilter0"
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
	equals(t, 6, len(topologies))

	for _, topology := range topologies {
		is200(t, ts, topology.URL)

		for _, subTopology := range topology.SubTopologies {
			is200(t, ts, subTopology.URL)
		}

		// TODO: add ECS nodes in report fixture
		if topology.Name == "Tasks" || topology.Name == "services" {
			continue
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

func TestContainerLabelFilter(t *testing.T) {
	topologySummaries, err := getTestContainerLabelFilterTopologySummary(t, false)
	if err != nil {
		t.Fatalf("Topology Registry Report error: %s", err)
	}

	// only the filtered container with fixture.TestLabelKey1 should be present
	equals(t, 1, len(topologySummaries))
	for key := range topologySummaries {
		equals(t, report.MakeContainerNodeID(fixture.ClientContainerID), key)
	}
}

func TestContainerLabelFilterExclude(t *testing.T) {
	topologySummaries, err := getTestContainerLabelFilterTopologySummary(t, true)
	if err != nil {
		t.Fatalf("Topology Registry Report error: %s", err)
	}

	// all containers but the excluded container should be present
	for key := range topologySummaries {
		id := report.MakeContainerNodeID(fixture.ServerContainerNodeID)
		if id == key {
			t.Errorf("Didn't expect to find %q in report", id)
		}
	}
}

func TestRendererForTopologyWithFiltering(t *testing.T) {
	ts := topologyServer()
	defer ts.Close()

	topologyRegistry := app.MakeRegistry()
	option := app.MakeAPITopologyOption(customAPITopologyOptionFilterID, "title", render.IsApplication, false)
	topologyRegistry.AddContainerFilters(option)

	urlvalues := url.Values{}
	urlvalues.Set(systemGroupID, customAPITopologyOptionFilterID)
	urlvalues.Set("stopped", "running")
	urlvalues.Set("pseudo", "hide")
	renderer, decorator, err := topologyRegistry.RendererForTopology("containers", urlvalues, fixture.Report)
	if err != nil {
		t.Fatalf("Topology Registry Report error: %s", err)
	}

	input := fixture.Report.Copy()
	input.Container.Nodes[fixture.ClientContainerNodeID] = input.Container.Nodes[fixture.ClientContainerNodeID].WithLatests(map[string]string{
		docker.LabelPrefix + "works.weave.role": "system",
	})
	have := utils.Prune(renderer.Render(input, decorator).Nodes)
	want := utils.Prune(expected.RenderedContainers.Copy())
	delete(want, fixture.ClientContainerNodeID)
	delete(want, render.MakePseudoNodeID(render.UncontainedID, fixture.ServerHostID))
	delete(want, render.OutgoingInternetID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestRendererForTopologyNoFiltering(t *testing.T) {
	ts := topologyServer()
	defer ts.Close()

	topologyRegistry := app.MakeRegistry()
	option := app.MakeAPITopologyOption(customAPITopologyOptionFilterID, "title", nil, false)
	topologyRegistry.AddContainerFilters(option)

	urlvalues := url.Values{}
	urlvalues.Set(systemGroupID, customAPITopologyOptionFilterID)
	urlvalues.Set("stopped", "running")
	urlvalues.Set("pseudo", "hide")
	renderer, decorator, err := topologyRegistry.RendererForTopology("containers", urlvalues, fixture.Report)
	if err != nil {
		t.Fatalf("Topology Registry Report error: %s", err)
	}

	input := fixture.Report.Copy()
	input.Container.Nodes[fixture.ClientContainerNodeID] = input.Container.Nodes[fixture.ClientContainerNodeID].WithLatests(map[string]string{
		docker.LabelPrefix + "works.weave.role": "system",
	})
	have := utils.Prune(renderer.Render(input, decorator).Nodes)
	want := utils.Prune(expected.RenderedContainers.Copy())
	delete(want, render.MakePseudoNodeID(render.UncontainedID, fixture.ServerHostID))
	delete(want, render.OutgoingInternetID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func getTestContainerLabelFilterTopologySummary(t *testing.T, exclude bool) (detailed.NodeSummaries, error) {
	ts := topologyServer()
	defer ts.Close()

	var (
		topologyRegistry = app.MakeRegistry()
		filter           render.FilterFunc
	)
	if exclude == true {
		filter = render.DoesNotHaveLabel(fixture.TestLabelKey2, fixture.ApplicationLabelValue2)
	} else {
		filter = render.HasLabel(fixture.TestLabelKey1, fixture.ApplicationLabelValue1)
	}
	option := app.MakeAPITopologyOption(customAPITopologyOptionFilterID, "title", filter, false)
	topologyRegistry.AddContainerFilters(option)

	urlvalues := url.Values{}
	urlvalues.Set(systemGroupID, customAPITopologyOptionFilterID)
	urlvalues.Set("stopped", "running")
	urlvalues.Set("pseudo", "hide")
	renderer, decorator, err := topologyRegistry.RendererForTopology("containers", urlvalues, fixture.Report)
	if err != nil {
		return nil, err
	}

	return detailed.Summaries(report.RenderContext{Report: fixture.Report}, renderer.Render(fixture.Report, decorator).Nodes), nil
}

func TestAPITopologyAddsKubernetes(t *testing.T) {
	router := mux.NewRouter()
	c := app.NewCollector(1 * time.Minute)
	app.RegisterReportPostHandler(c, router)
	app.RegisterTopologyRoutes(router, c, map[string]bool{"foo_capability": true})
	ts := httptest.NewServer(router)
	defer ts.Close()

	body := getRawJSON(t, ts, "/api/topology")

	var topologies []app.APITopologyDesc
	decoder := codec.NewDecoderBytes(body, &codec.JsonHandle{})
	if err := decoder.Decode(&topologies); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
	equals(t, 6, len(topologies))

	// Enable the kubernetes topologies
	rpt := report.MakeReport()
	rpt.Pod = report.MakeTopology()
	rpt.Pod.Nodes[fixture.ClientPodNodeID] = kubernetes.NewPod(&apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pong-a",
			Namespace: "ping",
			Labels:    map[string]string{"ponger": "true"},
		},
		Status: apiv1.PodStatus{
			HostIP: "1.2.3.4",
			ContainerStatuses: []apiv1.ContainerStatus{
				{ContainerID: "container1"},
				{ContainerID: "container2"},
			},
		},
		Spec: apiv1.PodSpec{
			SecurityContext: &apiv1.PodSecurityContext{},
		},
	}).GetNode("")
	buf := &bytes.Buffer{}
	encoder := codec.NewEncoder(buf, &codec.MsgpackHandle{})
	if err := encoder.Encode(rpt); err != nil {
		t.Fatalf("Msgpack encoding error: %s", err)
	}
	checkRequest(t, ts, "POST", "/api/report", buf.Bytes())

	body = getRawJSON(t, ts, "/api/topology")
	decoder = codec.NewDecoderBytes(body, &codec.JsonHandle{})
	if err := decoder.Decode(&topologies); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
	equals(t, 6, len(topologies))

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
