package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
)

func TestProcessRenderer(t *testing.T) {
	have := render.ProcessRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedProcesses
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	have := render.ProcessNameRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedProcessNames
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerRenderer(t *testing.T) {
	have := (render.ContainerWithImageNameRenderer.Render(fixture.Report)).Prune()
	want := expected.RenderedContainers
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
	have := render.FilterSystem(render.ContainerWithImageNameRenderer).Render(input).Prune()
	want := expected.RenderedContainers.Copy()
	delete(want, expected.ClientContainerRenderedID)
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
	haveNode, ok := nodes[render.MakeContainerID(fixture.ClientContainerID)]
	if !ok {
		t.Fatal("Expected output to have the client container node")
	}
	have, _ := haveNode.Sets.Lookup(docker.ContainerIPs)
	want := report.MakeStringSet("10.10.10.0")
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerFilterRendererImageName(t *testing.T) {
	// Test nodes are filtered by image name as well.
	input := fixture.Report.Copy()
	input.ContainerImage.Nodes[fixture.ClientContainerImageNodeID] = input.ContainerImage.Nodes[fixture.ClientContainerImageNodeID].WithLatests(map[string]string{
		docker.ImageName: "beta.gcr.io/google_containers/pause",
	})
	have := render.FilterSystem(render.ContainerWithImageNameRenderer).Render(input).Prune()
	want := expected.RenderedContainers.Copy()
	delete(want, expected.ClientContainerRenderedID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageRenderer(t *testing.T) {
	have := render.ContainerImageRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedContainerImages
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestHostRenderer(t *testing.T) {
	have := render.HostRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedHosts
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPodRenderer(t *testing.T) {
	have := render.PodRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedPods
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
	want := expected.RenderedPods.Copy()
	delete(want, expected.ClientPodRenderedID)
	delete(want, expected.ClientContainerRenderedID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPodServiceRenderer(t *testing.T) {
	have := render.PodServiceRenderer.Render(fixture.Report).Prune()
	want := expected.RenderedPodServices
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
