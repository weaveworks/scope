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
	r := Resolver{
		quit:  make(chan struct{}),
		add:   add,
		peers: prepareNames(peers),
	}
	go r.loop()
	return r
}

func prepareNames(strs []string) []peer {
	var results []peer
	for _, s := range strs {
		var (
			hostname string
			port     string
		)

		if strings.Contains(s, ":") {
			var err error
			hostname, port, err = net.SplitHostPort(s)
			if err != nil {
				log.Printf("invalid address %s: %v", s, err)
				continue
			}
		} else {
			hostname, port = s, strconv.Itoa(xfer.ProbePort)
		}

		results = append(results, peer{hostname, port})
	}
	return results
}

func (r Resolver) loop() {
	r.resolveHosts()
	t := tick(time.Minute)
	for {
		select {
		case <-t:
			r.resolveHosts()
		case <-r.quit:
			return
		}
	}
}

func (r Resolver) resolveHosts() {
	for _, peer := range r.peers {
		var addrs []net.IP
		if addr := net.ParseIP(peer.hostname); addr != nil {
			addrs = []net.IP{addr}
		} else {
			var err error
			addrs, err = lookupIP(peer.hostname)
			if err != nil {
				log.Printf("lookup %s: %v", peer.hostname, err)
				continue
			}
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

// Stop this Resolver.
func (r Resolver) Stop() {
	close(r.quit)
}
