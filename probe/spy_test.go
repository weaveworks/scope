package main

import (
	"encoding/json"
	"net"
	"strconv"
	"testing"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/report"
)

func TestScopedIP(t *testing.T) {
	const scope = "my-scope"

	for ip, want := range map[string]string{
		"1.2.3.4":     report.ScopeDelim + "1.2.3.4",
		"192.168.1.2": report.ScopeDelim + "192.168.1.2",
		"127.0.0.1":   scope + report.ScopeDelim + "127.0.0.1", // loopback
		"::1":         scope + report.ScopeDelim + "::1",       // loopback
		"fd00::451b:b714:85da:489e": report.ScopeDelim + "fd00::451b:b714:85da:489e", // global address
		"fe80::82ee:73ff:fe83:588f": report.ScopeDelim + "fe80::82ee:73ff:fe83:588f", // link-local address
	} {
		if have := scopedIP(scope, net.ParseIP(ip)); have != want {
			t.Errorf("%q: have %q, want %q", ip, have, want)
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
		procspy.Connection{
			Transport:     "tcp",
			LocalAddress:  fixLocalAddress,
			LocalPort:     fixLocalPort,
			RemoteAddress: fixRemoteAddress,
			RemotePort:    fixRemotePort,
		},
		procspy.Connection{
			Transport:     "tcp",
			LocalAddress:  fixLocalAddress,
			LocalPort:     fixLocalPort,
			RemoteAddress: fixRemoteAddress,
			RemotePort:    fixRemotePortB,
		},
	}

	fixConnectionsWithProcesses = []procspy.Connection{
		procspy.Connection{
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
		procspy.Connection{
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

func TestSpyNetwork(t *testing.T) {
	procspy.SetFixtures(fixConnections)

	const (
		nodeID   = "heinz-tomato-ketchup"
		nodeName = "frenchs-since-1904"
	)

	r := spy(nodeID, nodeName, false, []processMapper{})
	buf, _ := json.MarshalIndent(r, "", "    ")
	t.Logf("\n%s\n", buf)

	// No process nodes, please
	if want, have := 0, len(r.Process.Adjacency); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	var (
		scopedLocal  = scopedIP(nodeID, fixLocalAddress)
		scopedRemote = scopedIP(nodeID, fixRemoteAddress)
		localKey     = nodeID + report.IDDelim + scopedLocal
	)

	if want, have := 1, len(r.Network.Adjacency[localKey]); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	if want, have := scopedRemote, r.Network.Adjacency[localKey][0]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}

	if want, have := nodeName, r.Network.NodeMetadatas[scopedLocal]["name"]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}
}

func TestSpyProcess(t *testing.T) {
	procspy.SetFixtures(fixConnectionsWithProcesses)

	const (
		nodeID   = "nikon"
		nodeName = "fishermans-friend"
	)

	r := spy(nodeID, nodeName, true, []processMapper{})
	// buf, _ := json.MarshalIndent(r, "", "    ") ; t.Logf("\n%s\n", buf)

	var (
		scopedLocal  = scopedIPPort(nodeID, fixLocalAddress, fixLocalPort)
		scopedRemote = scopedIPPort(nodeID, fixRemoteAddress, fixRemotePort)
		localKey     = nodeID + report.IDDelim + scopedLocal
	)

	if want, have := 1, len(r.Process.Adjacency[localKey]); want != have {
		t.Fatalf("want %d, have %d", want, have)
	}

	if want, have := scopedRemote, r.Process.Adjacency[localKey][0]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}

	for key, want := range map[string]string{
		"domain": nodeID,
		"name":   fixProcessName,
		"pid":    strconv.FormatUint(uint64(fixProcessPID), 10),
	} {
		if have := r.Process.NodeMetadatas[scopedLocal][key]; want != have {
			t.Errorf("Process.NodeMetadatas[%q][%q]: want %q, have %q", scopedLocal, key, want, have)
		}
	}
}

func TestSpyProcessDataSource(t *testing.T) {
	procspy.SetFixtures(fixConnectionsWithProcesses)

	const (
		nodeID   = "chianti"
		nodeName = "harmonisch"
	)

	m := identityMapper{}
	r := spy(nodeID, nodeName, true, []processMapper{m})
	scopedLocal := scopedIPPort(nodeID, fixLocalAddress, fixLocalPort)

	k := m.Key()
	v, err := m.Map(fixProcessPID)
	if err != nil {
		t.Fatal(err)
	}

	if want, have := v, r.Process.NodeMetadatas[scopedLocal][k]; want != have {
		t.Fatalf("%s: want %q, have %q", k, want, have)
	}

	t.Logf("%s: %q OK", k, v)
}
