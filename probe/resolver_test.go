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
	sets := make(chan string)
	set := func(target string, endpoints []string) {
		for _, endpoint := range endpoints {
			sets <- endpoint
		}
	}

	r := newStaticResolver([]string{"symbolic.name" + port, "namewithnoport", ip1 + port, ip2}, set)

	assertAdd := func(want ...string) {
		remaining := map[string]struct{}{}
		for _, s := range want {
			remaining[s] = struct{}{}
		}
		_, _, line, _ := runtime.Caller(1)
		for len(remaining) > 0 {
			select {
			case s := <-sets:
				if _, ok := remaining[s]; ok {
					t.Logf("line %d: got %q OK", line, s)
					delete(remaining, s)
				} else {
					t.Errorf("line %d: got unexpected %q", line, s)
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatalf("line %d: didn't get the adds in time", line)
			}
		}
	}

	// Initial resolve should just give us IPs
	assertAdd(ip1+port, fmt.Sprintf("%s:%d", ip2, xfer.AppPort))

	// Trigger another resolve with a tick; again,
	// just want ips.
	c <- time.Now()
	assertAdd(ip1+port, fmt.Sprintf("%s:%d", ip2, xfer.AppPort))

	ip3 := "1.2.3.4"
	updateIPs("symbolic.name", makeIPs(ip3))
	c <- time.Now() // trigger a resolve
	assertAdd(ip3+port, ip1+port, fmt.Sprintf("%s:%d", ip2, xfer.AppPort))

	ip4 := "10.10.10.10"
	updateIPs("symbolic.name", makeIPs(ip3, ip4))
	c <- time.Now() // trigger another resolve, this time with 2 adds
	assertAdd(ip3+port, ip4+port, ip1+port, fmt.Sprintf("%s:%d", ip2, xfer.AppPort))

	done := make(chan struct{})
	go func() { r.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("didn't Stop in time")
	}
}

func TestSemaphore(t *testing.T) {
	n := 3
	s := newSemaphore(n)

	// First n should be fine
	for i := 0; i < n; i++ {
		ok := make(chan struct{})
		go func() { s.p(); close(ok) }()
		select {
		case <-ok:
		case <-time.After(10 * time.Millisecond):
			t.Errorf("p (%d) failed", i+1)
		}
	}

	// This should block
	ok := make(chan struct{})
	go func() { s.p(); close(ok) }()
	select {
	case <-ok:
		t.Errorf("%dth p OK, but should block", n+1)
	case <-time.After(10 * time.Millisecond):
		//t.Logf("%dth p blocks, as expected", n+1)
	}

	s.v()

	select {
	case <-ok:
	case <-time.After(10 * time.Millisecond):
		t.Errorf("%dth p didn't resolve in time", n+1)
	}
}

func makeIPs(addrs ...string) []net.IP {
	var ips []net.IP
	for _, addr := range addrs {
		ips = append(ips, net.ParseIP(addr))
	}
	return ips
}
