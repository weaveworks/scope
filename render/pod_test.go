package render_test

import (
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
	"github.com/weaveworks/scope/test/utils"
)

func TestPodRenderer(t *testing.T) {
	have := utils.Prune(render.PodRenderer.Render(fixture.Report, nil).Nodes)
	want := utils.Prune(expected.RenderedPods)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func filterNonKubeSystem(renderer render.Renderer) render.Renderer {
	return render.MakeFilter(render.Complement(render.IsNamespace("kube-system")), renderer)
}

func TestPodFilterRenderer(t *testing.T) {
	// tag on containers or pod namespace in the topology and ensure
	// it is filtered out correctly.
	input := fixture.Report.Copy()
	renderer := render.ApplyDecorator(render.PodRenderer)

	input.Pod.Nodes[fixture.ClientPodNodeID] = input.Pod.Nodes[fixture.ClientPodNodeID].WithLatests(map[string]string{
		kubernetes.Namespace: "kube-system",
	})

	have := utils.Prune(renderer.Render(input, filterNonKubeSystem).Nodes)
	want := utils.Prune(expected.RenderedPods.Copy())
	delete(want, fixture.ClientPodNodeID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPodServiceRenderer(t *testing.T) {
	have := utils.Prune(render.PodServiceRenderer.Render(fixture.Report, nil).Nodes)
	want := utils.Prune(expected.RenderedPodServices)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestPodServiceFilterRenderer(t *testing.T) {
	// tag on containers or pod namespace in the topology and ensure
	// it is filtered out correctly.
	input := fixture.Report.Copy()
	renderer := render.ApplyDecorator(render.PodServiceRenderer)

	input.Service.Nodes[fixture.ServiceNodeID] = input.Service.Nodes[fixture.ServiceNodeID].WithLatests(map[string]string{
		kubernetes.Namespace: "kube-system",
	})

	have := utils.Prune(renderer.Render(input, filterNonKubeSystem).Nodes)
	want := utils.Prune(expected.RenderedPodServices.Copy())
	delete(want, fixture.ServiceNodeID)
	delete(want, render.IncomingInternetID)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
