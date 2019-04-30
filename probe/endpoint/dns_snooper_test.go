// +build linux,amd64 linux,ppc64le

package endpoint

import (
	"net"
	"testing"

	"github.com/bluele/gcache"
	"github.com/google/gopacket/layers"
)

func TestProcessDNSMessageMultipleCNAME(t *testing.T) {
	domain := "dummy.com"
	question := layers.DNSQuestion{
		Name:  []byte(domain),
		Type:  layers.DNSTypeA,
		Class: layers.DNSClassIN,
	}

	ipAddressCNAME := "127.0.0.1"
	answers := []layers.DNSResourceRecord{
		layers.DNSResourceRecord{
			Name:  []byte("api.dummy.com"),
			CNAME: []byte("api.dummy.com"),
			Type:  layers.DNSTypeCNAME,
			Class: layers.DNSClassIN,
		},
		layers.DNSResourceRecord{
			Name:  []byte("star.c10r.dummy.com"),
			CNAME: []byte("star.c10r.dummy.com"),
			Type:  layers.DNSTypeCNAME,
			Class: layers.DNSClassIN,
		},
		layers.DNSResourceRecord{
			Name:  []byte("star.c10r.dummy.com"),
			Type:  layers.DNSTypeA,
			Class: layers.DNSClassIN,
			IP:    net.ParseIP(ipAddressCNAME),
		},
	}

	dns := layers.DNS{
		QR:           true,
		ResponseCode: layers.DNSResponseCodeNoErr,
		Questions:    []layers.DNSQuestion{question},
		Answers:      answers,
	}

	snooper := &DNSSnooper{
		reverseDNSCache: gcache.New(4).LRU().Build(),
	}

	snooper.processDNSMessage(&dns)

	existingDomains, err := snooper.reverseDNSCache.Get(ipAddressCNAME)

	if err != nil {
		t.Errorf("A domain should have been inserted for the given CNAME IP:%v", err)
	}

	if _, ok := existingDomains.(map[string]struct{})[domain]; !ok {
		t.Errorf("Domain %s should have been inserted", domain)
	}
}
