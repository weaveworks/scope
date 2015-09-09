package report_test

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
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

func TestParseNetworks(t *testing.T) {
	var (
		bignetStr   = "10.1.0.1/16"
		smallnetStr = "5.6.7.8/32"
		bignet      = mustParseCIDR(bignetStr)
		smallnet    = mustParseCIDR(smallnetStr)
	)
	for _, tc := range []struct {
		input string
		want  report.Networks
	}{
		{"", report.Networks{}},
		{fmt.Sprintf("%s", bignetStr), report.Networks([]*net.IPNet{bignet})},
		{fmt.Sprintf("%s %s", bignetStr, bignetStr), report.Networks([]*net.IPNet{bignet})},
		{fmt.Sprintf("%s foo %s oops  %s", smallnetStr, smallnetStr, smallnetStr), report.Networks([]*net.IPNet{smallnet})},
	} {
		if want, have := tc.want, report.ParseNetworks(tc.input); !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}

func mustParseCIDR(s string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return ipNet
}

type mockInterface struct {
	addrs []net.Addr
}

type mockAddr string

func (m mockInterface) Addrs() ([]net.Addr, error) {
	return m.addrs, nil
}

func (m mockAddr) Network() string {
	return "ip+net"
}

func (m mockAddr) String() string {
	return string(m)
}

func TestAddLocal(t *testing.T) {
	oldInterfaceByNameStub := report.InterfaceByNameStub
	defer func() { report.InterfaceByNameStub = oldInterfaceByNameStub }()

	report.InterfaceByNameStub = func(name string) (report.Interface, error) {
		return mockInterface{[]net.Addr{mockAddr("52.53.54.55/16")}}, nil
	}

	err := report.AddLocalBridge("foo")
	if err != nil {
		t.Errorf("%v", err)
	}

	want := report.Networks([]*net.IPNet{mustParseCIDR("52.53.54.55/16")})
	have := report.LocalNetworks

	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
