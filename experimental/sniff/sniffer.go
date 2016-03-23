package sniff

import (
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/weaveworks/scope/report"
)

/*
It's important to adjust GOMAXPROCS when using sniffer:

if *captureEnabled {
	var sniffers int
	for _, iface := range strings.Split(*captureInterfaces, ",") {
		source, err := sniff.NewSource(iface)
		if err != nil {
			log.Printf("warning: %v", err)
			continue
		}
		defer source.Close()
		log.Printf("capturing packets on %s", iface)
		reporters = append(reporters, sniff.New(hostID, localNets, source, *captureOn, *captureOff))
		sniffers++
	}
	// Packet capture can block OS threads on Linux, so we need to provide
	// sufficient overhead in GOMAXPROCS.
	if have, want := runtime.GOMAXPROCS(-1), (sniffers + 1); have < want {
		runtime.GOMAXPROCS(want)
	}
}

func interfaces() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print(err)
		return ""
	}
	a := make([]string, 0, len(ifaces))
	for _, iface := range ifaces {
		a = append(a, iface.Name)
	}
	return strings.Join(a, ",")
}

Also, the capture on/off sampling methodology is probably not worth keeping.
*/

// Sniffer is a packet-sniffing reporter.
type Sniffer struct {
	hostID    string
	localNets report.Networks
	reports   chan chan report.Report
	parser    *gopacket.DecodingLayerParser
	decoded   []gopacket.LayerType
	eth       layers.Ethernet
	ip4       layers.IPv4
	ip6       layers.IPv6
	tcp       layers.TCP
	udp       layers.UDP
	icmp4     layers.ICMPv4
	icmp6     layers.ICMPv6
}

