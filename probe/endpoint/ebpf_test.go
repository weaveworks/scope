package endpoint

import (
	"net"
	"reflect"
	"strconv"
	"testing"

	"github.com/weaveworks/tcptracer-bpf/pkg/tracer"
)

func TestHandleConnection(t *testing.T) {
	var (
		ServerPid  uint32 = 42
		ClientPid  uint32 = 43
		ServerIP          = net.IP("127.0.0.1")
		ClientIP          = net.IP("127.0.0.2")
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
				fromAddr: ClientIP.String(),
				toAddr:   ServerIP.String(),
				fromPort: ClientPort,
				toPort:   ServerPort,
			},
			networkNamespace: strconv.Itoa(int(NetNS)),
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
				fromAddr: ServerIP.String(),
				toAddr:   ClientIP.String(),
				fromPort: ServerPort,
				toPort:   ClientPort,
			},
			networkNamespace: strconv.Itoa(int(NetNS)),
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

	mockEbpfTracker := &EbpfTracker{
		readyToHandleConnections: true,
		dead: false,

		openConnections:   map[fourTuple]ebpfConnection{},
		closedConnections: []ebpfConnection{},
	}

	tuple := fourTuple{IPv4ConnectEvent.SAddr.String(), IPv4ConnectEvent.DAddr.String(), uint16(IPv4ConnectEvent.SPort), uint16(IPv4ConnectEvent.DPort)}
	mockEbpfTracker.handleConnection(IPv4ConnectEvent.Type, tuple, int(IPv4ConnectEvent.Pid), strconv.FormatUint(uint64(IPv4ConnectEvent.NetNS), 10))
	if !reflect.DeepEqual(mockEbpfTracker.openConnections[tuple], IPv4ConnectEbpfConnection) {
		t.Errorf("Connection mismatch connect event\nTarget connection:%v\nParsed connection:%v",
			IPv4ConnectEbpfConnection, mockEbpfTracker.openConnections[tuple])
	}

	tuple = fourTuple{IPv4ConnectCloseEvent.SAddr.String(), IPv4ConnectCloseEvent.DAddr.String(), uint16(IPv4ConnectCloseEvent.SPort), uint16(IPv4ConnectCloseEvent.DPort)}
	mockEbpfTracker.handleConnection(IPv4ConnectCloseEvent.Type, tuple, int(IPv4ConnectCloseEvent.Pid), strconv.FormatUint(uint64(IPv4ConnectCloseEvent.NetNS), 10))
	if len(mockEbpfTracker.openConnections) != 0 {
		t.Errorf("Connection mismatch close event\nConnection to close:%v",
			mockEbpfTracker.openConnections[tuple])
	}

	mockEbpfTracker = &EbpfTracker{
		readyToHandleConnections: true,
		dead: false,

		openConnections:   map[fourTuple]ebpfConnection{},
		closedConnections: []ebpfConnection{},
	}

	tuple = fourTuple{IPv4AcceptEvent.SAddr.String(), IPv4AcceptEvent.DAddr.String(), uint16(IPv4AcceptEvent.SPort), uint16(IPv4AcceptEvent.DPort)}
	mockEbpfTracker.handleConnection(IPv4AcceptEvent.Type, tuple, int(IPv4AcceptEvent.Pid), strconv.FormatUint(uint64(IPv4AcceptEvent.NetNS), 10))
	if !reflect.DeepEqual(mockEbpfTracker.openConnections[tuple], IPv4AcceptEbpfConnection) {
		t.Errorf("Connection mismatch connect event\nTarget connection:%v\nParsed connection:%v",
			IPv4AcceptEbpfConnection, mockEbpfTracker.openConnections[tuple])
	}

	tuple = fourTuple{IPv4AcceptCloseEvent.SAddr.String(), IPv4AcceptCloseEvent.DAddr.String(), uint16(IPv4AcceptCloseEvent.SPort), uint16(IPv4AcceptCloseEvent.DPort)}
	mockEbpfTracker.handleConnection(IPv4AcceptCloseEvent.Type, tuple, int(IPv4AcceptCloseEvent.Pid), strconv.FormatUint(uint64(IPv4AcceptCloseEvent.NetNS), 10))

	if len(mockEbpfTracker.openConnections) != 0 {
		t.Errorf("Connection mismatch close event\nConnection to close:%v",
			mockEbpfTracker.openConnections)
	}
}

func TestWalkConnections(t *testing.T) {
	var (
		cnt         int
		activeTuple = fourTuple{
			fromAddr: "",
			toAddr:   "",
			fromPort: 0,
			toPort:   0,
		}

		inactiveTuple = fourTuple{
			fromAddr: "",
			toAddr:   "",
			fromPort: 0,
			toPort:   0,
		}
	)
	mockEbpfTracker := &EbpfTracker{
		readyToHandleConnections: true,
		dead: false,
		openConnections: map[fourTuple]ebpfConnection{
			activeTuple: {
				tuple:            activeTuple,
				networkNamespace: "12345",
				incoming:         true,
				pid:              0,
			},
		},
		closedConnections: []ebpfConnection{
			{
				tuple:            inactiveTuple,
				networkNamespace: "12345",
				incoming:         false,
				pid:              0,
			},
		},
	}
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
		ServerIP          = net.IP("127.0.0.1")
		ClientIP          = net.IP("127.0.0.2")
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
	mockEbpfTracker := &EbpfTracker{
		readyToHandleConnections: true,
		dead:            false,
		openConnections: map[fourTuple]ebpfConnection{},
	}
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
	if !mockEbpfTracker.isDead() {
		t.Errorf("expected ebpfTracker to be set to dead after events with wrong order")
	}
}
