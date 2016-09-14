package render

import (
	"net"
	"regexp"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

var (
	// ServiceNodeIDPrefix is how the ID all service pseudo nodes begin
	ServiceNodeIDPrefix = "service-"

	// KnownServicesForHumans contains a human-readable format of the service Ids
	KnownServicesForHumans = map[string]string{
		"aws-dynamo": "AWS Dynamo",
		"aws-s3":     "AWS S3",
	}

	// Correspondence between hostnames and the service id they are part of
	knownServicesMatchers = map[*regexp.Regexp]string{
		regexp.MustCompile(`dynamodb.[^.]+.amazonaws.com`): "aws-dynamo",
		regexp.MustCompile(`s3-[^.]+.amazonaws.com`):       "aws-s3",
	}
)

func lookupKnownService(hostname string) (string, bool) {
	for re, id := range knownServicesMatchers {
		if re.MatchString(hostname) {
			return id, true
		}
	}
	return "", false
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
