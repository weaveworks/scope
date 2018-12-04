package conntrack

import (
	"net"
)

// localIPs gives all IPs we consider local.
func localIPs() map[string]struct{} {
	var l = map[string]struct{}{}
	if localNets, err := net.InterfaceAddrs(); err == nil {
		// Not all networks are IP networks.
		for _, localNet := range localNets {
			if net, ok := localNet.(*net.IPNet); ok {
				l[net.IP.String()] = struct{}{}
			}
		}
	}
	return l
}
