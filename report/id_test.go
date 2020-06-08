package report_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	client54001EndpointNodeID = report.MakeEndpointNodeID(clientHostID, "", clientAddress, "54001") // i.e. curl
	client54002EndpointNodeID = report.MakeEndpointNodeID(clientHostID, "", clientAddress, "54002") // also curl
	server80EndpointNodeID    = report.MakeEndpointNodeID(serverHostID, "", serverAddress, "80")    // i.e. apache
	unknown1EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, "", unknownAddress, "10001")
	unknown2EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, "", unknownAddress, "10002")
	unknown3EndpointNodeID    = report.MakeEndpointNodeID(unknownHostID, "", unknownAddress, "10003")

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
		report.MakeEndpointNodeID("host.com", "namespaceid", "127.0.0.1", "c"): {"host.com-namespaceid", "127.0.0.1", "c"},
		report.MakeEndpointNodeID("host.com", "", "1.2.3.4", "c"):              {"", "1.2.3.4", "c"},
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

func TestECSServiceNodeIDCompat(t *testing.T) {
	testID := "my-service;<ecs_service>"
	testName := "my-service"
	_, name, ok := report.ParseECSServiceNodeID(testID)
	if !ok {
		t.Errorf("Failed to parse backwards-compatible id %q", testID)
	}
	if name != testName {
		t.Errorf("Backwards-compatible id %q parsed name to %q, expected %q", testID, name, testName)
	}
}

func TestNodeIDType(t *testing.T) {
	ty, ok := report.NodeIDType("")
	assert.False(t, ok)
	ty, ok = report.NodeIDType(clientHostNodeID)
	assert.True(t, ok)
	assert.Equal(t, report.Host, ty)
	ty, ok = report.NodeIDType(client54001EndpointNodeID)
	assert.True(t, ok)
	assert.Equal(t, report.Endpoint, ty)
	rsetID := report.MakeReplicaSetNodeID("foo")
	ty, ok = report.NodeIDType(rsetID)
	assert.True(t, ok)
	assert.Equal(t, report.ReplicaSet, ty)
}
