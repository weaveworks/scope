package render

import (
	"net"
	"regexp"
	"strings"

	"camlistore.org/pkg/lru"

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
		`weave\.works`,             // Weave Cloud
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

	// Memoization for isKnownService.
	//
	// The 10000 comes from the observation that large reports contain
	// hundreds of names, and in a multi-tenant context we want to be
	// able to render a few dozen reports concurrently. Also, unlike
	// memoization in the reducers, which is keyed on reports, this
	// cache is effective when rendering multiple reports from the
	// same cluster of probes, e.g. from different points in time,
	// since names tend to change infrequently.
	//
	// Since names are generally <50 bytes, this shouldn't weight in
	// at more than a few MB of memory.
	knownServiceCache = lru.New(10000)
)

func purgeKnownServiceCache() {
	knownServiceCache = lru.New(10000)
}

// TODO: Make it user-customizable https://github.com/weaveworks/scope/issues/1876
// NB: this is a hotspot in rendering performance.
func isKnownService(hostname string) bool {
	if v, ok := knownServiceCache.Get(hostname); ok {
		return v.(bool)
	}

	known := knownServiceMatcher.MatchString(hostname) && !knownServiceExcluder.MatchString(hostname)
	knownServiceCache.Add(hostname, known)

	return known
}

// LocalNetworks returns a superset of the networks (think: CIDRs) that are
// "local" from the perspective of each host represented in the report. It's
// used to determine which nodes in the report are "remote", i.e. outside of
// our infrastructure.
func LocalNetworks(r report.Report) report.Networks {
	networks := report.MakeNetworks()

	for _, topology := range []report.Topology{r.Host, r.Overlay} {
		for _, md := range topology.Nodes {
			nets, _ := md.Sets.Lookup(report.HostLocalNetworks)
			for _, s := range nets {
				networks.AddCIDR(s)
			}
		}
	}
	if extra := kubeServiceNetwork(r.Service); extra != nil {
		networks.Add(extra)
	}
	return networks
}

// FIXME: Hideous hack to remove persistent-connection edges to
// virtual service IPs attributed to the internet. The global
// service-cluster-ip-range is not exposed by the API server (see
// https://github.com/kubernetes/kubernetes/issues/25533), so instead
// we synthesise it by computing the smallest network that contains
// all service IPs. That network may be smaller than the actual range
// but that is ok, since in the end all we care about is that it
// contains all the service IPs.
//
// The right way of fixing this is performing DNAT mapping on
// persistent connections for which we don't have a robust solution
// (see https://github.com/weaveworks/scope/issues/1491).
func kubeServiceNetwork(services report.Topology) *net.IPNet {
	serviceIPs := make([]net.IP, 0, len(services.Nodes))
	for _, md := range services.Nodes {
		serviceIP, _ := md.Latest.Lookup(report.KubernetesIP)
		if ip := net.ParseIP(serviceIP).To4(); ip != nil {
			serviceIPs = append(serviceIPs, ip)
		}
	}
	return report.ContainingIPv4Network(serviceIPs)
}
