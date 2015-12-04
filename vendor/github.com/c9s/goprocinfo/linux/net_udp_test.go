package linux

import (
	"reflect"
	"testing"
)

func TestReadNetUDP(t *testing.T) {

	udp, err := ReadNetUDPSockets("proc/net_udp", NetIPv4Decoder)

	if err != nil {
		t.Fatal("net udp read fail", err)
	}

	expected := &NetUDPSockets{
		Sockets: []NetUDPSocket{
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "127.0.0.1:53", RemoteAddress: "0.0.0.0:0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 11833, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "0.0.0.0:68", RemoteAddress: "0.0.0.0:0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 12616, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "192.168.1.111:123", RemoteAddress: "0.0.0.0:0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 18789, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "127.0.0.1:123", RemoteAddress: "0.0.0.0:0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 18788, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "0.0.0.0:123", RemoteAddress: "0.0.0.0:0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 18781, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "0.0.0.0:5353", RemoteAddress: "0.0.0.0:0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 109, Inode: 9025, SocketReferenceCount: 2,
				},
				Drops: 2237,
			},
		},
	}

	if !reflect.DeepEqual(udp, expected) {
		t.Errorf("not equal to expected %+v", expected)
	}

	t.Logf("%+v", udp)
}

func TestReadNetUDP6(t *testing.T) {

	udp, err := ReadNetUDPSockets("proc/net_udp6", NetIPv6Decoder)

	if err != nil {
		t.Fatal("net udp read fail", err)
	}

	expected := &NetUDPSockets{
		Sockets: []NetUDPSocket{
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "::1:123", RemoteAddress: ":::0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 840244, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "fe80::221:6aff:fea0:dd5e:123", RemoteAddress: ":::0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 840243, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "fe80::226:b9ff:fe1f:155e:123", RemoteAddress: ":::0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 840242, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "2a01:e35:2e12:f90:226:b9ff:fe1f:155e:123", RemoteAddress: ":::0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 840241, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: "2a01:e35:2e12:f90:adea:ed85:d1aa:4da6:123", RemoteAddress: ":::0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 840240, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: ":::123", RemoteAddress: ":::0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 840231, SocketReferenceCount: 2,
				},
				Drops: 8946,
			},
			NetUDPSocket{
				NetSocket: NetSocket{
					LocalAddress: ":::5353", RemoteAddress: ":::0", Status: 7,
					TxQueue: 0, RxQueue: 0, Uid: 109, Inode: 8944, SocketReferenceCount: 2,
				},
				Drops: 0,
			},
		},
	}

	if !reflect.DeepEqual(udp, expected) {
		t.Errorf("not equal to expected %+v", expected)
	}

	t.Logf("%+v", udp)
}
