package report_test

import (
	"testing"

	"github.com/weaveworks/scope/report"
)

var (
	clientHostID     = "client.host.com"
	clientHostName   = clientHostID
	clientHostNodeID = report.MakeHostNodeID(clientHostID)
	clientAddress    = "10.10.10.20"
	serverHostID     = "server.host.com"
	serverHostName   = serverHostID
	serverHostNodeID = report.MakeHostNodeID(serverHostID)
	serverAddress    = "10.10.10.1"
	unknownHostID    = ""              // by definition, we don't know it
	unknownAddress   = "172.16.93.112" // will be a pseudonode, no corresponding host

	client54001EndpointNodeID = report.MakeEndpointNodeID(clientHostID, clientAddress, "54001") // i.e. curl
	client54002EndpointNodeID = report.MakeEndpointNodeID(clientHostID, clientAddress, "54002") // also curl
	server80EndpointNodeID    = report.MakeEndpointNodeID(serverHostID, serverAddress, "80")    // i.e. apache
	unknown1EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, unknownAddress, "10001")
	unknown2EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, unknownAddress, "10002")
	unknown3EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, unknownAddress, "10003")

	clientAddressNodeID  = report.MakeAddressNodeID(clientHostID, clientAddress)
	serverAddressNodeID  = report.MakeAddressNodeID(serverHostID, serverAddress)
	unknownAddressNodeID = report.MakeAddressNodeID(unknownHostID, unknownAddress)
)

func TestEndpointNodeID(t *testing.T) {
	for _, bad := range []string{
		clientAddressNodeID,
		serverAddressNodeID,
		unknownAddressNodeID,
		clientHostNodeID,
		serverHostNodeID,
		"host.com;1.2.3.4",
		"a;b",
		"a;",
		";b",
		";",
		"",
	} {
		if haveName, haveAddress, havePort, ok := report.ParseEndpointNodeID(bad); ok {
			t.Errorf("%q: expected failure, but got {%q, %q, %q}", bad, haveName, haveAddress, havePort)
		}
	}

	for input, want := range map[string]struct{ name, address, port string }{
		report.MakeEndpointNodeID("host.com", "1.2.3.4", "c"): {"", "1.2.3.4", "c"},
		"a;b;c": {"a", "b", "c"},
	} {
		haveName, haveAddress, havePort, ok := report.ParseEndpointNodeID(input)
		if !ok {
			t.Errorf("%q: not OK", input)
			continue
		}
		if want.name != haveName ||
			want.address != haveAddress ||
			want.port != havePort {
			t.Errorf("%q: want %q, have {%q, %q, %q}", input, want, haveName, haveAddress, havePort)
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
