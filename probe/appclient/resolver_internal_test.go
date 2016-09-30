package appclient

import (
	"fmt"
	"net"
	"net/url"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/test"
)

func TestResolverCases(t *testing.T) {
	c := make(chan time.Time)
	ticker := func(_ time.Duration) <-chan time.Time { return c }

	ips := map[string][]net.IP{
		"foo": {net.IPv4(192, 168, 0, 1)},
		"bar": {net.IPv4(192, 168, 0, 2), net.IPv4(192, 168, 0, 3)},
	}
	lookupIP := func(host string) ([]net.IP, error) {
		addrs, ok := ips[host]
		if !ok {
			return nil, fmt.Errorf("Not found")
		}
		return addrs, nil
	}

	testResolver := func(target string, expected []url.URL) {
		mtx := sync.Mutex{}
		found := map[url.URL]struct{}{}
		set := func(target string, urls []url.URL) {
			mtx.Lock()
			defer mtx.Unlock()
			for _, url := range urls {
				found[url] = struct{}{}
			}
		}

		targets, err := ParseTargets([]string{target})
		if err != nil {
			t.Fatal(err)
		}

		r, err := NewResolver(ResolverConfig{
			Targets: targets,
			Lookup:  lookupIP,
			Set:     set,
			Ticker:  ticker,
		})
		if err != nil {
			t.Fatal(err)
		}
		defer r.Stop()

		c <- time.Now()
		test.Poll(t, 200*time.Millisecond, expected, func() interface{} {
			mtx.Lock()
			defer mtx.Unlock()
			have := []url.URL{}
			for url := range found {
				have = append(have, url)
			}
			return have
		})
	}

	for _, tc := range []struct {
		in       string
		expected []url.URL
	}{
		{"foo", []url.URL{{Scheme: "http", Host: "192.168.0.1:4040"}}},
		{"foo:80", []url.URL{{Scheme: "http", Host: "192.168.0.1:80"}}},
		{"foo:443", []url.URL{{Scheme: "https", Host: "192.168.0.1:443"}}},
		{"foo:1234", []url.URL{{Scheme: "http", Host: "192.168.0.1:1234"}}},
		{"http://foo", []url.URL{{Scheme: "http", Host: "192.168.0.1:80"}}},
		{"http://foo:80", []url.URL{{Scheme: "http", Host: "192.168.0.1:80"}}},
		{"http://foo:443", []url.URL{{Scheme: "http", Host: "192.168.0.1:443"}}},
		{"http://foo:1234", []url.URL{{Scheme: "http", Host: "192.168.0.1:1234"}}},
		{"https://foo", []url.URL{{Scheme: "https", Host: "192.168.0.1:443"}}},
		{"https://foo:80", []url.URL{{Scheme: "https", Host: "192.168.0.1:80"}}},
		{"https://foo:443", []url.URL{{Scheme: "https", Host: "192.168.0.1:443"}}},
		{"https://foo:1234", []url.URL{{Scheme: "https", Host: "192.168.0.1:1234"}}},
		{"user:pass@foo", []url.URL{{Scheme: "http", Host: "192.168.0.1:4040", User: url.UserPassword("user", "pass")}}},
		{"bar", []url.URL{{Scheme: "http", Host: "192.168.0.2:4040"}, {Scheme: "http", Host: "192.168.0.3:4040"}}},
	} {
		testResolver(tc.in, tc.expected)
	}
}

func TestResolver(t *testing.T) {
	c := make(chan time.Time)
	ticker := func(_ time.Duration) <-chan time.Time { return c }

	ipsLock := sync.Mutex{}
	ips := map[string][]net.IP{}
	lookupIP := func(host string) ([]net.IP, error) {
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
	sets := make(chan url.URL)
	set := func(target string, endpoints []url.URL) {
		for _, endpoint := range endpoints {
			sets <- endpoint
		}
	}

	targets, err := ParseTargets([]string{"symbolic.name" + port, "namewithnoport", ip1 + port, ip2})
	if err != nil {
		t.Fatal(err)
	}

	r, err := NewResolver(ResolverConfig{
		Targets: targets,
		Lookup:  lookupIP,
		Set:     set,
		Ticker:  ticker,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertAdd := func(want ...url.URL) {
		remaining := map[url.URL]struct{}{}
		for _, s := range want {
			remaining[s] = struct{}{}
		}
		_, _, line, _ := runtime.Caller(1)
		for len(remaining) > 0 {
			select {
			case s := <-sets:
				if _, ok := remaining[s]; ok {
					t.Logf("line %d: got %v OK", line, s)
					delete(remaining, s)
				} else {
					t.Errorf("line %d: got unexpected %v", line, s)
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatalf("line %d: didn't get the adds in time", line)
			}
		}
	}

	// Initial resolve should just give us IPs
	assertAdd(
		url.URL{Scheme: "http", Host: ip1 + port},
		url.URL{Scheme: "http", Host: fmt.Sprintf("%s:%d", ip2, xfer.AppPort)},
	)

	// Trigger another resolve with a tick; again,
	// just want ips.
	c <- time.Now()
	assertAdd(
		url.URL{Scheme: "http", Host: ip1 + port},
		url.URL{Scheme: "http", Host: fmt.Sprintf("%s:%d", ip2, xfer.AppPort)},
	)

	ip3 := "1.2.3.4"
	updateIPs("symbolic.name", makeIPs(ip3))
	c <- time.Now() // trigger a resolve
	assertAdd(
		url.URL{Scheme: "http", Host: ip3 + port},
		url.URL{Scheme: "http", Host: ip1 + port}, url.URL{Scheme: "http", Host: fmt.Sprintf("%s:%d", ip2, xfer.AppPort)},
	)

	ip4 := "10.10.10.10"
	updateIPs("symbolic.name", makeIPs(ip3, ip4))
	c <- time.Now() // trigger another resolve, this time with 2 adds
	assertAdd(
		url.URL{Scheme: "http", Host: ip3 + port},
		url.URL{Scheme: "http", Host: ip4 + port},
		url.URL{Scheme: "http", Host: ip1 + port},
		url.URL{Scheme: "http", Host: fmt.Sprintf("%s:%d", ip2, xfer.AppPort)},
	)

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
