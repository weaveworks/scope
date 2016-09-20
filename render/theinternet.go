package render

import (
	"net"
	"regexp"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

var (
	// ServiceNodeIDPrefix is how the ID of all service pseudo nodes begin
	ServiceNodeIDPrefix = "service-"

	knownServiceMatchers = []*regexp.Regexp{
		// See http://docs.aws.amazon.com/general/latest/gr/rande.html for fainer grained
		// details
		regexp.MustCompile(`^.+\.amazonaws\.com$`),
		regexp.MustCompile(`^.+\.googleapis\.com$`),
	}

	knownServiceExcluders = []*regexp.Regexp{
		// We exclude ec2 machines because they are too generic
		// and having separate nodes for them makes visualizations worse
		regexp.MustCompile(`^ec2.*\.amazonaws\.com$`),
	}
)

// TODO: Make it user-customizable https://github.com/weaveworks/scope/issues/1876
func isKnownService(hostname string) bool {
	foundMatch := false
	for _, matcher := range knownServiceMatchers {
		if matcher.MatchString(hostname) {
			foundMatch = true
			break
		}
	}
	if !foundMatch {
		return false
	}

	for _, excluder := range knownServiceExcluders {
		if excluder.MatchString(hostname) {
			return false
		}
	}
	return true
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
