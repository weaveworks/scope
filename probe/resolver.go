package main

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/weaveworks/scope/xfer"
)

var (
	tick     = time.Tick
	lookupIP = net.LookupIP
)

const maxConcurrentLookup = 10

type staticResolver struct {
	set     func(string, []string)
	targets []target
	sema    semaphore
	quit    chan struct{}
}

type target struct{ host, port string }

func (t target) String() string { return net.JoinHostPort(t.host, t.port) }

// newStaticResolver periodically resolves the targets, and calls the set
// function with all the resolved IPs. It explictiy supports targets which
// resolve to multiple IPs.
func newStaticResolver(targets []string, set func(target string, endpoints []string)) staticResolver {
	r := staticResolver{
		targets: prepare(targets),
		set:     set,
		sema:    newSemaphore(maxConcurrentLookup),
		quit:    make(chan struct{}),
	}
	go r.loop()
	return r
}

func (r staticResolver) loop() {
	r.resolve()
	t := tick(time.Minute)
	for {
		select {
		case <-t:
			r.resolve()
		case <-r.quit:
			return
		}
	}
}

func (r staticResolver) Stop() {
	close(r.quit)
}

func prepare(strs []string) []target {
	var targets []target
	for _, s := range strs {
		var host, port string
		if strings.Contains(s, ":") {
			var err error
			host, port, err = net.SplitHostPort(s)
			if err != nil {
				log.Printf("invalid address %s: %v", s, err)
				continue
			}
		} else {
			host, port = s, strconv.Itoa(xfer.AppPort)
		}
		targets = append(targets, target{host, port})
	}
	return targets
}

func (r staticResolver) resolve() {
	for t, endpoints := range resolveMany(r.sema, r.targets) {
		r.set(t.String(), endpoints)
	}
}

func resolveMany(s semaphore, targets []target) map[target][]string {
	result := map[target][]string{}
	for _, t := range targets {
		c := make(chan []string)
		go func(t target) { s.p(); defer s.v(); c <- resolveOne(t) }(t)
		result[t] = <-c
	}
	return result
}

func resolveOne(t target) []string {
	var addrs []net.IP
	if addr := net.ParseIP(t.host); addr != nil {
		addrs = []net.IP{addr}
	} else {
		var err error
		addrs, err = lookupIP(t.host)
		if err != nil {
			return []string{}
		}
	}
	endpoints := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		// For now, ignore IPv6
		if addr.To4() == nil {
			continue
		}
		endpoints = append(endpoints, net.JoinHostPort(addr.String(), t.port))
	}
	return endpoints
}

type semaphore chan struct{}

func newSemaphore(n int) semaphore {
	c := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		c <- struct{}{}
	}
	return semaphore(c)
}
func (s semaphore) p() { <-s }
func (s semaphore) v() { s <- struct{}{} }
