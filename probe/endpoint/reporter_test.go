package endpoint_test

import (
	"net"
	"strconv"
	"testing"

	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/endpoint/procspy"
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
				PID:  fixProcessPID,
				Name: fixProcessName,
			},
		},
	}
)

func TestSpyNoProcesses(t *testing.T) {
	const (
		nodeID   = "heinz-tomato-ketchup" // TODO rename to hostID
		nodeName = "frenchs-since-1904"   // TODO rename to hostName
	)

	scanner := procspy.FixedScanner(fixConnections)
	reporter := endpoint.NewReporter(nodeID, nodeName, false, false, false, scanner)
	r, _ := reporter.Report()
	//buf, _ := json.MarshalIndent(r, "", "    ")
	//t.Logf("\n%s\n", buf)

	if want, have := 0, len(r.Endpoint.Nodes); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
}

func TestSpyWithProcesses(t *testing.T) {
	const (
		nodeID   = "nikon"             // TODO rename to hostID
		nodeName = "fishermans-friend" // TODO rename to hostName
	)

	scanner := procspy.FixedScanner(fixConnectionsWithProcesses)
	reporter := endpoint.NewReporter(nodeID, nodeName, true, false, true, scanner)
	r, _ := reporter.Report()
	// buf, _ := json.MarshalIndent(r, "", "    ") ; t.Logf("\n%s\n", buf)

	var (
		scopedLocal  = report.MakeEndpointNodeID(nodeID, "", fixLocalAddress.String(), strconv.Itoa(int(fixLocalPort)))
		scopedRemote = report.MakeEndpointNodeID(nodeID, "", fixRemoteAddress.String(), strconv.Itoa(int(fixRemotePort)))
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
		have, _ := r.Endpoint.Nodes[scopedLocal].Const.Lookup(key)
		if want != have {
			t.Errorf("Process.Nodes[%q][%q]: want %q, have %q", scopedLocal, key, want, have)
		}
	}
}
