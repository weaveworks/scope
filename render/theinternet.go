package render

import (
	"net"
	"strings"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
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

	for _, md := range r.Host.NodeMetadatas {
		val, ok := md[host.LocalNetworks]
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
