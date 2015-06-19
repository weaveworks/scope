package render

import (
	"net"
	"strings"

	"github.com/weaveworks/scope/report"
)

// Networks represent a set of subnets local to a report.
type Networks []*net.IPNet

// LocalNetworks returns a superset of the networks (think: CIDRs) that are
// "local" from the perspective of each host represented in the report. It's
// used to determine which nodes in the report are "remote", i.e. outside of
// our infrastructure.
func LocalNetworks(r report.Report) Networks {
	var (
		result   = Networks{}
		networks = map[string]struct{}{}
	)

	for _, md := range r.Host.NodeMetadatas {
		val, ok := md["local_networks"]
		if !ok {
			continue
		}
		for _, s := range strings.Fields(val) {
			_, ipNet, err := net.ParseCIDR(s)
			if err != nil {
				continue
			}
			_, ok := networks[ipNet.String()]
			if !ok {
				result = append(result, ipNet)
				networks[ipNet.String()] = struct{}{}
			}
		}
	}
	return result
}

// Contains returns true if IP is in Networks.
func (n Networks) Contains(ip net.IP) bool {
	for _, net := range n {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}
