package sniff

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// Source describes a packet data source that can be terminated.
type Source interface {
	gopacket.ZeroCopyPacketDataSource
	Close()
}

const (
	snaplen = 65535
	promisc = true
	timeout = pcap.BlockForever
)

// NewSource returns a live packet data source via the passed device
// (interface).
func NewSource(device string) (Source, error) {
	return pcap.OpenLive(device, snaplen, promisc, timeout)
}
