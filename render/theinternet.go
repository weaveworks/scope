package render

import (
	"net"

	"$GITHUB_URI/probe/host"
	"$GITHUB_URI/report"
)

// LocalNetworks returns a superset of the networks (think: CIDRs) that are
// "local" from the perspective of each host represented in the report. It's
// used to determine which nodes in the report are "remote", i.e. outside of
// our infrastructure.
func LocalNetworks(r report.Report) report.Networks {
	var (
		result   = report.Networks{}
		networks = map[string]struct{}{}
	)

	for _, topology := range []report.Topology{r.Host, r.Overlay} {
		for _, md := range topology.Nodes {
			nets, _ := md.Sets.Lookup(host.LocalNetworks)
			for _, s := range nets {
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
	}
	return result
}
