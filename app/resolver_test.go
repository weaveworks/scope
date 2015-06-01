package main

import (
	"fmt"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/weaveworks/scope/xfer"
)

func TestResolver(t *testing.T) {
	oldTick := tick
	defer func() { tick = oldTick }()
	c := make(chan time.Time)
	tick = func(_ time.Duration) <-chan time.Time { return c }

	oldLookupIP := lookupIP
	defer func() { lookupIP = oldLookupIP }()
	ips := map[string][]net.IP{}
	lookupIP = func(host string) ([]net.IP, error) {
		addrs, ok := ips[host]
		if !ok {
			return nil, fmt.Errorf("Not found")
		}
		return addrs, nil
	}

	port := ":80"
	ip1 := "192.168.0.1"
	ip2 := "192.168.0.10"
	adds := make(chan string)
	add := func(s string) { adds <- s }
	r := NewResolver([]string{"symbolic.name" + port, "namewithnoport", ip1 + port, ip2}, add)

	assertAdd := func(want string) {
		select {
		case have := <-adds:
			if want != have {
				_, _, line, _ := runtime.Caller(1)
				t.Errorf("line %d: want %q, have %q", line, want, have)
			}
		case <-time.After(time.Millisecond):
			t.Fatal("didn't get add in time")
		}
	}

	// Initial resolve should just give us IPs
	assertAdd(ip1 + port)
	assertAdd(fmt.Sprintf("%s:%d", ip2, xfer.ProbePort))

	// Trigger another resolve with a tick; again,
	// just want ips.
	c <- time.Now()
	assertAdd(ip1 + port)
	assertAdd(fmt.Sprintf("%s:%d", ip2, xfer.ProbePort))

	ip3 := "1.2.3.4"
	ips = map[string][]net.IP{"symbolic.name": makeIPs(ip3)}
	c <- time.Now()       // trigger a resolve
	assertAdd(ip3 + port) // we want 1 add
	assertAdd(ip1 + port)
	assertAdd(fmt.Sprintf("%s:%d", ip2, xfer.ProbePort))

	ip4 := "10.10.10.10"
	ips = map[string][]net.IP{"symbolic.name": makeIPs(ip3, ip4)}
	c <- time.Now()       // trigger another resolve, this time with 2 adds
	assertAdd(ip3 + port) // first add
	assertAdd(ip4 + port) // second add
	assertAdd(ip1 + port)
	assertAdd(fmt.Sprintf("%s:%d", ip2, xfer.ProbePort))

	done := make(chan struct{})
	go func() { r.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Millisecond):
		t.Errorf("didn't Stop in time")
	}
}

func makeIPs(addrs ...string) []net.IP {
	var ips []net.IP
	for _, addr := range addrs {
		ips = append(ips, net.ParseIP(addr))
	}
	return ips
}
