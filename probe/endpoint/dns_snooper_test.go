// +build linux,amd64 linux,ppc64le

package endpoint

import (
	"net"
	"testing"

	"github.com/google/gopacket/layers"
)

func TestprocessDNSMessageMultipleCNAME(t *testing.T) {
	domain := "dummy.com"
	question := layers.DNSQuestion{
		Name: []byte(domain),
		Type: layers.DNSTypeA,
	}

	ipAddressCNAME := "127.0.0.1"
	ipAddress := "127.0.1.1"
	answers := []layers.DNSResourceRecord{
		layers.DNSResourceRecord{
			Name:  []byte("api.dummy.com"),
			Type:  layers.DNSTypeCNAME,
			Class: layers.DNSClassIN,
		},
		layers.DNSResourceRecord{
			Name:  []byte("star.c10r.dummy.com"),
			Type:  layers.DNSTypeCNAME,
			Class: layers.DNSClassIN,
		},
		layers.DNSResourceRecord{
			Name:  []byte("dummy.com"),
			Type:  layers.DNSTypeA,
			Class: layers.DNSClassIN,
			IP:    net.ParseIP(ipAddress),
		},
		layers.DNSResourceRecord{
			Name:  []byte("star.c10r.dummy.com"),
			Type:  layers.DNSTypeA,
			Class: layers.DNSClassIN,
			IP:    net.ParseIP(ipAddressCNAME),
		},
	}

	dns := layers.DNS{
		ResponseCode: layers.DNSResponseCodeNoErr,
		Questions:    []layers.DNSQuestion{question},
		Answers:      answers,
	}

	snooper := &DNSSnooper{}

	snooper.processDNSMessage(&dns)

	existingDomains, err := snooper.reverseDNSCache.Get(ipAddressCNAME)

	if err != nil {
		t.Errorf("A domain should have been inserted for the given CNAME IP:%v", err)
	}

	if _, ok := existingDomains.(map[string]struct{})[domain]; !ok {
		t.Errorf("Domain %s should have been inserted", domain)
	}

	existingDomains, err = snooper.reverseDNSCache.Get(ipAddress)

	if err != nil {
		t.Errorf("A domain should have been inserted for the given IP:%v", err)
	}

	if _, ok := existingDomains.(map[string]struct{})[domain]; !ok {
		t.Errorf("Domain %s should have been inserted", domain)
	}
}
