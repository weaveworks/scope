package sniff

import (
	"io"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/weaveworks/scope/report"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Sniffer is a packet-sniffing reporter.
type Sniffer struct {
	hostID  string
	reports chan chan report.Report
	parser  *gopacket.DecodingLayerParser
	decoded []gopacket.LayerType
	eth     layers.Ethernet
	ip4     layers.IPv4
	ip6     layers.IPv6
	tcp     layers.TCP
	udp     layers.UDP
	icmp4   layers.ICMPv4
	icmp6   layers.ICMPv6
}

// New returns a new sniffing reporter that samples traffic by turning its
// packet capture facilities on and off. Note that the on and off durations
// represent a way to bound CPU burn. Effective sample rate needs to be
// calculated as (packets decoded / packets observed).
func New(hostID string, src gopacket.ZeroCopyPacketDataSource, on, off time.Duration) *Sniffer {
	s := &Sniffer{
		hostID:  hostID,
		reports: make(chan chan report.Report),
	}
	s.parser = gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&s.eth, &s.ip4, &s.ip6, &s.tcp, &s.udp, &s.icmp4, &s.icmp6,
	)
	go s.loop(src, on, off)
	return s
}

// Report implements the Reporter interface.
func (s *Sniffer) Report() (report.Report, error) {
	c := make(chan report.Report)
	s.reports <- c
	return <-c, nil
}

func (s *Sniffer) loop(src gopacket.ZeroCopyPacketDataSource, on, off time.Duration) {
	var (
		process = uint64(1)               // initially enabled
		total   = uint64(0)               // total packets seen
		count   = uint64(0)               // count of packets captured
		packets = make(chan Packet, 1024) // decoded packets
		rpt     = report.MakeReport()     // the report we build
		turnOn  = (<-chan time.Time)(nil) // signal to start capture (initially enabled)
		turnOff = time.After(on)          // signal to stop capture
		done    = make(chan struct{})     // when src is finished, we're done too
	)

	go func() {
		s.read(src, packets, &process, &total, &count)
		close(done)
	}()

	for {
		select {
		case p := <-packets:
			s.Merge(p, rpt)

		case <-turnOn:
			atomic.StoreUint64(&process, 1) // enable packet capture
			turnOn = nil                    // disable the on switch
			turnOff = time.After(on)        // enable the off switch

		case <-turnOff:
			atomic.StoreUint64(&process, 0) // disable packet capture
			turnOn = time.After(off)        // enable the on switch
			turnOff = nil                   // disable the off switch

		case c := <-s.reports:
			rpt.Sampling.Count = atomic.LoadUint64(&count)
			rpt.Sampling.Total = atomic.LoadUint64(&total)
			interpolateCounts(rpt)
			c <- rpt
			atomic.StoreUint64(&count, 0)
			atomic.StoreUint64(&total, 0)
			rpt = report.MakeReport()

		case <-done:
			return
		}
	}
}

// interpolateCounts compensates for sampling by artifically inflating counts
// throughout the report. It should be run once for each report, within the
// probe, before it gets emitted into the rest of the system.
func interpolateCounts(r report.Report) {
	rate := r.Sampling.Rate()
	if rate >= 1.0 {
		return
	}
	factor := 1.0 / rate
	for _, topology := range r.Topologies() {
		for _, emd := range topology.EdgeMetadatas {
			if emd.PacketCount != nil {
				*emd.PacketCount = uint64(float64(*emd.PacketCount) * factor)
			}
			if emd.ByteCount != nil {
				*emd.ByteCount = uint64(float64(*emd.ByteCount) * factor)
			}
		}
	}
}

// Packet is an intermediate, decoded form of a packet, with the information
// that the Scope data model cares about. Designed to decouple the packet data
// source loop, which should be as fast as possible, and the process of
// merging the packet information to a report, which may take some time and
// allocations.
type Packet struct {
	SrcIP, DstIP       string
	SrcPort, DstPort   string
	Network, Transport int // byte counts
}

