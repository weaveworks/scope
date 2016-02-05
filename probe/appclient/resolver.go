package appclient

import (
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/scope/common/xfer"
)

const (
	dnsPollInterval = 10 * time.Second
)

var (
	tick     = time.Tick
	lookupIP = net.LookupIP
)

type setter func(string, []string)

// Resolver is a thing that can be stopped...
type Resolver interface {
	Stop()
}

type staticResolver struct {
	setters []setter
	targets []target
	quit    chan struct{}
}

type target struct{ host, port string }

func (t target) String() string { return net.JoinHostPort(t.host, t.port) }

// NewStaticResolver periodically resolves the targets, and calls the set
// function with all the resolved IPs. It explictiy supports targets which
// resolve to multiple IPs.
func NewStaticResolver(targets []string, setters ...setter) Resolver {
	r := staticResolver{
		targets: prepare(targets),
		setters: setters,
		quit:    make(chan struct{}),
	}
	go r.loop()
	return r
}

func (r staticResolver) loop() {
	r.resolve()
	t := tick(dnsPollInterval)
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
				log.Errorf("invalid address %s: %v", s, err)
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
	for t, endpoints := range resolveMany(r.targets) {
		for _, setter := range r.setters {
			setter(t.String(), endpoints)
		}
	}
}

func resolveMany(targets []target) map[target][]string {
	result := map[target][]string{}
	for _, t := range targets {
		result[t] = resolveOne(t)
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
