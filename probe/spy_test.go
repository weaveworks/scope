package main

import (
	"fmt"
	"net"
	"testing"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/report"
)

func TestSpyNetwork(t *testing.T) {
	procspy.SetFixtures(fixConnections)
	const (
		hostID   = "heinz-tomato-ketchup"
		hostName = "frenchs-since-1904"
	)
	r := spy(hostID, hostName, false)
	//buf, _ := json.MarshalIndent(r, "", "    ")
	//t.Logf("\n%s\n", buf)

	// We passed fixConnections without processes. Make sure we don't get
	// process nodes.
	if want, have := 0, len(r.Process.NodeMetadatas); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	var (
		localAddressNodeID  = report.MakeAddressNodeID(hostID, fixLocalAddress.String())
		remoteAddressNodeID = report.MakeAddressNodeID(hostID, fixRemoteAddress.String())
		adjacencyID         = report.MakeAdjacencyID(hostID, localAddressNodeID)
	)
	if want, have := 1, len(r.Address.Adjacency[adjacencyID]); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
	if want, have := remoteAddressNodeID, r.Address.Adjacency[adjacencyID][0]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}
	if want, have := hostName, r.Address.NodeMetadatas[localAddressNodeID]["host_name"]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}
}

func TestSpyProcess(t *testing.T) {
	procspy.SetFixtures(fixConnectionsWithProcesses)
	const (
		hostID   = "nikon"
		hostName = "fishermans-friend"
	)
	r := spy(hostID, hostName, true)
	//buf, _ := json.MarshalIndent(r, "", "    ")
	//t.Logf("\n%s\n", buf)

	var (
		processNodeID        = report.MakeProcessNodeID(hostID, fmt.Sprint(fixProcessPIDB))
		localEndpointNodeID  = report.MakeEndpointNodeID(hostID, fixLocalAddress.String(), fmt.Sprint(fixLocalPort))
		remoteEndpointNodeID = report.MakeEndpointNodeID(hostID, fixRemoteAddress.String(), fmt.Sprint(fixRemotePort))
		adjacencyID          = report.MakeAdjacencyID(hostID, localEndpointNodeID)
	)
	if want, have := 1, len(r.Endpoint.Adjacency[adjacencyID]); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}
	if want, have := remoteEndpointNodeID, r.Endpoint.Adjacency[adjacencyID][0]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}
	for key, want := range map[string]string{
		"host_id":         hostID,
		"process_node_id": processNodeID,
	} {
		if have := r.Endpoint.NodeMetadatas[localEndpointNodeID][key]; want != have {
			t.Errorf("Endpoint.NodeMetadatas[%q][%q]: want %q, have %q", localEndpointNodeID, key, want, have)
		}
	}
	for key, want := range map[string]string{
		"pid":          fmt.Sprint(fixProcessPIDB),
		"process_name": fixProcessName,
	} {
		if have := r.Process.NodeMetadatas[processNodeID][key]; want != have {
			t.Errorf("Process.NodeMetadatas[%q][%q]: want %q, have %q", processNodeID, key, want, have)
		}
	}
}

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
