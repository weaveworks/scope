package main

import (
	"fmt"
	"net"
	"runtime"
	"sync"
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
	ipsLock := sync.Mutex{}
	ips := map[string][]net.IP{}
	lookupIP = func(host string) ([]net.IP, error) {
		ipsLock.Lock()
		defer ipsLock.Unlock()
		addrs, ok := ips[host]
		if !ok {
			return nil, fmt.Errorf("Not found")
		}
		return addrs, nil
	}
	updateIPs := func(key string, values []net.IP) {
		ipsLock.Lock()
		defer ipsLock.Unlock()
		ips = map[string][]net.IP{key: values}
	}

	port := ":80"
	ip1 := "192.168.0.1"
	ip2 := "192.168.0.10"
	adds := make(chan string)
	add := func(s string) { adds <- s }

	r := newStaticResolver([]string{"symbolic.name" + port, "namewithnoport", ip1 + port, ip2}, add)

	assertAdd := func(want string) {
		_, _, line, _ := runtime.Caller(1)
		select {
		case have := <-adds:
			if want != have {
				t.Errorf("line %d: want %q, have %q", line, want, have)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("line %d: didn't get add in time", line)
		}
	}

	// Initial resolve should just give us IPs
	assertAdd(ip1 + port)
	assertAdd(fmt.Sprintf("%s:%d", ip2, xfer.AppPort))

	// Trigger another resolve with a tick; again,
	// just want ips.
	c <- time.Now()
	assertAdd(ip1 + port)
	assertAdd(fmt.Sprintf("%s:%d", ip2, xfer.AppPort))

	ip3 := "1.2.3.4"
	updateIPs("symbolic.name", makeIPs(ip3))
	c <- time.Now()       // trigger a resolve
	assertAdd(ip3 + port) // we want 1 add
	assertAdd(ip1 + port)
	assertAdd(fmt.Sprintf("%s:%d", ip2, xfer.AppPort))

	ip4 := "10.10.10.10"
	updateIPs("symbolic.name", makeIPs(ip3, ip4))
	c <- time.Now()       // trigger another resolve, this time with 2 adds
	assertAdd(ip3 + port) // first add
	assertAdd(ip4 + port) // second add
	assertAdd(ip1 + port)
	assertAdd(fmt.Sprintf("%s:%d", ip2, xfer.AppPort))

	done := make(chan struct{})
	go func() { r.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
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
