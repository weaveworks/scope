package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestRenderEndpointAsProcessNameWithoutPseudoNodes(t *testing.T) {
	have := report.Render(reportFixture, report.SelectEndpoint, report.ProcessName, nil)
	want := map[string]report.RenderableNode{
		"apache": {
			ID:         "apache",
			LabelMajor: "apache",
			LabelMinor: "(unknown)", // because we didn't set a host_name
			Rank:       "apache",
			Pseudo:     false,
			Adjacency: report.MakeIDList(
				"curl",
			),
			Origins: report.MakeIDList(
				server80EndpointNodeID,
				serverHostNodeID,
			),
			Metadata: report.AggregateMetadata{
				report.KeyBytesEgress:  100, // Note that when we don't render pseudonodes
				report.KeyBytesIngress: 10,  // we lose their metadata. Could be dangerous!
			},
		},
		"curl": {
			ID:         "curl",
			LabelMajor: "curl",
			LabelMinor: "client.host.com",
			Rank:       "curl",
			Pseudo:     false,
			Adjacency: report.MakeIDList(
				"apache",
			),
			Origins: report.MakeIDList(
				client54001EndpointNodeID,
				//client54002EndpointNodeID,
				clientHostNodeID,
			),
			Metadata: report.AggregateMetadata{
				report.KeyBytesEgress:  10,
				report.KeyBytesIngress: 100,
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestRenderEndpointAsProcessNameWithBasicPseudoNodes(t *testing.T) {
	have := report.Render(reportFixture, report.SelectEndpoint, report.ProcessName, report.BasicPseudoNode)
	want := map[string]report.RenderableNode{
		"apache": {
			ID:         "apache",
			LabelMajor: "apache",
			LabelMinor: "(unknown)", // because we didn't set a host_name
			Rank:       "apache",
			Pseudo:     false,
			Adjacency: report.MakeIDList(
				"curl",
				report.MakePseudoNodeID(unknownHostID, unknownAddress, "10001"),
				report.MakePseudoNodeID(unknownHostID, unknownAddress, "10002"),
				report.MakePseudoNodeID(unknownHostID, unknownAddress, "10003"),
				report.MakePseudoNodeID(clientHostID, clientAddress, "54002"),
			),
			Origins: report.MakeIDList(
				server80EndpointNodeID,
				serverHostNodeID,
			),
			Metadata: report.AggregateMetadata{
				report.KeyBytesEgress:  3100, // Here, the metadata is preserved,
				report.KeyBytesIngress: 310,  // thanks to the pseudonode.
			},
		},
		"curl": {
			ID:         "curl",
			LabelMajor: "curl",
			LabelMinor: clientHostID,
			Rank:       "curl",
			Pseudo:     false,
			Adjacency: report.MakeIDList(
				"apache",
			),
			Origins: report.MakeIDList(
				client54001EndpointNodeID,
				//client54002EndpointNodeID,
				clientHostNodeID,
			),
			Metadata: report.AggregateMetadata{
				report.KeyBytesEgress:  10,  // Here, we lose some outgoing metadata,
				report.KeyBytesIngress: 100, // but that's to be expected.
			},
		},
		report.MakePseudoNodeID(unknownHostID, unknownAddress, "10001"): {
			ID:         report.MakePseudoNodeID(unknownHostID, unknownAddress, "10001"),
			LabelMajor: unknownAddress + ":10001",
			LabelMinor: "",
			Rank:       report.MakePseudoNodeID(unknownHostID, unknownAddress, "10001"),
			Pseudo:     true,
			Adjacency:  report.IDList{},
			Origins:    report.IDList{},
			Metadata:   report.AggregateMetadata{},
		},
		report.MakePseudoNodeID(unknownHostID, unknownAddress, "10002"): {
			ID:         report.MakePseudoNodeID(unknownHostID, unknownAddress, "10002"),
			LabelMajor: unknownAddress + ":10002",
			LabelMinor: "",
			Rank:       report.MakePseudoNodeID(unknownHostID, unknownAddress, "10002"),
			Pseudo:     true,
			Adjacency:  report.IDList{},
			Origins:    report.IDList{},
			Metadata:   report.AggregateMetadata{},
		},
		report.MakePseudoNodeID(unknownHostID, unknownAddress, "10003"): {
			ID:         report.MakePseudoNodeID(unknownHostID, unknownAddress, "10003"),
			LabelMajor: unknownAddress + ":10003",
			LabelMinor: "",
			Rank:       report.MakePseudoNodeID(unknownHostID, unknownAddress, "10003"),
			Pseudo:     true,
			Adjacency:  report.IDList{},
			Origins:    report.IDList{},
			Metadata:   report.AggregateMetadata{},
		},
		report.MakePseudoNodeID(clientHostID, clientAddress, "54002"): {
			ID:         report.MakePseudoNodeID(clientHostID, clientAddress, "54002"),
			LabelMajor: clientAddress + ":54002",
			LabelMinor: "",
			Rank:       report.MakePseudoNodeID(clientHostID, clientAddress, "54002"),
			Pseudo:     true,
			Adjacency: report.MakeIDList(
				"apache",
			),
			Origins: report.MakeIDList(
				client54002EndpointNodeID,
				clientHostNodeID,
			),
			Metadata: report.AggregateMetadata{
				report.KeyBytesEgress:  20,
				report.KeyBytesIngress: 200,
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestRenderPanicOnBadAdjacencyID(t *testing.T) {
	if panicked := func() (recovered bool) {
		defer func() {
			if r := recover(); r != nil {
				recovered = true
			}
		}()
		r := report.MakeReport()
		r.Endpoint.Adjacency["bad-adjacency-id"] = report.MakeIDList(server80EndpointNodeID)
		report.Render(r, report.SelectEndpoint, report.ProcessPID, nil)
		return false
	}(); !panicked {
		t.Errorf("expected panic, didn't get it")
	}

}
