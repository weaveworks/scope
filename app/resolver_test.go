package main

import (
	"net"
	"runtime"
	"testing"
	"time"
)

func TestResolver(t *testing.T) {
	oldTick := tick
	defer func() { tick = oldTick }()
	c := make(chan time.Time)
	tick = func(_ time.Duration) <-chan time.Time { return c }

	oldLookupIP := lookupIP
	defer func() { lookupIP = oldLookupIP }()
	ips := []net.IP{}
	lookupIP = func(host string) ([]net.IP, error) { return ips, nil }

	port := ":80"
	adds := make(chan string)
	add := func(s string) { adds <- s }
	r := NewResolver([]string{"symbolic.name" + port}, add)

	c <- time.Now() // trigger initial resolve, with no endpoints
	select {
	case <-time.After(time.Millisecond):
	case s := <-adds:
		t.Errorf("got unexpected add: %q", s)
	}

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

	ip1 := "1.2.3.4"
	ips = makeIPs(ip1)
	c <- time.Now()       // trigger a resolve
	assertAdd(ip1 + port) // we want 1 add

	ip2 := "10.10.10.10"
	ips = makeIPs(ip1, ip2)
	c <- time.Now()       // trigger another resolve, this time with 2 adds
	assertAdd(ip1 + port) // first add
	assertAdd(ip2 + port) // second add

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
