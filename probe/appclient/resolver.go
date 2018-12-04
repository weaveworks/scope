package appclient

import (
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/scope/common/xfer"
)

const (
	dnsPollInterval = 10 * time.Second
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

// Resolver is a thing that can be stopped...
type Resolver interface {
	Stop()
}

type staticResolver struct {
	ResolverConfig

	failedResolutions map[string]struct{}
	quit              chan struct{}
}

// LookupIP type is used for looking up IPs.
type LookupIP func(host string) (ips []net.IP, err error)

// Target is a parsed representation of the app location.
type Target struct {
	original string   // the original url string
	url      *url.URL // the parsed url
	hostname string   // the hostname (without port) from the url
	port     int      // the port, or a sensible default
}

func (t Target) String() string {
	return net.JoinHostPort(t.hostname, strconv.Itoa(t.port))
}

// ResolverConfig is the config for a resolver.
type ResolverConfig struct {
	Targets []Target
	Set     func(string, []url.URL)

	// Optional
	Lookup LookupIP
	Ticker func(time.Duration) <-chan time.Time
}

// NewResolver periodically resolves the targets, and calls the set
// function with all the resolved IPs. It explictiy supports targets which
// resolve to multiple IPs.  It uses the supplied DNS server name.
func NewResolver(config ResolverConfig) (Resolver, error) {
	if config.Lookup == nil {
		config.Lookup = net.LookupIP
	}
	if config.Ticker == nil {
		config.Ticker = fastStartTicker
	}
	r := staticResolver{
		ResolverConfig:    config,
		failedResolutions: map[string]struct{}{},
		quit:              make(chan struct{}),
	}
	go r.loop()
	return r, nil
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
	t := r.Ticker(dnsPollInterval)
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

// ParseTargets deals with missing information in the targets string, defaulting
// the scheme, port etc.
func ParseTargets(urls []string) ([]Target, error) {
	var targets []Target
	for _, u := range urls {
		// naked hostnames (such as "localhost") are interpreted as relative URLs
		// so we add a scheme if u doesn't have one.
		prefixAdded := false
		if !strings.Contains(u, "://") {
			prefixAdded = true
			if strings.HasSuffix(u, ":443") {
				u = "https://" + u
			} else {
				u = "http://" + u
			}
		}
		parsed, err := url.Parse(u)
		if err != nil {
			return nil, err
		}

		var hostname string
		var port int
		if strings.Contains(parsed.Host, ":") {
			var portStr string
			hostname, portStr, err = net.SplitHostPort(parsed.Host)
			if err != nil {
				return nil, err
			}
			port, err = strconv.Atoi(portStr)
			if err != nil {
				return nil, err
			}
		} else {
			if prefixAdded {
				port = xfer.AppPort
			} else if strings.HasPrefix(u, "https://") {
				port = 443
			} else {
				port = 80
			}
			hostname = parsed.Host
		}
		targets = append(targets, Target{
			original: u,
			url:      parsed,
			hostname: hostname,
			port:     port,
		})
	}
	return targets, nil
}

func (r staticResolver) resolve() {
	for _, t := range r.Targets {
		ips := r.resolveOne(t)
		urls := makeURLs(t, ips)
		r.Set(t.hostname, urls)
	}
}

func makeURLs(t Target, ips []string) []url.URL {
	result := []url.URL{}
	for _, ip := range ips {
		u := *t.url
		u.Host = net.JoinHostPort(ip, strconv.Itoa(t.port))
		result = append(result, u)
	}
	return result
}

func (r staticResolver) resolveOne(t Target) []string {
	var addrs []net.IP
	if addr := net.ParseIP(t.hostname); addr != nil {
		addrs = []net.IP{addr}
	} else {
		var err error
		addrs, err = r.Lookup(t.hostname)
		if err != nil {
			if _, ok := r.failedResolutions[t.hostname]; !ok {
				log.Warnf("Cannot resolve '%s': %v", t.hostname, err)
				// Only log the error once
				r.failedResolutions[t.hostname] = struct{}{}
			}
			return []string{}
		}
		// Allow logging errors in future resolutions
		delete(r.failedResolutions, t.hostname)
	}
	endpoints := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		// For now, ignore IPv6
		if addr.To4() == nil {
			continue
		}
		endpoints = append(endpoints, addr.String())
	}
	return endpoints
}
