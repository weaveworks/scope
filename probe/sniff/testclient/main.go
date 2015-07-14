package main

import (
	"flag"
	"io"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func main() {
	var (
		device = flag.String("device", "eth0", "device to sniff")
	)
	flag.Parse()

	const (
		snaplen = 1024 * 1024
		promisc = true
		timeout = pcap.BlockForever
	)
	handle, err := pcap.OpenLive(*device, snaplen, promisc, timeout)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		time.Sleep(5 * time.Second)
		handle.Close()
	}()

	var (
		eth   layers.Ethernet
		ip4   layers.IPv4
		ip6   layers.IPv6
		tcp   layers.TCP
		udp   layers.UDP
		icmp4 layers.ICMPv4
		icmp6 layers.ICMPv6
	)
	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&eth, &ip4, &ip6, &tcp, &udp, &icmp4, &icmp6,
	)
	decoded := []gopacket.LayerType{}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			data, ci, err := handle.ZeroCopyReadPacketData()
			if err == io.EOF {
				log.Print("read: EOF")
				return
			}
			if err != nil {
				log.Printf("read: %v", err)
				continue
			}
			log.Println(ci.Timestamp.String())
			err = parser.DecodeLayers(data, &decoded)
			for _, t := range decoded {
				switch t {
				case layers.LayerTypeEthernet:
					log.Println(" Ethernet", eth.EthernetType, eth.SrcMAC, eth.DstMAC, eth.Length)
				case layers.LayerTypeIPv6:
					log.Println(" IP6", ip6.Version, ip6.SrcIP, ip6.DstIP, ip6.Length, ip6.TrafficClass)
				case layers.LayerTypeIPv4:
					log.Println(" IP4", ip4.Version, ip4.SrcIP, ip4.DstIP, ip4.Length, ip4.TTL, ip4.TOS)
				case layers.LayerTypeTCP:
					log.Println(" TCP", tcp.SrcPort, tcp.DstPort, tcp.Seq, tcp.Ack, tcp.Window)
				case layers.LayerTypeUDP:
					log.Println(" UDP", udp.SrcPort, udp.DstPort, udp.Length)
				case layers.LayerTypeICMPv4:
					log.Println(" ICMP4", icmp4.Id, icmp4.Seq)
				case layers.LayerTypeICMPv6:
					log.Println(" ICMP6")
				}
			}
			log.Println()
		}
	}()
	<-done
}
