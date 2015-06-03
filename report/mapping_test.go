package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestSelectors(t *testing.T) {
	for _, tuple := range []struct {
		want     report.Topology
		selector report.TopologySelector
	}{
		{reportFixture.Endpoint, report.SelectEndpoint},
		{reportFixture.Address, report.SelectAddress},
		{reportFixture.Process, report.SelectProcess},
		{reportFixture.Host, report.SelectHost},
	} {
		if want, have := tuple.want, tuple.selector(reportFixture); !reflect.DeepEqual(want, have) {
			t.Error(diff(want, have))
		}
	}
}

func TestProcessPID(t *testing.T) {
	if _, ok := report.ProcessPID(reportFixture, report.SelectEndpoint, "invalid-node-ID"); ok {
		t.Errorf("want %v, have %v", false, true)
	}
	if _, ok := report.ProcessPID(reportFixture, report.SelectEndpoint, "process-not-available"); ok {
		t.Errorf("want %v, have %v", false, true)
	}
	if _, ok := report.ProcessPID(reportFixture, report.SelectEndpoint, "process-badly-linked"); ok {
		t.Errorf("want %v, have %v", false, true)
	}

	want := report.MappedNode{
		ID:    report.MakeProcessNodeID(clientHostID, "4242"),
		Major: "curl",
		Minor: "client.host.com (4242)",
		Rank:  "client.host.com",
	}
	have, ok := report.ProcessPID(reportFixture, report.SelectEndpoint, client54001EndpointNodeID)
	if !ok {
		t.Fatalf("want %v, have %v", true, false)
	}
	if !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestProcessName(t *testing.T) {
	if _, ok := report.ProcessName(reportFixture, report.SelectEndpoint, "invalid-node-ID"); ok {
		t.Errorf("want %v, have %v", false, true)
	}
	if _, ok := report.ProcessName(reportFixture, report.SelectEndpoint, "process-not-available"); ok {
		t.Errorf("want %v, have %v", false, true)
	}
	if _, ok := report.ProcessName(reportFixture, report.SelectEndpoint, "process-badly-linked"); ok {
		t.Errorf("want %v, have %v", false, true)
	}

	want := report.MappedNode{
		ID:    "curl",
		Major: "curl",
		Minor: "client.host.com",
		Rank:  "curl",
	}
	have, ok := report.ProcessName(reportFixture, report.SelectEndpoint, client54001EndpointNodeID)
	if !ok {
		t.Fatalf("want %v, have %v", true, false)
	}
	if !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestProcessContainer(t *testing.T) {
	if _, ok := report.ProcessContainer(reportFixture, report.SelectEndpoint, "invalid-node-ID"); ok {
		t.Errorf("want %v, have %v", false, true)
	}
	if _, ok := report.ProcessContainer(reportFixture, report.SelectEndpoint, "process-not-available"); ok {
		t.Errorf("want %v, have %v", false, true)
	}
	if _, ok := report.ProcessContainer(reportFixture, report.SelectEndpoint, "process-badly-linked"); ok {
		t.Errorf("want %v, have %v", false, true)
	}

	have, ok := report.ProcessContainer(reportFixture, report.SelectEndpoint, "process-no-container")
	if !ok {
		t.Fatalf("want %v, have %v", true, false)
	}
	if want := uncontained; !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}

	have, ok = report.ProcessContainer(reportFixture, report.SelectEndpoint, client54001EndpointNodeID)
	if !ok {
		t.Fatalf("want %v, have %v", true, false)
	}
	if want := (report.MappedNode{
		ID:    "a1b2c3d4e5",
		Major: "fixture-container",
		Minor: "client.host.com",
		Rank:  "0000000000",
	}); !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestProcessContainerImage(t *testing.T) {
	if _, ok := report.ProcessContainerImage(reportFixture, report.SelectEndpoint, "invalid-node-ID"); ok {
		t.Errorf("want %v, have %v", false, true)
	}
	if _, ok := report.ProcessContainerImage(reportFixture, report.SelectEndpoint, "process-not-available"); ok {
		t.Errorf("want %v, have %v", false, true)
	}
	if _, ok := report.ProcessContainerImage(reportFixture, report.SelectEndpoint, "process-badly-linked"); ok {
		t.Errorf("want %v, have %v", false, true)
	}

	have, ok := report.ProcessContainerImage(reportFixture, report.SelectEndpoint, "process-no-container")
	if !ok {
		t.Fatalf("want %v, have %v", true, false)
	}
	if want := uncontained; !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}

	have, ok = report.ProcessContainerImage(reportFixture, report.SelectEndpoint, client54001EndpointNodeID)
	if !ok {
		t.Fatalf("want %v, have %v", true, false)
	}
	if want := (report.MappedNode{
		ID:    "0000000000",
		Major: "fixture/container:latest",
		Minor: "",
		Rank:  "0000000000",
	}); !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestAddressHostname(t *testing.T) {
	if _, ok := report.AddressHostname(reportFixture, report.SelectEndpoint, "invalid-node-ID"); ok {
		t.Errorf("want %v, have %v", false, true)
	}
	if _, ok := report.AddressHostname(reportFixture, report.SelectEndpoint, "address-not-available"); ok {
		t.Errorf("want %v, have %v", false, true)
	}
	if _, ok := report.AddressHostname(reportFixture, report.SelectEndpoint, "address-badly-linked"); ok {
		t.Errorf("want %v, have %v", false, true)
	}

	want := report.MappedNode{
		ID:    report.MakeAddressNodeID(clientHostID, clientAddress),
		Major: "client",
		Minor: "host.com",
		Rank:  report.MakeAddressNodeID(clientHostID, clientAddress),
	}
	have, ok := report.AddressHostname(reportFixture, report.SelectEndpoint, client54001EndpointNodeID)
	if !ok {
		t.Fatalf("want %v, have %v", true, false)
	}
	if !reflect.DeepEqual(want, have) {
		t.Error(diff(want, have))
	}
}

func TestBasicPseudoNode(t *testing.T) {
	for nodeID, want := range map[string]report.MappedNode{
		report.TheInternet: {report.TheInternet, "the Internet", "", "theinternet"},
		"x;y;z":            {report.MakePseudoNodeID("x", "y", "z"), "y:z", "", report.MakePseudoNodeID("x", "y", "z")},
		"x;y":              {report.MakePseudoNodeID("x", "y"), "y", "", report.MakePseudoNodeID("x", "y")},
		"x":                {report.MakePseudoNodeID("x"), "x", "", report.MakePseudoNodeID("x")},
	} {
		have := report.BasicPseudoNode(nodeID)
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: %s", nodeID, diff(want, have))
			continue
		}
	}
}

func TestGroupedPseudoNode(t *testing.T) {
	for nodeID, want := range map[string]report.MappedNode{
		report.TheInternet: {report.TheInternet, "the Internet", "", "theinternet"},
		"x;y;z":            {report.MakePseudoNodeID("unknown"), "Unknown", "", report.MakePseudoNodeID("unknown")},
		"x;y":              {report.MakePseudoNodeID("unknown"), "Unknown", "", report.MakePseudoNodeID("unknown")},
		"x":                {report.MakePseudoNodeID("unknown"), "Unknown", "", report.MakePseudoNodeID("unknown")},
	} {
		have := report.GroupedPseudoNode(nodeID)
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: %s", nodeID, diff(want, have))
			continue
		}
	}
}

var uncontained = report.MappedNode{
	ID:    "uncontained",
	Major: "Uncontained",
	Minor: "",
	Rank:  "uncontained",
}
