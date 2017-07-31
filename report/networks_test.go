package report_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/scope/report"
)

func TestContains(t *testing.T) {
	networks := report.MakeNetworks()
	for _, cidr := range []string{"10.0.0.1/8", "192.168.1.1/24"} {
		if err := networks.AddCIDR(cidr); err != nil {
			panic(err)
		}
	}

	if networks.Contains(net.ParseIP("52.52.52.52")) {
		t.Errorf("52.52.52.52 not in %v", networks)
	}

	if !networks.Contains(net.ParseIP("10.0.0.1")) {
		t.Errorf("10.0.0.1 in %v", networks)
	}
}

func TestContainingIPv4Network(t *testing.T) {
	assert.Nil(t, containingIPv4Networks([]string{}))
	assert.Equal(t, "10.0.0.1/32", containingIPv4Networks([]string{"10.0.0.1"}).String())
	assert.Equal(t, "10.0.0.0/17", containingIPv4Networks([]string{"10.0.0.1", "10.0.2.55", "10.0.106.48"}).String())
	assert.Equal(t, "0.0.0.0/0", containingIPv4Networks([]string{"10.0.0.1", "192.168.0.1"}).String())
}

func containingIPv4Networks(ipstrings []string) *net.IPNet {
	ips := make([]net.IP, len(ipstrings))
	for i, ip := range ipstrings {
		ips[i] = net.ParseIP(ip).To4()
	}
	return report.ContainingIPv4Network(ips)
}
