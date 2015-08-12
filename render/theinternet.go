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
	result := report.Networks{}
	for _, md := range r.Host.NodeMetadatas {
		val, ok := md.Metadata[host.LocalNetworks]
		if !ok {
			continue
		}
		result = append(result, ParseNetworks(val)...)
	}
	return result
}

// ParseNetworks converts a string of space-separated CIDRs to a
// report.Networks.
func ParseNetworks(v string) report.Networks {
	var (
		nets = report.Networks{}
		set  = map[string]struct{}{}
	)
	for _, s := range strings.Fields(v) {
		_, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			continue
		}
		if _, ok := set[ipNet.String()]; !ok {
			nets = append(nets, ipNet)
			set[ipNet.String()] = struct{}{}
		}
	}
	return nets
}
