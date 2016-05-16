package appclient

import (
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"

	"github.com/weaveworks/scope/common/xfer"
)

const (
	dnsPollInterval = 10 * time.Second
)

var (
	tick = fastStartTicker
)

// fastStartTicker is a ticker that 'ramps up' from 1 sec to duration.
func fastStartTicker(duration time.Duration) <-chan time.Time {
	c := make(chan time.Time, 1)
	go func() {
		d := 1 * time.Second
		for {
			time.Sleep(d)
			d = d * 2
			if d > duration {
				d = duration
			}

			select {
			case c <- time.Now():
			default:
			}
		}
	}()
	return c
}

type setter func(string, []string)

// Resolver is a thing that can be stopped...
type Resolver interface {
	Stop()
}

type staticResolver struct {
	setters []setter
	targets []target
	quit    chan struct{}
	lookup  LookupIP
}

// LookupIP type is used for looking up IPs.
type LookupIP func(host string) (ips []net.IP, err error)

type target struct{ host, port string }

func (t target) String() string { return net.JoinHostPort(t.host, t.port) }

// NewResolver periodically resolves the targets, and calls the set
// function with all the resolved IPs. It explictiy supports targets which
// resolve to multiple IPs.  It uses the supplied DNS server name.
func NewResolver(targets []string, lookup LookupIP, setters ...setter) Resolver {
	r := staticResolver{
		targets: prepare(targets),
		setters: setters,
		quit:    make(chan struct{}),
		lookup:  lookup,
	}
	go r.loop()
	return r
}

// LookupUsing produces a LookupIP function for the given DNS server.
func LookupUsing(dnsServer string) func(host string) (ips []net.IP, err error) {
	client := dns.Client{
		Net: "tcp",
	}
	return func(host string) (ips []net.IP, err error) {
		m := &dns.Msg{}
		m.SetQuestion(dns.Fqdn(host), dns.TypeA)
		in, _, err := client.Exchange(m, dnsServer)
		if err != nil {
			return nil, err
		}
		result := []net.IP{}
		for _, answer := range in.Answer {
			if a, ok := answer.(*dns.A); ok {
				result = append(result, a.A)
			}
		}
		return result, nil
	}
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
	for t, endpoints := range r.resolveMany(r.targets) {
		for _, setter := range r.setters {
			setter(t.String(), endpoints)
		}
	}
}

func (r staticResolver) resolveMany(targets []target) map[target][]string {
	result := map[target][]string{}
	for _, t := range targets {
		result[t] = r.resolveOne(t)
	}
	return result
}

func (r staticResolver) resolveOne(t target) []string {
	var addrs []net.IP
	if addr := net.ParseIP(t.host); addr != nil {
		addrs = []net.IP{addr}
	} else {
		var err error
		addrs, err = r.lookup(t.host)
		if err != nil {
			log.Debugf("Error resolving %s: %v", t.host, err)
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
