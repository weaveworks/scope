package report_test

import (
	"net"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestAdjacencyID(t *testing.T) {
	for _, bad := range []string{
		client54001EndpointNodeID,
		client54002EndpointNodeID,
		unknown1EndpointNodeID,
		unknown2EndpointNodeID,
		unknown3EndpointNodeID,
		clientAddressNodeID,
		serverAddressNodeID,
		unknownAddressNodeID,
		clientHostNodeID,
		serverHostNodeID,
		";",
		"",
	} {
		if srcNodeID, ok := report.ParseAdjacencyID(bad); ok {
			t.Errorf("%q: expected failure, but got (%q)", bad, srcNodeID)
		}
	}

	for input, want := range map[string]struct{ srcNodeID string }{
		report.MakeAdjacencyID(report.MakeEndpointNodeID("a", "b", "c")): {report.MakeEndpointNodeID("a", "b", "c")},
		report.MakeAdjacencyID(report.MakeAddressNodeID("a", "b")):       {report.MakeAddressNodeID("a", "b")},
		report.MakeAdjacencyID(report.MakeProcessNodeID("a", "b")):       {report.MakeProcessNodeID("a", "b")},
		report.MakeAdjacencyID(report.MakeHostNodeID("a")):               {report.MakeHostNodeID("a")},
		">host.com;1.2.3.4":                                              {"host.com;1.2.3.4"},
		">a;b;c":                                                         {"a;b;c"},
		">a;b":                                                           {"a;b"},
		">a;":                                                            {"a;"},
		">;b":                                                            {";b"},
		">;":                                                             {";"},
	} {
		srcNodeID, ok := report.ParseAdjacencyID(input)
		if !ok {
			t.Errorf("%q: not OK", input)
			continue
		}
		if want, have := want.srcNodeID, srcNodeID; want != have {
			t.Errorf("%q: want %q, have %q", input, want, have)
		}
	}
}

func TestEdgeID(t *testing.T) {
	for _, bad := range []string{
		client54001EndpointNodeID,
		client54002EndpointNodeID,
		unknown1EndpointNodeID,
		unknown2EndpointNodeID,
		unknown3EndpointNodeID,
		clientAddressNodeID,
		serverAddressNodeID,
		unknownAddressNodeID,
		clientHostNodeID,
		serverHostNodeID,
		">1.2.3.4",
		">",
		";",
		"",
	} {
		if srcNodeID, dstNodeID, ok := report.ParseEdgeID(bad); ok {
			t.Errorf("%q: expected failure, but got (%q, %q)", bad, srcNodeID, dstNodeID)
		}
	}

	for input, want := range map[string]struct{ srcNodeID, dstNodeID string }{
		report.MakeEdgeID("a", report.MakeEndpointNodeID("a", "b", "c")): {"a", report.MakeEndpointNodeID("a", "b", "c")},
		report.MakeEdgeID("a", report.MakeAddressNodeID("a", "b")):       {"a", report.MakeAddressNodeID("a", "b")},
		report.MakeEdgeID("a", report.MakeProcessNodeID("a", "b")):       {"a", report.MakeProcessNodeID("a", "b")},
		report.MakeEdgeID("a", report.MakeHostNodeID("a")):               {"a", report.MakeHostNodeID("a")},
		"host.com|1.2.3.4":                                               {"host.com", "1.2.3.4"},
		"a|b;c":                                                          {"a", "b;c"},
		"a|b":                                                            {"a", "b"},
		"a|":                                                             {"a", ""},
		"|b":                                                             {"", "b"},
		"|":                                                              {"", ""},
	} {
		srcNodeID, dstNodeID, ok := report.ParseEdgeID(input)
		if !ok {
			t.Errorf("%q: not OK", input)
			continue
		}
		if want, have := want.srcNodeID, srcNodeID; want != have {
			t.Errorf("%q: want %q, have %q", input, want, have)
		}
		if want, have := want.dstNodeID, dstNodeID; want != have {
			t.Errorf("%q: want %q, have %q", input, want, have)
		}
	}
}

func TestEndpointIDAddresser(t *testing.T) {
	if nodeID := "1.2.4.5"; report.EndpointIDAddresser(nodeID) != nil {
		t.Errorf("%q: bad node ID parsed as good", nodeID)
	}
	var (
		nodeID = report.MakeEndpointNodeID(clientHostID, clientAddress, "12345")
		want   = net.ParseIP(clientAddress)
		have   = report.EndpointIDAddresser(nodeID)
	)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %s, have %s", want, have)
	}
}

func TestAddressIDAddresser(t *testing.T) {
	if nodeID := "1.2.4.5"; report.AddressIDAddresser(nodeID) != nil {
		t.Errorf("%q: bad node ID parsed as good", nodeID)
	}
	var (
		nodeID = report.MakeAddressNodeID(clientHostID, clientAddress)
		want   = net.ParseIP(clientAddress)
		have   = report.AddressIDAddresser(nodeID)
	)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %s, have %s", want, have)
	}
}
