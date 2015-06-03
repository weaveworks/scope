package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestGetDefault(t *testing.T) {
	md := report.NodeMetadata{"a": "1"}
	if want, have := "1", md.GetDefault("a", "2"); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
	if want, have := "2", md.GetDefault("b", "2"); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestTopologyMerge(t *testing.T) {
	want := report.Topology{
		Adjacency: report.Adjacency{
			report.MakeAdjacencyID(clientHostID, client54001EndpointNodeID): report.MakeIDList(server80EndpointNodeID),
			report.MakeAdjacencyID(fooHostID, foo42001EndpointNodeID):       report.MakeIDList(server80EndpointNodeID),
		},
		NodeMetadatas: report.NodeMetadatas{
			client54001EndpointNodeID: report.NodeMetadata{},
			foo42001EndpointNodeID:    report.NodeMetadata{},
		},
		EdgeMetadatas: report.EdgeMetadatas{
			report.MakeEdgeID(client54001EndpointNodeID, server80EndpointNodeID): report.EdgeMetadata{
				WithBytes:        true,
				BytesEgress:      30,
				BytesIngress:     300,
				WithConnCountTCP: true,
				MaxConnCountTCP:  9,
			},
			report.MakeEdgeID(foo42001EndpointNodeID, server80EndpointNodeID): report.EdgeMetadata{
				WithBytes:        true,
				BytesEgress:      40,
				BytesIngress:     400,
				WithConnCountTCP: false,
			},
		},
	}
	have := topologyFixtureA.Copy().Merge(topologyFixtureB)
	if !reflect.DeepEqual(want, have) {
		t.Errorf(diff(want, have))
	}
}

func TestAdjacencyCopy(t *testing.T) {
	one := report.Adjacency{"a": report.MakeIDList("b", "c")}
	two := one.Copy()
	one["a"].Add("d")
	if want, have := (report.Adjacency{"a": report.MakeIDList("b", "c")}), two; !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestNodeMetadatasCopy(t *testing.T) {
	one := report.NodeMetadatas{"a": report.NodeMetadata{"b": "c"}}
	two := one.Copy()
	one["a"].Merge(report.NodeMetadata{"d": "e"})
	if want, have := (report.NodeMetadatas{"a": report.NodeMetadata{"b": "c"}}), two; !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestEdgeMetadatasCopy(t *testing.T) {
	one := report.EdgeMetadatas{"a": report.EdgeMetadata{WithBytes: true}}
	two := one.Copy()
	one["a"].Merge(report.EdgeMetadata{WithConnCountTCP: true})
	if want, have := (report.EdgeMetadatas{"a": report.EdgeMetadata{WithBytes: true}}), two; !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestEdgeMetadataFlatten(t *testing.T) {
	want := report.EdgeMetadata{
		WithBytes:        true,
		BytesEgress:      30,
		BytesIngress:     300,
		WithConnCountTCP: true,
		MaxConnCountTCP:  11, // not 9!
	}
	var (
		edgeID = report.MakeEdgeID(client54001EndpointNodeID, server80EndpointNodeID)
		have   = topologyFixtureA.EdgeMetadatas[edgeID].Flatten(topologyFixtureB.EdgeMetadatas[edgeID])
	)
	if !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestEdgeMetadataExport(t *testing.T) {
	want := report.AggregateMetadata{
		report.KeyBytesEgress:     10,
		report.KeyBytesIngress:    100,
		report.KeyMaxConnCountTCP: 2,
	}
	edgeID := report.MakeEdgeID(client54001EndpointNodeID, server80EndpointNodeID)
	have := topologyFixtureA.EdgeMetadatas[edgeID].Export()
	if !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestAggregateMetadataMerge(t *testing.T) {
	want := report.AggregateMetadata{report.KeyBytesIngress: 10 + 20}
	have := (report.AggregateMetadata{report.KeyBytesIngress: 10}).Merge(report.AggregateMetadata{report.KeyBytesIngress: 20})
	if !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

var (
	fooHostID              = "foo.host.com"
	fooAddress             = "10.10.10.99"
	foo42001EndpointNodeID = report.MakeEndpointNodeID(fooHostID, fooAddress, "42001")
	fooAddressNodeID       = report.MakeAddressNodeID(fooHostID, fooAddress)

	topologyFixtureA = report.Topology{
		Adjacency: report.Adjacency{
			report.MakeAdjacencyID(clientHostID, client54001EndpointNodeID): report.MakeIDList(server80EndpointNodeID),
		},
		NodeMetadatas: report.NodeMetadatas{
			client54001EndpointNodeID: report.NodeMetadata{},
		},
		EdgeMetadatas: report.EdgeMetadatas{
			report.MakeEdgeID(client54001EndpointNodeID, server80EndpointNodeID): report.EdgeMetadata{
				WithBytes:        true,
				BytesEgress:      10,
				BytesIngress:     100,
				WithConnCountTCP: true,
				MaxConnCountTCP:  2,
			},
		},
	}

	topologyFixtureB = report.Topology{
		Adjacency: report.Adjacency{
			report.MakeAdjacencyID(clientHostID, client54001EndpointNodeID): report.MakeIDList(server80EndpointNodeID),
			report.MakeAdjacencyID(fooHostID, foo42001EndpointNodeID):       report.MakeIDList(server80EndpointNodeID),
		},
		NodeMetadatas: report.NodeMetadatas{
			client54001EndpointNodeID: report.NodeMetadata{},
			foo42001EndpointNodeID:    report.NodeMetadata{},
		},
		EdgeMetadatas: report.EdgeMetadatas{
			report.MakeEdgeID(client54001EndpointNodeID, server80EndpointNodeID): report.EdgeMetadata{
				WithBytes:        true,
				BytesEgress:      20,
				BytesIngress:     200,
				WithConnCountTCP: true,
				MaxConnCountTCP:  9,
			},
			report.MakeEdgeID(foo42001EndpointNodeID, server80EndpointNodeID): report.EdgeMetadata{
				WithBytes:    true,
				BytesEgress:  40,
				BytesIngress: 400,
			},
		},
	}
)
