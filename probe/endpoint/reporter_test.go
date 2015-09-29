package endpoint_test

import (
	"net"
	"strconv"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

var (
	fixLocalAddress  = net.ParseIP("192.168.1.1")
	fixLocalPort     = uint16(80)
	fixRemoteAddress = net.ParseIP("192.168.1.2")
	fixRemotePort    = uint16(12345)
	fixRemotePortB   = uint16(12346)
	fixProcessPID    = int(4242)
	fixProcessName   = "nginx"

	fixConnections = []process.Connection{
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

	fixConnectionsWithProcesses = []process.Connection{
		{
			Transport:     "tcp",
			LocalAddress:  fixLocalAddress,
			LocalPort:     fixLocalPort,
			RemoteAddress: fixRemoteAddress,
			RemotePort:    fixRemotePort,
			Process: process.Process{
				PID:  fixProcessPID,
				Comm: fixProcessName,
			},
		},
		{
			Transport:     "tcp",
			LocalAddress:  fixLocalAddress,
			LocalPort:     fixLocalPort,
			RemoteAddress: fixRemoteAddress,
			RemotePort:    fixRemotePort,
			Process: process.Process{
				PID:  fixProcessPID,
				Comm: fixProcessName,
			},
		},
	}
)

func TestSpyNoProcesses(t *testing.T) {
	const (
		nodeID   = "heinz-tomato-ketchup" // TODO rename to hostID
		nodeName = "frenchs-since-1904"   // TODO rename to hostNmae
	)

	procReader := process.MockedReader{Conns: fixConnections}
	reporter := endpoint.NewReporter(nodeID, nodeName, false, &procReader, false)
	r, _ := reporter.Report()
	//buf, _ := json.MarshalIndent(r, "", "    ")
	//t.Logf("\n%s\n", buf)

	// No process nodes, please
	if want, have := 0, len(r.Endpoint.Nodes); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	var (
		scopedLocal  = report.MakeAddressNodeID(nodeID, fixLocalAddress.String())
		scopedRemote = report.MakeAddressNodeID(nodeID, fixRemoteAddress.String())
	)

	if want, have := nodeName, r.Address.Nodes[scopedLocal].Metadata[docker.Name]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}

	if want, have := 1, len(r.Address.Nodes[scopedRemote].Adjacency); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	if want, have := scopedLocal, r.Address.Nodes[scopedRemote].Adjacency[0]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}
}

func TestSpyWithProcesses(t *testing.T) {
	const (
		nodeID   = "nikon"             // TODO rename to hostID
		nodeName = "fishermans-friend" // TODO rename to hostNmae
	)

	procReader := process.MockedReader{Conns: fixConnectionsWithProcesses}
	reporter := endpoint.NewReporter(nodeID, nodeName, true, &procReader, false)
	r, _ := reporter.Report()
	// buf, _ := json.MarshalIndent(r, "", "    ") ; t.Logf("\n%s\n", buf)

	var (
		scopedLocal  = report.MakeEndpointNodeID(nodeID, fixLocalAddress.String(), strconv.Itoa(int(fixLocalPort)))
		scopedRemote = report.MakeEndpointNodeID(nodeID, fixRemoteAddress.String(), strconv.Itoa(int(fixRemotePort)))
	)

	if want, have := 1, len(r.Endpoint.Nodes[scopedRemote].Adjacency); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	if want, have := scopedLocal, r.Endpoint.Nodes[scopedRemote].Adjacency[0]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}

	for key, want := range map[string]string{
		"pid": strconv.FormatUint(uint64(fixProcessPID), 10),
	} {
		if have := r.Endpoint.Nodes[scopedLocal].Metadata[key]; want != have {
			t.Errorf("Process.Nodes[%q][%q]: want %q, have %q", scopedLocal, key, want, have)
		}
	}
}
