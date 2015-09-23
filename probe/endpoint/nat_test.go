package endpoint_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

type mockConntracker struct {
	flows []endpoint.Flow
}

func (m *mockConntracker) WalkFlows(f func(endpoint.Flow)) {
	for _, flow := range m.flows {
		f(flow)
	}
}

func (m *mockConntracker) Stop() {}

func TestNat(t *testing.T) {
	oldNewConntracker := endpoint.NewConntracker
	defer func() { endpoint.NewConntracker = oldNewConntracker }()

	endpoint.NewConntracker = func(existingConns bool, args ...string) (endpoint.Conntracker, error) {
		flow := makeFlow("")
		addIndependant(&flow, 1, "")
		flow.Original = addMeta(&flow, "original", "10.0.47.1", "2.3.4.5", 80, 22222)
		flow.Reply = addMeta(&flow, "reply", "2.3.4.5", "1.2.3.4", 22222, 80)

		return &mockConntracker{
			flows: []endpoint.Flow{flow},
		}, nil
	}

	have := report.MakeReport()
	originalID := report.MakeEndpointNodeID("host1", "10.0.47.1", "80")
	have.Endpoint.AddNode(originalID, report.MakeNodeWith(report.Metadata{
		endpoint.Addr: "10.0.47.1",
		endpoint.Port: "80",
		"foo":         "bar",
	}))

	want := have.Copy()
	want.Endpoint.AddNode(report.MakeEndpointNodeID("host1", "1.2.3.4", "80"), report.MakeNodeWith(report.Metadata{
		endpoint.Addr: "1.2.3.4",
		endpoint.Port: "80",
		"copy_of":     originalID,
		"foo":         "bar",
	}))

	natmapper, err := endpoint.NewNATMapper()
	if err != nil {
		t.Fatal(err)
	}

	natmapper.ApplyNAT(have, "host1")
	if !reflect.DeepEqual(want, have) {
		t.Fatal(test.Diff(want, have))
	}
}
