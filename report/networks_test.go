package report_test

import (
	"net"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
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
