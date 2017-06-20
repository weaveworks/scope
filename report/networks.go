package report

import (
	"net"
	"strings"

	"github.com/k-sone/critbitgo"
)

// Networks represent a set of subnets
type Networks struct{ *critbitgo.Net }

// LocalNetworks helps in determining which addresses a probe reports
// as being host-scoped.
//
// TODO this design is broken, make it consistent with probe networks.
var LocalNetworks = NewNetworks()

// Create datastructure for set of subnets.
func NewNetworks() Networks {
	return Networks{critbitgo.NewNet()}
}

// Add a network.
func (n Networks) Add(ipnet *net.IPNet) error {
	return n.Net.Add(ipnet, struct{}{})
}

// Add a network, represented as CIDR.
func (n Networks) AddCIDR(cidr string) error {
	return n.Net.AddCIDR(cidr, struct{}{})
}

// Contains returns true if IP is in Networks.
func (n Networks) Contains(ip net.IP) bool {
	network, _, _ := n.MatchIP(ip)
	return network != nil
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

	for _, ipnet := range ipv4Nets(addrs) {
		LocalNetworks.Add(ipnet)
	}

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
	nets := []*net.IPNet{}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			nets = append(nets, ipnet)
		}
	}
	return nets
}
