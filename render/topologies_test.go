package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test"
)

func TestProcessRenderer(t *testing.T) {
	have := render.ProcessRenderer.Render(test.Report).Prune()
	want := expected.RenderedProcesses
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	have := render.ProcessNameRenderer.Render(test.Report).Prune()
	want := expected.RenderedProcessNames
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerRenderer(t *testing.T) {
	have := (render.ContainerWithImageNameRenderer.Render(test.Report)).Prune()
	want := expected.RenderedContainers
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerFilterRenderer(t *testing.T) {
	// tag on of the containers in the topology and ensure
	// it is filtered out correctly.
	input := test.Report.Copy()
	input.Container.Nodes[test.ClientContainerNodeID].Metadata[docker.LabelPrefix+"works.weave.role"] = "system"
	have := render.FilterSystem(render.ContainerWithImageNameRenderer).Render(input).Prune()
	want := expected.RenderedContainers.Copy()
	delete(want, test.ClientContainerID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestContainerImageRenderer(t *testing.T) {
	have := render.ContainerImageRenderer.Render(test.Report).Prune()
	want := expected.RenderedContainerImages
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestHostRenderer(t *testing.T) {
	have := render.HostRenderer.Render(test.Report).Prune()
	want := expected.RenderedHosts
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPodRenderer(t *testing.T) {
	have := render.PodRenderer.Render(test.Report).Prune()
	want := expected.RenderedPods
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPodFilterRenderer(t *testing.T) {
	// tag on containers or pod namespace in the topology and ensure
	// it is filtered out correctly.
	input := test.Report.Copy()
	input.Pod.Nodes[test.ClientPodNodeID].Metadata[kubernetes.PodID] = "kube-system/foo"
	input.Pod.Nodes[test.ClientPodNodeID].Metadata[kubernetes.Namespace] = "kube-system"
	input.Pod.Nodes[test.ClientPodNodeID].Metadata[kubernetes.PodName] = "foo"
	input.Container.Nodes[test.ClientContainerNodeID].Metadata[docker.LabelPrefix+"io.kubernetes.pod.name"] = "kube-system/foo"
	have := render.FilterSystem(render.PodRenderer).Render(input).Prune()
	want := expected.RenderedPods.Copy()
	delete(want, test.ClientPodID)
	delete(want, test.ClientContainerID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPodServiceRenderer(t *testing.T) {
	have := render.PodServiceRenderer.Render(test.Report).Prune()
	want := expected.RenderedPodServices
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
