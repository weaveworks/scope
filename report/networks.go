package report

import (
	"net"
	"strings"
)

// Networks represent a set of subnets
type Networks []*net.IPNet

// ParseNetworks converts a string of space-separated CIDRs to a Networks.
func ParseNetworks(v string) Networks {
	set := map[string]struct{}{}
	for _, s := range strings.Fields(v) {
		_, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			continue
		}
		set[ipNet.String()] = struct{}{}
	}
	nets := Networks{}
	for s := range set {
		_, ipNet, _ := net.ParseCIDR(s)
		nets = append(nets, ipNet)
	}
	return nets
}

// Interface is exported for testing.
type Interface interface {
	Addrs() ([]net.Addr, error)
}

// Variables exposed for testing.
// TODO this design is broken, make it consistent with probe networks.
var (
	LocalNetworks       = Networks{}
	InterfaceByNameStub = func(name string) (Interface, error) { return net.InterfaceByName(name) }
)

// Contains returns true if IP is in Networks.
func (n Networks) Contains(ip net.IP) bool {
	for _, net := range n {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

// AddLocalBridge records the subnet address associated with the bridge name
// supplied, such that MakeAddressNodeID will scope addresses in this subnet
// as local.
func AddLocalBridge(name string) error {
	inf, err := InterfaceByNameStub(name)
	if err != nil {
		return err
	}

	addrs, err := inf.Addrs()
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		_, network, err := net.ParseCIDR(addr.String())
		if err != nil {
			return err
		}

		if network == nil {
			continue
		}

		LocalNetworks = append(LocalNetworks, network)
	}

	return nil
}
