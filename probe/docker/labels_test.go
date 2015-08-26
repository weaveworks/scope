package docker_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestLabels(t *testing.T) {
	want := map[string]string{
		"foo1": "bar1",
		"foo2": "bar2",
	}
	nmd := report.MakeNodeMetadata()

	docker.AddLabels(nmd, want)
	have := docker.ExtractLabels(nmd)

	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
