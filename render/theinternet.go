package render

import (
	"net"
	"regexp"
	"strings"

	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/report"
)

var (
	// ServiceNodeIDPrefix is how the ID of all service pseudo nodes begin
	ServiceNodeIDPrefix = "service-"

	knownServiceMatcher = regexp.MustCompile(`^.+\.(` + strings.Join([]string{
		// See http://docs.aws.amazon.com/general/latest/gr/rande.html
		// for finer grained details
		`amazonaws\.com`,
		`googleapis\.com`,
		`core\.windows\.net`,       // Azure Storage - Blob, Tables, Files & Queues
		`servicebus\.windows\.net`, // Azure Service Bus
		`azure-api\.net`,           // Azure API Management
		`onmicrosoft\.com`,         // Azure Active Directory
		`cloudapp\.azure\.com`,     // Azure IaaS
		`database\.windows\.net`,   // Azure SQL DB
		`documents\.azure\.com`,    // Azure DocumentDB/CosmosDB
	}, `|`) + `)$`)

	knownServiceExcluder = regexp.MustCompile(`^(` + strings.Join([]string{
		// We exclude ec2 machines because they are too generic
		// and having separate nodes for them makes visualizations worse
		`ec2.*\.amazonaws\.com`,
	}, `|`) + `)$`)
)

// TODO: Make it user-customizable https://github.com/weaveworks/scope/issues/1876
// NB: this is a hotspot in rendering performance.
func isKnownService(hostname string) bool {
	return knownServiceMatcher.MatchString(hostname) && !knownServiceExcluder.MatchString(hostname)
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
