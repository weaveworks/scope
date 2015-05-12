package main

import (
	"log"
	"net"
	"time"
)

// Resolver periodically tries to resolve the IP addresses for a given
// set of hostnames.
type Resolver struct {
	quit  chan struct{}
	add   func(string)
	peers []peer
}

type peer struct {
	hostname string
	port     string
}

// NewResolver starts a new resolver that periodically
// tries to resolve peers and the calls add() with all the
// resolved IPs.  It explictiy supports hostnames which
// resolve to multiple IPs; it will repeatedly call
// add with the same IP, expecting the target to dedupe.
func NewResolver(peers []string, add func(string)) Resolver {
	resolver := Resolver{
		quit:  make(chan struct{}),
		add:   add,
		peers: prepareNames(peers),
	}

	go resolver.loop()
	return resolver
}

func prepareNames(peers []string) []peer {
	var results []peer
	for _, p := range peers {
		hostname, port, err := net.SplitHostPort(p)
		if err != nil {
			log.Printf("invalid address %s: %v", p, err)
			continue
		}
		results = append(results, peer{hostname, port})
	}
	return results
}

func (r Resolver) loop() {
	r.resolveHosts()
	for {
		tick := time.Tick(1 * time.Minute)
		select {
		case <-tick:
			r.resolveHosts()
		case <-r.quit:
			return
		}
	}
}

func (r Resolver) resolveHosts() {
	for _, peer := range r.peers {
		addrs, err := net.LookupIP(peer.hostname)
		if err != nil {
			log.Printf("lookup %s: %v", peer.hostname, err)
			continue
		}

		for _, addr := range addrs {
			// For now, ignore IPv6
			if addr.To4() == nil {
				continue
			}

			r.add(net.JoinHostPort(addr.String(), peer.port))
		}
	}
}

// Stop this resolver.
func (r Resolver) Stop() {
	r.quit <- struct{}{}
}
