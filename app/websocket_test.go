package main

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestTopologyDiff(t *testing.T) {
	for i, tuple := range []struct {
		a, b map[string]report.RenderableNode
		want diff
	}{
		{
			map[string]report.RenderableNode{},
			map[string]report.RenderableNode{},
			diff{},
		},
		{
			map[string]report.RenderableNode{},
			map[string]report.RenderableNode{"a": {ID: "a"}},
			diff{Add: []report.RenderableNode{{ID: "a"}}},
		},
		{
			map[string]report.RenderableNode{"a": {ID: "a"}},
			map[string]report.RenderableNode{},
			diff{Remove: []string{"a"}},
		},
		{
			map[string]report.RenderableNode{"a": {ID: "a"}},
			map[string]report.RenderableNode{"a": {ID: "a", LabelMajor: "different"}},
			diff{Update: []report.RenderableNode{{ID: "a", LabelMajor: "different"}}},
		},
		{
			map[string]report.RenderableNode{"c": {ID: "c"}, "b": {ID: "b"}},
			map[string]report.RenderableNode{"a": {ID: "a"}, "b": {ID: "b", LabelMajor: "different"}},
			diff{
				Add:    []report.RenderableNode{{ID: "a"}},
				Update: []report.RenderableNode{{ID: "b", LabelMajor: "different"}},
				Remove: []string{"c"},
			},
		},
	} {
		if want, have := tuple.want, topologyDiff(tuple.a, tuple.b); !reflect.DeepEqual(want, have) {
			t.Errorf("%d: want\n\t%#+v, have\n\t%#+v", i, want, have)
		}
	}
}
