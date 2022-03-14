//go:build darwin || arm || arm64 || s390x
// +build darwin arm arm64 s390x

// Cross-compiling the snooper requires having pcap binaries,
// let's disable it for now.
// See http://stackoverflow.com/questions/31648793/go-programming-cross-compile-for-revel-framework

package endpoint

// DNSSnooper is a snopper of DNS queries
type DNSSnooper struct{}

// NewDNSSnooper creates a new snooper of DNS queries
func NewDNSSnooper() (*DNSSnooper, error) {
	return nil, nil
}

// CachedNamesForIP obtains the domains associated to an IP,
// obtained while snooping A-record queries
func (s *DNSSnooper) CachedNamesForIP(ip string) []string {
	return []string{}
}

// Stop makes the snooper stop inspecting DNS communications
func (s *DNSSnooper) Stop() {
}
