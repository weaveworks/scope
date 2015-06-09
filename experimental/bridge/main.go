package main

import (
	"flag"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

const (
	trafficTimeout = 2 * time.Minute
)

func main() {
	var (
		listen  = flag.String("listen", ":"+strconv.Itoa(xfer.ProbePort), "listen address")
		probes  = flag.String("probes", "", "list of all initial probes, comma separated")
		batch   = flag.Duration("batch", 1*time.Second, "batch interval")
		version = flag.Bool("version", false, "print version number and exit")
	)
	flag.Parse()

	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *version {
		//fmt.Printf("%s\n", probe.Version)
		return
	}

	if *probes == "" {
		log.Fatal("no probes given via -probes")
	}

	log.Printf("starting")

	fixedAddresses := strings.Split(*probes, ",")

	// Collector deals with the probes, and generates a single merged report
	// every second.
	c := xfer.NewCollector(*batch)
	for _, addr := range fixedAddresses {
		c.Add(addr)
	}
	defer c.Stop()

	publisher, err := xfer.NewTCPPublisher(*listen)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()
	log.Printf("listening on %s\n", *listen)

	var fixedIPs []string
	for _, a := range fixedAddresses {
		if addr, _, err := net.SplitHostPort(a); err == nil {
			fixedIPs = append(fixedIPs, addr)
		}
	}
	go discover(c, publisher, fixedIPs)

	<-interrupt()

	log.Printf("shutting down")
}

func interrupt() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}

type collector interface {
	Reports() <-chan report.Report
	Remove(string)
	Add(string)
}

type publisher xfer.Publisher

// makeAvoid makes a map with IPs we don't want to consider in discover(). It
// is the set of IPs which the bridge is configured to connect to, and the all
// the IPs from the local interfaces.
func makeAvoid(fixed []string) map[string]struct{} {
	avoid := map[string]struct{}{}

	// Don't discover fixed probes. This way we'll never remove them.
	for _, a := range fixed {
		avoid[a] = struct{}{}
	}

	// Don't go Ouroboros.
	if localNets, err := net.InterfaceAddrs(); err == nil {
		for _, n := range localNets {
			if net, ok := n.(*net.IPNet); ok {
				avoid[net.IP.String()] = struct{}{}
			}
		}
	}

	return avoid
}

// discover reads reports from a collector and republishes them on the
// publisher, while scanning the reports for IPs to connect to. Only addresses
// in the network topology of the report are considered. IPs listed in fixed
// are always skipped.
func discover(c collector, p publisher, fixed []string) {
	lastSeen := map[string]time.Time{}

	avoid := makeAvoid(fixed)

	for r := range c.Reports() {
		// log.Printf("got a report")
		p.Publish(r)

		var (
			now       = time.Now()
			localNets = r.LocalNets()
		)

		for _, adjacent := range r.Address.Adjacency {
			for _, a := range adjacent {
				ip := report.AddressIDAddresser(a) // address id -> IP
				if ip == nil {
					continue
				}

				addr := ip.String()
				if _, ok := avoid[addr]; ok {
					continue
				}
				// log.Printf("potential address: %v (via %s)", addr, src)
				if _, ok := lastSeen[addr]; !ok {
					if interestingAddress(localNets, addr) {
						log.Printf("discovery %v: potential probe address", addr)
						c.Add(addressToDial(addr))
					} else {
						log.Printf("discovery %v: non-probe address", addr)
					}
				}

				// We always add addr to lastSeen[], even if it's a non-local
				// address. This way they are part of the normal timeout logic,
				// and we won't analyze the address again until it re-appears
				// after a timeout.
				lastSeen[addr] = now
			}
		}

		for addr, last := range lastSeen {
			if now.Sub(last) > trafficTimeout {
				// Timeout can be for a non-local address, which we didn't
				// connect to. In that case the RemoveAddress() call won't do
				// anything.
				log.Printf("discovery %v: traffic timeout", addr)
				delete(lastSeen, addr)
				c.Remove(addressToDial(addr))
			}
		}
	}
}

// interestingAddress tells whether the address is a local and normal address,
// which we want to try to connect to.
func interestingAddress(localNets []*net.IPNet, addr string) bool {
	if addr == "" {
		return false
	}

	// The address is expected to be an IPv{4,6} address.
	ip := net.ParseIP(addr)
	if ip == nil {
		return false
	}

	// Filter out localhost, broadcast, and other non-connectable addresses.
	if !validateRemoteAddr(ip) {
		return false
	}

	// Only connect to addresses we know are localnet.
	for _, n := range localNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// addressToDial formats an IP address so we can pass it on to Dial().
func addressToDial(address string) string {
	// return fmt.Sprintf("[%s]:%d", addr, probePort)
	return net.JoinHostPort(address, strconv.Itoa(xfer.ProbePort))
}
