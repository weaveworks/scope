package render

import (
	"net"
	"strings"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

var (
	// ServiceNodeIDPrefix is how the ID all service pseudo nodes begin
	ServiceNodeIDPrefix = "service-"

	// Correspondence between hostnames and the service id they are part of
	knownServicesSuffixes = []string{
		// See http://docs.aws.amazon.com/general/latest/gr/rande.html for fainer grained
		// details
		"amazonaws.com",
		"googleapis.com",
	}
)

func isKnownService(hostname string) bool {
	for _, suffix := range knownServicesSuffixes {
		if strings.HasSuffix(hostname, suffix) {
			return true
		}
	}
	return false
}

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