// New returns a new sniffing reporter that samples traffic by turning its
// packet capture facilities on and off. Note that the on and off durations
// represent a way to bound CPU burn. Effective sample rate needs to be
// calculated as (packets decoded / packets observed).
func New(hostID string, localNets report.Networks, src gopacket.ZeroCopyPacketDataSource, on, off time.Duration) *Sniffer {
	s := &Sniffer{
		hostID:    hostID,
		localNets: localNets,
		reports:   make(chan chan report.Report),
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

	// As a special case, if our off duty cycle is zero, i.e. 100% sample
	// rate, we simply disable the turn-off signal channel.
	if off == 0 {
		turnOff = nil
	}

	go func() {
		s.read(src, packets, &process, &total, &count)
		close(done)
	}()

	for {
		select {
		case p := <-packets:
			s.Merge(p, &rpt)

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

// interpolateCounts compensates for sampling by artificially inflating counts
// throughout the report. It should be run once for each report, within the
// probe, before it gets emitted into the rest of the system.
func interpolateCounts(r report.Report) {
	rate := r.Sampling.Rate()
	if rate >= 1.0 {
		return
	}
	factor := 1.0 / rate
	for _, topology := range r.Topologies() {
		for _, nmd := range topology.Nodes {
			nmd.Edges.ForEach(func(_ string, emd report.EdgeMetadata) {
				if emd.EgressPacketCount != nil {
					*emd.EgressPacketCount = uint64(float64(*emd.EgressPacketCount) * factor)
				}
				if emd.IngressPacketCount != nil {
					*emd.IngressPacketCount = uint64(float64(*emd.IngressPacketCount) * factor)
				}
				if emd.EgressByteCount != nil {
					*emd.EgressByteCount = uint64(float64(*emd.EgressByteCount) * factor)
				}
				if emd.IngressByteCount != nil {
					*emd.IngressByteCount = uint64(float64(*emd.IngressByteCount) * factor)
				}
			})
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
		var p Packet
		for _, t := range s.decoded {
			switch t {
			case layers.LayerTypeEthernet:
				//

			case layers.LayerTypeICMPv4:
				p.Network += len(s.icmp4.Payload)

			case layers.LayerTypeICMPv6:
				p.Network += len(s.icmp6.Payload)

			case layers.LayerTypeIPv4:
				p.SrcIP = s.ip4.SrcIP.String()
				p.DstIP = s.ip4.DstIP.String()
				p.Network += len(s.ip4.Payload)

			case layers.LayerTypeIPv6:
				p.SrcIP = s.ip6.SrcIP.String()
				p.DstIP = s.ip6.DstIP.String()
				p.Network += len(s.ip6.Payload)

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
//
// Note that, for the moment, we encode bidirectional traffic as ingress and
// egress traffic on a single edge whose src is local and dst is remote. That
// is, if we see a packet from the remote addr 9.8.7.6 to the local addr
// 1.2.3.4, we apply it as *ingress* on the edge (1.2.3.4 -> 9.8.7.6).
func (s *Sniffer) Merge(p Packet, rpt *report.Report) {
	if p.SrcIP == "" || p.DstIP == "" {
		return
	}

	// One end of the traffic has to be local. Otherwise, we don't know how to
	// construct the edge.
	//
	// If we need to get around this limitation, we may be able to change the
	// semantics of the report, and allow the src side of edges to be from
	// anywhere. But that will have ramifications throughout Scope (read: it
	// may violate implicit invariants) and needs to be thought through.
	var (
		srcLocal   = s.localNets.Contains(net.ParseIP(p.SrcIP))
		dstLocal   = s.localNets.Contains(net.ParseIP(p.DstIP))
		localIP    string
		remoteIP   string
		localPort  string
		remotePort string
		egress     bool
	)
	switch {
	case srcLocal && !dstLocal:
		localIP, localPort, remoteIP, remotePort, egress = p.SrcIP, p.SrcPort, p.DstIP, p.DstPort, true
	case !srcLocal && dstLocal:
		localIP, localPort, remoteIP, remotePort, egress = p.DstIP, p.DstPort, p.SrcIP, p.SrcPort, false
	case srcLocal && dstLocal:
		localIP, localPort, remoteIP, remotePort, egress = p.SrcIP, p.SrcPort, p.DstIP, p.DstPort, true // loopback
	case !srcLocal && !dstLocal:
		log.Printf("sniffer ignoring remote-to-remote (%s -> %s) traffic", p.SrcIP, p.DstIP)
		return
	}

	addAdjacency := func(t report.Topology, srcNodeID, dstNodeID string) report.Topology {
		result := t.AddNode(srcNodeID, report.MakeNode().WithAdjacent(dstNodeID))
		result = result.AddNode(dstNodeID, report.MakeNode())
		return result
	}

	// If we have ports, we can add to the endpoint topology, too.
	if p.SrcPort != "" && p.DstPort != "" {
		var (
			srcNodeID = report.MakeEndpointNodeID(s.hostID, localIP, localPort)
			dstNodeID = report.MakeEndpointNodeID(s.hostID, remoteIP, remotePort)
		)

		rpt.Endpoint = addAdjacency(rpt.Endpoint, srcNodeID, dstNodeID)

		node := rpt.Endpoint.Nodes[srcNodeID]
		emd, _ := node.Edges.Lookup(dstNodeID)
		if egress {
			if emd.EgressPacketCount == nil {
				emd.EgressPacketCount = new(uint64)
			}
			*emd.EgressPacketCount++
			if emd.EgressByteCount == nil {
				emd.EgressByteCount = new(uint64)
			}
			*emd.EgressByteCount += uint64(p.Transport)
		} else {
			if emd.IngressPacketCount == nil {
				emd.IngressPacketCount = new(uint64)
			}
			*emd.IngressPacketCount++
			if emd.IngressByteCount == nil {
				emd.IngressByteCount = new(uint64)
			}
			*emd.IngressByteCount += uint64(p.Transport)
		}
		rpt.Endpoint.Nodes[srcNodeID] = node.WithEdge(dstNodeID, emd)
	}
}
