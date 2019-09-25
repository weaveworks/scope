// +build linux

package endpoint

import (
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/tcptracer-bpf/pkg/tracer"

	"github.com/weaveworks/scope/probe/host"
)

func newMockEbpfTracker() *EbpfTracker {
	return &EbpfTracker{
		ready: true,
		dead:  false,

		openConnections: map[fourTuple]ebpfConnection{},
	}
}

func TestHandleConnection(t *testing.T) {
	var (
		ServerPid  uint32 = 42
		ClientPid  uint32 = 43
		ServerAddr        = [net.IPv4len]byte{127, 0, 0, 1}
		ServerIP          = net.IP(ServerAddr[:])
		ClientAddr        = [net.IPv4len]byte{127, 0, 0, 2}
		ClientIP          = net.IP(ClientAddr[:])
		ServerPort uint16 = 12345
		ClientPort uint16 = 6789
		NetNS      uint32 = 123456789

		IPv4ConnectEvent = tracer.TcpV4{
			CPU:   0,
			Type:  tracer.EventConnect,
			Pid:   ClientPid,
			Comm:  "cmd",
			SAddr: ClientIP,
			DAddr: ServerIP,
			SPort: ClientPort,
			DPort: ServerPort,
			NetNS: NetNS,
		}

		IPv4ConnectEbpfConnection = ebpfConnection{
			tuple: fourTuple{
				fromAddr: ClientAddr,
				toAddr:   ServerAddr,
				fromPort: ClientPort,
				toPort:   ServerPort,
			},
			networkNamespace: uint64(NetNS),
			incoming:         false,
			pid:              int(ClientPid),
		}

		IPv4ConnectCloseEvent = tracer.TcpV4{
			CPU:   0,
			Type:  tracer.EventClose,
			Pid:   ClientPid,
			Comm:  "cmd",
			SAddr: ClientIP,
			DAddr: ServerIP,
			SPort: ClientPort,
			DPort: ServerPort,
			NetNS: NetNS,
		}

		IPv4AcceptEvent = tracer.TcpV4{
			CPU:   0,
			Type:  tracer.EventAccept,
			Pid:   ServerPid,
			Comm:  "cmd",
			SAddr: ServerIP,
			DAddr: ClientIP,
			SPort: ServerPort,
			DPort: ClientPort,
			NetNS: NetNS,
		}

		IPv4AcceptEbpfConnection = ebpfConnection{
			tuple: fourTuple{
				fromAddr: ServerAddr,
				toAddr:   ClientAddr,
				fromPort: ServerPort,
				toPort:   ClientPort,
			},
			networkNamespace: uint64(NetNS),
			incoming:         true,
			pid:              int(ServerPid),
		}

		IPv4AcceptCloseEvent = tracer.TcpV4{
			CPU:   0,
			Type:  tracer.EventClose,
			Pid:   ClientPid,
			Comm:  "cmd",
			SAddr: ServerIP,
			DAddr: ClientIP,
			SPort: ServerPort,
			DPort: ClientPort,
			NetNS: NetNS,
		}
	)

	mockEbpfTracker := newMockEbpfTracker()

	tuple := fourTuple{ClientAddr, ServerAddr, uint16(IPv4ConnectEvent.SPort), uint16(IPv4ConnectEvent.DPort)}
	mockEbpfTracker.handleConnection(IPv4ConnectEvent.Type, tuple, int(IPv4ConnectEvent.Pid), uint64(NetNS))
	if !reflect.DeepEqual(mockEbpfTracker.openConnections[tuple], IPv4ConnectEbpfConnection) {
		t.Errorf("Connection mismatch connect event\nTarget connection:%v\nParsed connection:%v",
			IPv4ConnectEbpfConnection, mockEbpfTracker.openConnections[tuple])
	}

	tuple = fourTuple{ClientAddr, ServerAddr, uint16(IPv4ConnectCloseEvent.SPort), uint16(IPv4ConnectCloseEvent.DPort)}
	mockEbpfTracker.handleConnection(IPv4ConnectCloseEvent.Type, tuple, int(IPv4ConnectCloseEvent.Pid), uint64(NetNS))
	if len(mockEbpfTracker.openConnections) != 0 {
		t.Errorf("Connection mismatch close event\nConnection to close:%v",
			mockEbpfTracker.openConnections[tuple])
	}

	mockEbpfTracker = newMockEbpfTracker()

	tuple = fourTuple{ServerAddr, ClientAddr, uint16(IPv4AcceptEvent.SPort), uint16(IPv4AcceptEvent.DPort)}
	mockEbpfTracker.handleConnection(IPv4AcceptEvent.Type, tuple, int(IPv4AcceptEvent.Pid), uint64(NetNS))
	if !reflect.DeepEqual(mockEbpfTracker.openConnections[tuple], IPv4AcceptEbpfConnection) {
		t.Errorf("Connection mismatch connect event\nTarget connection:%v\nParsed connection:%v",
			IPv4AcceptEbpfConnection, mockEbpfTracker.openConnections[tuple])
	}

	tuple = fourTuple{ServerAddr, ClientAddr, uint16(IPv4AcceptCloseEvent.SPort), uint16(IPv4AcceptCloseEvent.DPort)}
	mockEbpfTracker.handleConnection(IPv4AcceptCloseEvent.Type, tuple, int(IPv4AcceptCloseEvent.Pid), uint64(NetNS))

	if len(mockEbpfTracker.openConnections) != 0 {
		t.Errorf("Connection mismatch close event\nConnection to close:%v",
			mockEbpfTracker.openConnections)
	}
}

