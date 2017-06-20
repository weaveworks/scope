package report

import (
	"net"
	"strings"
)

// Networks represent a set of subnets
type Networks []*net.IPNet

// LocalNetworks helps in determining which addresses a probe reports
// as being host-scoped.
//
// TODO this design is broken, make it consistent with probe networks.
var LocalNetworks = Networks{}

// Contains returns true if IP is in Networks.
func (n Networks) Contains(ip net.IP) bool {
	for _, net := range n {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

// LocalAddresses returns a list of the local IP addresses.
func LocalAddresses() ([]net.IP, error) {
	result := []net.IP{}

	infs, err := net.Interfaces()
	if err != nil {
		return []net.IP{}, err
	}

	for _, inf := range infs {
		if strings.HasPrefix(inf.Name, "veth") ||
			strings.HasPrefix(inf.Name, "docker") ||
			strings.HasPrefix(inf.Name, "lo") {
			continue
		}

		addrs, err := inf.Addrs()
		if err != nil {
			return []net.IP{}, err
		}

		for _, ipnet := range ipv4Nets(addrs) {
			result = append(result, ipnet.IP)
		}
	}

	return result, nil
}

// AddLocalBridge records the subnet address associated with the bridge name
// supplied, such that MakeAddressNodeID will scope addresses in this subnet
// as local.
func AddLocalBridge(name string) error {
	inf, err := net.InterfaceByName(name)
	if err != nil {
		return err
	}

	addrs, err := inf.Addrs()
	if err != nil {
		return err
	}

	LocalNetworks = ipv4Nets(addrs)

	return nil
}

// GetLocalNetworks returns all the local networks.
func GetLocalNetworks() ([]*net.IPNet, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	return ipv4Nets(addrs), nil
}

func ipv4Nets(addrs []net.Addr) []*net.IPNet {
	nets := Networks{}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			nets = append(nets, ipnet)
		}
	}
	return nets
}
