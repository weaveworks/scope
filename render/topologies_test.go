package render_test

import (
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

func TestEndpointRenderer(t *testing.T) {
	have := render.EndpointRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedEndpoints.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessRenderer(t *testing.T) {
	have := render.ProcessRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedProcesses.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	have := render.ProcessNameRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedProcessNames.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerRenderer(t *testing.T) {
	have := render.ContainerRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedContainers.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerFilterRenderer(t *testing.T) {
	// tag on of the containers in the topology and ensure
	// it is filtered out correctly.
	input := fixture.Report.Copy()
	input.Container.Nodes[fixture.ClientContainerNodeID] = input.Container.Nodes[fixture.ClientContainerNodeID].WithLatests(map[string]string{
		docker.LabelPrefix + "works.weave.role": "system",
	})
	have := render.FilterSystem(render.ContainerRenderer).Render(input).Prune()
	want := expected.RenderedContainers.Copy().Prune()
	delete(want, fixture.ClientContainerNodeID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerWithHostIPsRenderer(t *testing.T) {
	input := fixture.Report.Copy()
	input.Container.Nodes[fixture.ClientContainerNodeID] = input.Container.Nodes[fixture.ClientContainerNodeID].WithLatests(map[string]string{
		docker.ContainerNetworkMode: "host",
	})
	nodes := render.ContainerWithHostIPsRenderer.Render(input)

	// Test host network nodes get the host IPs added.
	haveNode, ok := nodes[fixture.ClientContainerNodeID]
	if !ok {
		t.Fatal("Expected output to have the client container node")
	}
	have, ok := haveNode.Sets.Lookup(docker.ContainerIPs)
	if !ok {
		t.Fatal("Container had no IPs set.")
	}
	want := report.MakeStringSet("10.10.10.0")
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageRenderer(t *testing.T) {
	have := render.ContainerImageRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedContainerImages.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestHostRenderer(t *testing.T) {
	have := render.HostRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedHosts.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPodRenderer(t *testing.T) {
	have := render.PodRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedPods.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPodFilterRenderer(t *testing.T) {
	// tag on containers or pod namespace in the topology and ensure
	// it is filtered out correctly.
	input := fixture.Report.Copy()
	input.Pod.Nodes[fixture.ClientPodNodeID] = input.Pod.Nodes[fixture.ClientPodNodeID].WithLatests(map[string]string{
		kubernetes.PodID:     "pod:kube-system/foo",
		kubernetes.Namespace: "kube-system",
		kubernetes.PodName:   "foo",
	})
	input.Container.Nodes[fixture.ClientContainerNodeID] = input.Container.Nodes[fixture.ClientContainerNodeID].WithLatests(map[string]string{
		docker.LabelPrefix + "io.kubernetes.pod.name": "kube-system/foo",
	})
	have := render.FilterSystem(render.PodRenderer).Render(input).Prune()
	want := expected.RenderedPods.Copy().Prune()
	delete(want, fixture.ClientPodNodeID)
	delete(want, fixture.ClientContainerNodeID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPodServiceRenderer(t *testing.T) {
	have := render.PodServiceRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedPodServices.Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
