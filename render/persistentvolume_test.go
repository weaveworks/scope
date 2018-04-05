package render_test

import (
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
	"github.com/weaveworks/scope/test/utils"
)

func TestPersistentVolumeRenderer(t *testing.T) {
	have := utils.Prune(render.PersistentVolumeRenderer.Render(fixture.Report).Nodes)
	want := utils.Prune(expected.RenderedPersistentVolume)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
