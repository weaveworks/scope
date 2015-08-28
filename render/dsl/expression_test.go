package dsl_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render/dsl"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

var fixture = report.Topology{
	Adjacency: map[string]report.IDList{
		report.MakeAdjacencyID("a"): report.MakeIDList("b", "c"), // a -> b, a-> c
		report.MakeAdjacencyID("b"): report.MakeIDList("c"),      // b -> c
		report.MakeAdjacencyID("c"): report.MakeIDList("c"),      // c -> c
	},
	NodeMetadatas: map[string]report.NodeMetadata{
		"a": report.MakeNodeMetadataWith(map[string]string{"is-a-or-b-or-c": "true", "is-a-or-b": "true", "is-a": "true"}),
		"b": report.MakeNodeMetadataWith(map[string]string{"is-a-or-b-or-c": "true", "is-a-or-b": "true"}),
		"c": report.MakeNodeMetadataWith(map[string]string{"is-a-or-b-or-c": "true"}),
	},
	EdgeMetadatas: map[string]report.EdgeMetadata{
		report.MakeEdgeID("a", "c"): report.EdgeMetadata{EgressPacketCount: newu64(1)},
		report.MakeEdgeID("b", "c"): report.EdgeMetadata{EgressPacketCount: newu64(2)},
	},
}

func TestSingleExpressions(t *testing.T) {
	for _, testcase := range []struct {
		input string
		want  report.Topology
	}{
		{
			"ALL REMOVE",
			report.MakeTopology(),
		},
		{
			"NOT WITH {{is-a-or-b}} REMOVE",
			report.Topology{
				Adjacency: map[string]report.IDList{
					report.MakeAdjacencyID("a"): report.MakeIDList("b"),
				},
				NodeMetadatas: map[string]report.NodeMetadata{
					"a": report.MakeNodeMetadataWith(map[string]string{"is-a-or-b-or-c": "true", "is-a-or-b": "true", "is-a": "true"}),
					"b": report.MakeNodeMetadataWith(map[string]string{"is-a-or-b-or-c": "true", "is-a-or-b": "true"}),
				},
				EdgeMetadatas: map[string]report.EdgeMetadata{},
			},
		},
	} {
		expr, err := dsl.ParseExpression(testcase.input)
		if err != nil {
			t.Errorf("%q: %v", testcase.input, err)
			continue
		}
		if want, have := testcase.want, expr.Eval(fixture); !reflect.DeepEqual(want, have) {
			t.Errorf("%s: %s", testcase.input, test.Diff(want, have))
			continue
		}
	}
}

func newu64(value uint64) *uint64 { return &value }