func TestWalkConnections(t *testing.T) {
	var (
		cnt           int
		activeTuple   = fourTuple{}
		inactiveTuple = fourTuple{}
	)
	mockEbpfTracker := newMockEbpfTracker()
	mockEbpfTracker.openConnections[activeTuple] = ebpfConnection{
		tuple:            activeTuple,
		networkNamespace: 12345,
		incoming:         true,
		pid:              0,
	}
	mockEbpfTracker.closedConnections = append(mockEbpfTracker.closedConnections,
		ebpfConnection{
			tuple:            inactiveTuple,
			networkNamespace: 12345,
			incoming:         false,
			pid:              0,
		})
	mockEbpfTracker.walkConnections(func(e ebpfConnection) {
		cnt++
	})
	if cnt != 2 {
		t.Errorf("walkConnections found %v instead of 2 connections", cnt)
	}
}

func TestInvalidTimeStampDead(t *testing.T) {
	var (
		cnt        int
		ClientPid  uint32 = 43
		ServerIP          = net.ParseIP("127.0.0.1")
		ClientIP          = net.ParseIP("127.0.0.2")
		ServerPort uint16 = 12345
		ClientPort uint16 = 6789
		NetNS      uint32 = 123456789
		event             = tracer.TcpV4{
			CPU:   0,
			Type:  tracer.EventConnect,
			Pid:   ClientPid,
			Comm:  "cmd",
			SAddr: ClientIP,
			DAddr: ServerIP,
			SPort: ClientPort,
			DPort: ServerPort,
			NetNS: NetNS,
		}
	)
	mockEbpfTracker := newMockEbpfTracker()
	event.Timestamp = 0
	mockEbpfTracker.TCPEventV4(event)
	event2 := event
	event2.SPort = 1
	event2.Timestamp = 2
	mockEbpfTracker.TCPEventV4(event2)
	mockEbpfTracker.walkConnections(func(e ebpfConnection) {
		cnt++
	})
	if cnt != 2 {
		t.Errorf("walkConnections found %v instead of 2 connections", cnt)
	}
	if mockEbpfTracker.isDead() {
		t.Errorf("expected ebpfTracker to be alive after events with valid order")
	}
	cnt = 0
	event.Timestamp = 1
	mockEbpfTracker.TCPEventV4(event)
	mockEbpfTracker.walkConnections(func(e ebpfConnection) {
		cnt++
	})
	if cnt != 2 {
		t.Errorf("walkConnections found %v instead of 2 connections", cnt)
	}
	// EbpfTracker is marked as dead asynchronously.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if mockEbpfTracker.isDead() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !mockEbpfTracker.isDead() {
		t.Errorf("expected ebpfTracker to be set to dead after events with wrong order")
	}
}

func TestIsKernelSupported(t *testing.T) {
	var release, version string
	oldGetKernelReleaseAndVersion := host.GetKernelReleaseAndVersion
	defer func() {
		host.GetKernelReleaseAndVersion = oldGetKernelReleaseAndVersion
	}()
	host.GetKernelReleaseAndVersion = func() (string, string, error) { return release, version, nil }
	testVersions := []struct {
		release   string
		version   string
		supported bool
	}{
		{
			"4.1",
			"",
			false,
		},
		{
			"4.4",
			"",
			true,
		},
		{
			"4.4-custom",
			"",
			true,
		},
		{
			"4.4.127",
			"",
			true,
		},
		{
			"4.4.0-119-generic",
			"#143-Ubuntu SMP Mon Apr 2 16:08:24 UTC 2018",
			false,
		},
		{
			"4.4.0-127-generic",
			"#153-Ubuntu SMP Sat May 19 10:58:46 UTC 2018",
			true,
		},
		{
			"4.4.0-116-generic",
			"#140-Ubuntu SMP Mon Feb 12 21:23:04 UTC 2018",
			true,
		},
		{
			"4.13.0-38-generic",
			"#43-Ubuntu SMP Wed Mar 14 15:20:44 UTC 2018",
			true,
		},
		{
			"4.13.0-119-generic",
			"#43-Ubuntu SMP Wed Apr 1 00:00:00 UTC 2018",
			true,
		},
		{
			"4.9.0-6-amd64",
			"#1 SMP Debian 4.9.82-1+deb9u3 (2018-03-02)",
			true,
		},
	}
	for _, tv := range testVersions {
		release = tv.release
		version = tv.version
		err := isKernelSupported()
		if tv.supported && err != nil {
			t.Errorf("expected kernel release %q version %q to be supported but got error: %v", release, version, err)
		}
		if !tv.supported && err == nil {
			t.Errorf("expected kernel release %q version %q to not be supported", release, version)
		}
	}
}
