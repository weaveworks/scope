package endpoint

import (
	"bytes"
	"fmt"
	"math"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bluele/gcache"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const (
	bufSize              = 8 * 1024 * 1024 // 8MB
	maxReverseDNSrecords = 10000
)

// DNSSnooper is a snopper of DNS queries
type DNSSnooper struct {
	stop            chan struct{}
	pcapHandle      *pcap.Handle
	reverseDNSCache gcache.Cache
}

// NewDNSSnooper creates a new snooper of DNS queries
func NewDNSSnooper() (*DNSSnooper, error) {
	pcapHandle, err := newPcapHandle()
	if err != nil {
		return nil, err
	}
	reverseDNSCache := gcache.New(maxReverseDNSrecords).LRU().Build()

	s := &DNSSnooper{
		stop:            make(chan struct{}),
		pcapHandle:      pcapHandle,
		reverseDNSCache: reverseDNSCache,
	}
	go s.run()
	return s, nil
}

func newPcapHandle() (*pcap.Handle, error) {
	inactive, err := pcap.NewInactiveHandle("any")
	if err != nil {
		return nil, err
	}
	defer inactive.CleanUp()
	// pcap timeout blackmagic copied from Weave Net to reduce CPU consumption
	// see https://github.com/weaveworks/weave/commit/025315363d5ea8b8265f1b3ea800f24df2be51a4
	if err = inactive.SetTimeout(time.Duration(math.MaxInt64)); err != nil {
		return nil, err
	}
	if err = inactive.SetImmediateMode(true); err != nil {
		// If gopacket is compiled against an older pcap.h that
		// doesn't have pcap_set_immediate_mode, it supplies a dummy
		// definition that always returns PCAP_ERROR.  That becomes
		// "Generic error", which is not very helpful.  The real
		// pcap_set_immediate_mode never returns PCAP_ERROR, so this
		// turns it into a more informative message.
		if fmt.Sprint(err) == "Generic error" {
			return nil, fmt.Errorf("compiled against an old version of libpcap; please compile against libpcap-1.5.0 or later")
		}

		return nil, err
	}
	if err = inactive.SetBufferSize(bufSize); err != nil {
		return nil, err
	}
	pcapHandle, err := inactive.Activate()
	if err != nil {
		return nil, err
	}
	if err := pcapHandle.SetDirection(pcap.DirectionIn); err != nil {
		pcapHandle.Close()
		return nil, err
	}
	if err := pcapHandle.SetBPFFilter("inbound and port 53"); err != nil {
		pcapHandle.Close()
		return nil, err
	}

	return pcapHandle, nil
}

// CachedNamesForIP obtains the domains associated to an IP,
// obtained while snooping A-record queries
func (s *DNSSnooper) CachedNamesForIP(ip string) []string {
	result := []string{}
	if s == nil {
		return result
	}
	domains, err := s.reverseDNSCache.Get(ip)
	if err != nil {
		return result
	}

	for domain := range domains.(map[string]struct{}) {
		result = append(result, domain)
	}

	return result
}

// Stop makes the snooper stop inspecting DNS communications
func (s *DNSSnooper) Stop() {
	if s != nil {
		close(s.stop)
	}
}

func (s *DNSSnooper) run() {
	var (
		decodedLayers []gopacket.LayerType
		dns           layers.DNS
		udp           layers.UDP
		tcp           layers.TCP
		ip4           layers.IPv4
		ip6           layers.IPv6
		eth           layers.Ethernet
		sll           layers.LinuxSLL
	)

	// assumes that the "any" interface is being used (see https://wiki.wireshark.org/SLL)
	packetParser := gopacket.NewDecodingLayerParser(layers.LayerTypeLinuxSLL, &sll, &eth, &ip4, &ip6, &udp, &tcp, &dns)

	for {
		select {
		case <-s.stop:
			s.pcapHandle.Close()
			return
		default:
		}

		packet, _, err := s.pcapHandle.ZeroCopyReadPacketData()
		if err != nil {
			// TimeoutExpired is acceptable due to the Timeout black magic
			// on the handle
			if err != pcap.NextErrorTimeoutExpired {
				log.Errorf("DNSSnooper: error reading packet data: %s", err)
			}
			continue
		}

		if err := packetParser.DecodeLayers(packet, &decodedLayers); err != nil {
			log.Errorf("DNSSnooper: error decoding packet: %s", err)
			continue
		}

		for _, layerType := range decodedLayers {
			if layerType == layers.LayerTypeDNS {
				s.processDNSMessage(&dns)
			}
		}
	}
}

func (s *DNSSnooper) processDNSMessage(dns *layers.DNS) {

	// Only consider responses to singleton, A-record questions
	if !dns.QR || dns.ResponseCode != 0 || len(dns.Questions) != 1 {
		return
	}
	question := dns.Questions[0]
	if question.Type != layers.DNSTypeA || question.Class != layers.DNSClassIN {
		return
	}

	var (
		domainQueried = question.Name
		records       = append(dns.Answers, dns.Additionals...)
		ips           = map[string]struct{}{}
		alias         []byte
	)

	// Traverse records for a CNAME first since the DNS RFCs don't seem to guarantee it
	// appearing before its A-records
	for _, record := range records {
		if record.Type == layers.DNSTypeCNAME && record.Class == layers.DNSClassIN && bytes.Equal(domainQueried, record.Name) {
			alias = record.CNAME
			break
		}
	}

	// Finally, get the answer
	for _, record := range records {
		if record.Type != layers.DNSTypeA || record.Class != layers.DNSClassIN {
			continue
		}
		if bytes.Equal(domainQueried, record.Name) || (alias != nil && bytes.Equal(alias, record.Name)) {
			ips[record.IP.String()] = struct{}{}
		}
	}

	// Update cache
	newDomain := string(domainQueried)
	log.Debugf("DNSSnooper: caught DNS lookup: %s -> %v", newDomain, ips)
	for ip := range ips {
		if existingDomains, err := s.reverseDNSCache.Get(ip); err != nil {
			s.reverseDNSCache.Set(ip, map[string]struct{}{newDomain: {}})
		} else {
			// TODO: Be smarter about the expiration of entries with pre-existing associated domains
			existingDomains.(map[string]struct{})[newDomain] = struct{}{}
		}
	}
}
