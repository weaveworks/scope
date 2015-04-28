package main

import (
	"net"
	"testing"
)

func TestValidRemoteAddr(t *testing.T) {
	for input, expected := range map[string]bool{
		"localhost": false,

		"127.0.0.1":       false, // should be same as loopback
		"127.1.2.3":       false, // 127.0.0.0/8 is all loopback
		"0.0.0.0":         false,
		"224.0.0.1":       false, // all systems
		"224.0.0.2":       false, // all routers
		"224.0.0.22":      false,
		"255.255.255.255": false, // broadcast
		"1.2.3.4":         true,

		"::":                        false, // unspecified
		"0:0:0:0:0:0:0:0":           false, // unspecified (alternate form)
		"b8:27:eb:a4:bf:6e":         false,
		"fe80::1240:f3ff:fe80:5474": false, // loopback
		"fe80::1":                   false, // loopback
		"::1":                       false, // loopback
		"0:0:0:0:0:0:0:1":           false, // loopback (alternate form)
		"2001:db8::1:0:0:1":         true,  // valid

		// http://www.iana.org/assignments/ipv6-multicast-addresses/ipv6-multicast-addresses.xhtml
		"FF01:0:0:0:0:0:0:1": false, // Node-local all-nodes
		// "FF01:0:0:0:0:0:0:2": false, // Node-local all-routers, isn't spec'd
		// in package net/ip
		"FF02:0:0:0:0:0:0:1": false, // Link-local all-nodes
		"FF02:0:0:0:0:0:0:2": false, // Link-local all-routers
	} {
		if got := validateRemoteAddr(net.ParseIP(input)); expected != got {
			t.Errorf("%s: expected %v, got %v", input, expected, got)
		}
	}
}
