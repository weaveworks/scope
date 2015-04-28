package main

import (
	"net"
)

func validateRemoteAddr(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsInterfaceLocalMulticast() {
		return false
	}
	if ip.IsLinkLocalMulticast() {
		return false
	}
	if ip.IsLinkLocalUnicast() {
		return false
	}
	if ip.IsLoopback() {
		return false
	}
	if ip.IsMulticast() {
		return false
	}
	if ip.IsUnspecified() {
		return false
	}
	if isBroadcasty(ip) {
		return false
	}

	return true
}

func isBroadcasty(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		if ip4.Equal(net.IPv4bcast) {
			return true
		}
		if ip4.Equal(net.IPv4allsys) {
			return true
		}
		if ip4.Equal(net.IPv4allrouter) {
			return true
		}
		if ip4.Equal(net.IPv4zero) {
			return true
		}
		return false
	}
	if ip.Equal(net.IPv6zero) {
		return true
	}
	if ip.Equal(net.IPv6unspecified) {
		return true
	}
	if ip.Equal(net.IPv6loopback) {
		return true
	}
	if ip.Equal(net.IPv6interfacelocalallnodes) {
		return true
	}
	if ip.Equal(net.IPv6linklocalallnodes) {
		return true
	}
	if ip.Equal(net.IPv6linklocalallrouters) {
		return true
	}
	return false
}
