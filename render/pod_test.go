package render_test

import (
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

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
