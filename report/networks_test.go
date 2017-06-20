package report_test

import (
	"net"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestContains(t *testing.T) {
	networks := report.Networks([]*net.IPNet{
		mustParseCIDR("10.0.0.1/8"),
		mustParseCIDR("192.168.1.1/24"),
	})

	if networks.Contains(net.ParseIP("52.52.52.52")) {
		t.Errorf("52.52.52.52 not in %v", networks)
	}

	if !networks.Contains(net.ParseIP("10.0.0.1")) {
		t.Errorf("10.0.0.1 in %v", networks)
	}
}

func mustParseCIDR(s string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return ipNet
}
