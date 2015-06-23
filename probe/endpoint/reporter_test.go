package endpoint_test

import (
	"net"
	"strconv"
	"testing"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/report"
)

var (
	fixLocalAddress  = net.ParseIP("192.168.1.1")
	fixLocalPort     = uint16(80)
	fixRemoteAddress = net.ParseIP("192.168.1.2")
	fixRemotePort    = uint16(12345)
	fixRemotePortB   = uint16(12346)
	fixProcessPID    = uint(4242)
	fixProcessName   = "nginx"
	fixProcessPIDB   = uint(4243)

	fixConnections = []procspy.Connection{
		{
			Transport:     "tcp",
			LocalAddress:  fixLocalAddress,
			LocalPort:     fixLocalPort,
			RemoteAddress: fixRemoteAddress,
			RemotePort:    fixRemotePort,
		},
		{
			Transport:     "tcp",
			LocalAddress:  fixLocalAddress,
			LocalPort:     fixLocalPort,
			RemoteAddress: fixRemoteAddress,
			RemotePort:    fixRemotePortB,
		},
	}

	fixConnectionsWithProcesses = []procspy.Connection{
		{
			Transport:     "tcp",
			LocalAddress:  fixLocalAddress,
			LocalPort:     fixLocalPort,
			RemoteAddress: fixRemoteAddress,
			RemotePort:    fixRemotePort,
			Proc: procspy.Proc{
				PID:  fixProcessPID,
				Name: fixProcessName,
			},
		},
		{
			Transport:     "tcp",
			LocalAddress:  fixLocalAddress,
			LocalPort:     fixLocalPort,
			RemoteAddress: fixRemoteAddress,
			RemotePort:    fixRemotePort,
			Proc: procspy.Proc{
				PID:  fixProcessPIDB,
				Name: fixProcessName,
			},
		},
	}
)

func TestSpyNoProcesses(t *testing.T) {
	procspy.SetFixtures(fixConnections)

	const (
		nodeID   = "heinz-tomato-ketchup" // TODO rename to hostID
		nodeName = "frenchs-since-1904"   // TODO rename to hostNmae
	)

	reporter := endpoint.NewReporter(nodeID, nodeName, false)
	r, _ := reporter.Report()
	//buf, _ := json.MarshalIndent(r, "", "    ")
	//t.Logf("\n%s\n", buf)

	// No process nodes, please
	if want, have := 0, len(r.Endpoint.Adjacency); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	var (
		scopedLocal  = report.MakeAddressNodeID(nodeID, fixLocalAddress.String())
		scopedRemote = report.MakeAddressNodeID(nodeID, fixRemoteAddress.String())
		localKey     = report.MakeAdjacencyID(scopedLocal)
	)

	if want, have := 1, len(r.Address.Adjacency[localKey]); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	if want, have := scopedRemote, r.Address.Adjacency[localKey][0]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}

	if want, have := nodeName, r.Address.NodeMetadatas[scopedLocal]["name"]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}
}

func TestSpyWithProcesses(t *testing.T) {
	procspy.SetFixtures(fixConnectionsWithProcesses)

	const (
		nodeID   = "nikon"             // TODO rename to hostID
		nodeName = "fishermans-friend" // TODO rename to hostNmae
	)

	reporter := endpoint.NewReporter(nodeID, nodeName, false)
	r, _ := reporter.Report()
	// buf, _ := json.MarshalIndent(r, "", "    ") ; t.Logf("\n%s\n", buf)

	var (
		scopedLocal  = report.MakeEndpointNodeID(nodeID, fixLocalAddress.String(), strconv.Itoa(int(fixLocalPort)))
		scopedRemote = report.MakeEndpointNodeID(nodeID, fixRemoteAddress.String(), strconv.Itoa(int(fixRemotePort)))
		localKey     = report.MakeAdjacencyID(scopedLocal)
	)

	if want, have := 1, len(r.Endpoint.Adjacency[localKey]); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	if want, have := scopedRemote, r.Endpoint.Adjacency[localKey][0]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}

	for key, want := range map[string]string{
		"pid": strconv.FormatUint(uint64(fixProcessPID), 10),
	} {
		if have := r.Endpoint.NodeMetadatas[scopedLocal][key]; want != have {
			t.Errorf("Process.NodeMetadatas[%q][%q]: want %q, have %q", scopedLocal, key, want, have)
		}
	}
}
