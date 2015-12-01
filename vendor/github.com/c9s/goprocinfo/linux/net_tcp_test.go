package linux

import (
	"reflect"
	"testing"
)

func TestReadNetTCP(t *testing.T) {

	tcp, err := ReadNetTCPSockets("proc/net_tcp", NetIPv4Decoder)

	if err != nil {
		t.Fatal("net tcp read fail", err)
	}

	expected := &NetTCPSockets{
		Sockets: []NetTCPSocket{
			NetTCPSocket{
				NetSocket: NetSocket{
					LocalAddress: "127.0.0.1:8080", RemoteAddress: "0.0.0.0:0", Status: 10,
					TxQueue: 0, RxQueue: 0, Uid: 1000, Inode: 569261, SocketReferenceCount: 1,
				},
				RetransmitTimeout: 100, PredictedTick: 0, AckQuick: 0, AckPingpong: false,
				SendingCongestionWindow: 10, SlowStartSizeThreshold: -1,
			},
			NetTCPSocket{
				NetSocket: NetSocket{
					LocalAddress: "0.0.0.0:80", RemoteAddress: "0.0.0.0:0", Status: 10,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 4609, SocketReferenceCount: 1,
				},
				RetransmitTimeout: 100, PredictedTick: 0, AckQuick: 0, AckPingpong: false,
				SendingCongestionWindow: 10, SlowStartSizeThreshold: -1,
			},
			NetTCPSocket{
				NetSocket: NetSocket{
					LocalAddress: "0.0.0.0:22", RemoteAddress: "0.0.0.0:0", Status: 10,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 420553, SocketReferenceCount: 1,
				},
				RetransmitTimeout: 100, PredictedTick: 0, AckQuick: 0, AckPingpong: false,
				SendingCongestionWindow: 10, SlowStartSizeThreshold: -1,
			},
			NetTCPSocket{
				NetSocket: NetSocket{
					LocalAddress: "10.0.7.21:22", RemoteAddress: "10.0.251.11:53280", Status: 1,
					TxQueue: 96, RxQueue: 0, Uid: 0, Inode: 582338, SocketReferenceCount: 4,
				},
				RetransmitTimeout: 29, PredictedTick: 4, AckQuick: 13, AckPingpong: true,
				SendingCongestionWindow: 10, SlowStartSizeThreshold: -1,
			},
		},
	}

	if !reflect.DeepEqual(tcp, expected) {
		t.Errorf("not equal to expected %+v", expected)
	}

	t.Logf("%+v", tcp)

}

func TestReadNetTCP6(t *testing.T) {

	tcp, err := ReadNetTCPSockets("proc/net_tcp6", NetIPv6Decoder)

	if err != nil {
		t.Fatal("net tcp read fail", err)
	}

	expected := &NetTCPSockets{
		Sockets: []NetTCPSocket{
			NetTCPSocket{
				NetSocket: NetSocket{
					LocalAddress: ":::22", RemoteAddress: ":::0", Status: 10,
					TxQueue: 0, RxQueue: 0, Uid: 0, Inode: 420555, SocketReferenceCount: 1,
				},
				RetransmitTimeout: 100, PredictedTick: 0, AckQuick: 0, AckPingpong: false,
				SendingCongestionWindow: 2, SlowStartSizeThreshold: -1,
			},
		},
	}

	if !reflect.DeepEqual(tcp, expected) {
		t.Errorf("not equal to expected %+v", expected)
	}

	t.Logf("%+v", tcp)

}