func (s *Sniffer) read(src gopacket.ZeroCopyPacketDataSource, dst chan Packet, process, total, count *uint64) {
	var (
		p    Packet
		data []byte
		err  error
	)
	for {
		data, _, err = src.ZeroCopyReadPacketData()
		if err == io.EOF {
			return // done
		}
		if err != nil {
			log.Printf("sniffer: read: %v", err)
			continue
		}
		atomic.AddUint64(total, 1)
		if atomic.LoadUint64(process) == 0 {
			continue
		}

		if err := s.parser.DecodeLayers(data, &s.decoded); err != nil {
			// We'll always get an error when we encounter a layer type for
			// which we haven't configured a decoder.
		}
		for _, t := range s.decoded {
			switch t {
			case layers.LayerTypeEthernet:
				//

			case layers.LayerTypeICMPv4:
				p.Network += len(s.icmp4.Payload)

			case layers.LayerTypeICMPv6:
				p.Network += len(s.icmp6.Payload)

			case layers.LayerTypeIPv6:
				p.SrcIP = s.ip6.SrcIP.String()
				p.DstIP = s.ip6.DstIP.String()
				p.Network += len(s.ip6.Payload)

			case layers.LayerTypeIPv4:
				p.SrcIP = s.ip4.SrcIP.String()
				p.DstIP = s.ip4.DstIP.String()
				p.Network += len(s.ip4.Payload)

			case layers.LayerTypeTCP:
				p.SrcPort = strconv.Itoa(int(s.tcp.SrcPort))
				p.DstPort = strconv.Itoa(int(s.tcp.DstPort))
				p.Transport += len(s.tcp.Payload)

			case layers.LayerTypeUDP:
				p.SrcPort = strconv.Itoa(int(s.udp.SrcPort))
				p.DstPort = strconv.Itoa(int(s.udp.DstPort))
				p.Transport += len(s.udp.Payload)
			}
		}

		select {
		case dst <- p:
			atomic.AddUint64(count, 1)
		default:
			log.Printf("sniffer dropped packet")
		}
	}
}

// Merge puts the packet into the report.
func (s *Sniffer) Merge(p Packet, rpt report.Report) {
	// With a src and dst IP, we can add to the address topology.
	if p.SrcIP != "" && p.DstIP != "" {
		var (
			srcNodeID      = report.MakeAddressNodeID(s.hostID, p.SrcIP)
			dstNodeID      = report.MakeAddressNodeID(s.hostID, p.DstIP)
			edgeID         = report.MakeEdgeID(srcNodeID, dstNodeID)
			srcAdjacencyID = report.MakeAdjacencyID(srcNodeID)
		)
		rpt.Address.NodeMetadatas[srcNodeID] = report.MakeNodeMetadata()
		rpt.Address.NodeMetadatas[dstNodeID] = report.MakeNodeMetadata()

		emd := rpt.Address.EdgeMetadatas[edgeID]
		if emd.PacketCount == nil {
			emd.PacketCount = new(uint64)
		}
		*emd.PacketCount++
		if emd.ByteCount == nil {
			emd.ByteCount = new(uint64)
		}
		*emd.ByteCount += uint64(p.Network)
		rpt.Address.EdgeMetadatas[edgeID] = emd

		rpt.Address.Adjacency[srcAdjacencyID] = rpt.Address.Adjacency[srcAdjacencyID].Add(dstNodeID)
	}

	// With a src and dst IP and port, we can add to the endpoints.
	if p.SrcIP != "" && p.DstIP != "" && p.SrcPort != "" && p.DstPort != "" {
		var (
			srcNodeID      = report.MakeEndpointNodeID(s.hostID, p.SrcIP, p.SrcPort)
			dstNodeID      = report.MakeEndpointNodeID(s.hostID, p.DstIP, p.DstPort)
			edgeID         = report.MakeEdgeID(srcNodeID, dstNodeID)
			srcAdjacencyID = report.MakeAdjacencyID(srcNodeID)
		)
		rpt.Endpoint.NodeMetadatas[srcNodeID] = report.MakeNodeMetadata()
		rpt.Endpoint.NodeMetadatas[dstNodeID] = report.MakeNodeMetadata()

		emd := rpt.Endpoint.EdgeMetadatas[edgeID]
		if emd.PacketCount == nil {
			emd.PacketCount = new(uint64)
		}
		*emd.PacketCount++
		if emd.ByteCount == nil {
			emd.ByteCount = new(uint64)
		}
		*emd.ByteCount += uint64(p.Transport)
		rpt.Endpoint.EdgeMetadatas[edgeID] = emd

		rpt.Endpoint.Adjacency[srcAdjacencyID] = rpt.Endpoint.Adjacency[srcAdjacencyID].Add(dstNodeID)
	}
}
