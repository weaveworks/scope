package render_test

import (
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestTopology2RenderableNodes(t *testing.T) {

	var (
		newu64     = func(value uint64) *uint64 { return &value }
		srcNodeID  = "srcNode"
		dstNode1ID = "dstNode1"
		dstNode2ID = "dstNode2"
		srcNode    = report.MakeNode().
				WithEdge(dstNode1ID, report.EdgeMetadata{EgressPacketCount: newu64(100), EgressByteCount: newu64(1000)}).
				WithEdge(dstNode2ID, report.EdgeMetadata{EgressPacketCount: newu64(200), EgressByteCount: newu64(2000)})
		dstNode1 = report.MakeNode()
		dstNode2 = report.MakeNode()
		topology = report.MakeTopology().
				AddNode(srcNodeID, srcNode).
				AddNode(dstNode1ID, dstNode1).
				AddNode(dstNode2ID, dstNode2)
	)

	result := render.Topology2RenderableNodes(topology)
	mustLookup := func(id string) render.RenderableNode {
		node, ok := result.Lookup(id)
		if !ok {
			t.Fatalf("Expected result to contain node: %q, got: %v", id, result)
		}
		return node
	}

	// Source nodes should have the flattened edge metadata
	{
		have := mustLookup(srcNodeID).EdgeMetadata
		want := report.EdgeMetadata{EgressPacketCount: newu64(300), EgressByteCount: newu64(3000)}
		if !reflect.DeepEqual(want, have) {
			t.Errorf(test.Diff(want, have))
		}
	}

	// Result destination nodes should have the reverse of the source nodes
	{
		have := mustLookup(dstNode1ID).EdgeMetadata
		want := report.EdgeMetadata{IngressPacketCount: newu64(100), IngressByteCount: newu64(1000)}
		if !reflect.DeepEqual(want, have) {
			t.Errorf(test.Diff(want, have))
		}

		have = mustLookup(dstNode2ID).EdgeMetadata
		want = report.EdgeMetadata{IngressPacketCount: newu64(200), IngressByteCount: newu64(2000)}
		if !reflect.DeepEqual(want, have) {
			t.Errorf(test.Diff(want, have))
		}
	}
}
