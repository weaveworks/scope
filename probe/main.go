package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/probe/tag"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

var version = "dev" // set at build time

func main() {
	var (
		httpListen         = flag.String("http.listen", "", "listen address for HTTP profiling and instrumentation server")
		publishInterval    = flag.Duration("publish.interval", 3*time.Second, "publish (output) interval")
		spyInterval        = flag.Duration("spy.interval", time.Second, "spy (scan) interval")
		listen             = flag.String("listen", ":"+strconv.Itoa(xfer.ProbePort), "listen address")
		prometheusEndpoint = flag.String("prometheus.endpoint", "/metrics", "Prometheus metrics exposition endpoint (requires -http.listen)")
		spyProcs           = flag.Bool("processes", true, "report processes (needs root)")
		dockerTagger       = flag.Bool("docker", true, "collect Docker-related attributes for processes")
		dockerInterval     = flag.Duration("docker.interval", 10*time.Second, "how often to update Docker attributes")
		procRoot           = flag.String("proc.root", "/proc", "location of the proc filesystem")
	)
	flag.Parse()

	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("probe version %s", version)

	procspy.SetProcRoot(*procRoot)

	if *httpListen != "" {
		log.Printf("profiling data being exported to %s", *httpListen)
		log.Printf("go tool pprof http://%s/debug/pprof/{profile,heap,block}", *httpListen)
		if *prometheusEndpoint != "" {
			log.Printf("exposing Prometheus endpoint at %s%s", *httpListen, *prometheusEndpoint)
			http.Handle(*prometheusEndpoint, makePrometheusHandler())
		}
		go func(err error) { log.Print(err) }(http.ListenAndServe(*httpListen, nil))
	}

	if *spyProcs && os.Getegid() != 0 {
		log.Printf("warning: process reporting enabled, but not running as root: will miss some things")
	}

	publisher, err := xfer.NewTCPPublisher(*listen)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	taggers := []tag.Tagger{tag.NewTopologyTagger()}
	if *dockerTagger {
		t := tag.NewDockerTagger(*procRoot, *dockerInterval)
		defer t.Stop()
		taggers = append(taggers, t)
	}

	log.Printf("listening on %s", *listen)

	quit := make(chan struct{})
	defer close(quit)

	go func() {
		var (
			hostname = hostname()
			hostID   = hostname // TODO: we should sanitize the hostname
			pubTick  = time.Tick(*publishInterval)
			spyTick  = time.Tick(*spyInterval)
			r        = report.MakeReport()
		)

		for {
			select {
			case <-pubTick:
				publishTicks.WithLabelValues().Add(1)
				r.Host = hostTopology(hostID, hostname)
				r = tag.Apply(r, taggers)
				publisher.Publish(r)
				r = report.MakeReport()

			case <-spyTick:
				r.Merge(spy(hostname, hostname, *spyProcs))
				// log.Printf("merged report:\n%#v\n", r)

			case <-quit:
				return
			}
		}
	}()

	log.Printf("%s", <-interrupt())
}

func hostname() string {
	if hostname := os.Getenv("SCOPE_HOSTNAME"); hostname != "" {
		return hostname
	}
	hostname, err := os.Hostname()
	if err != nil {
		return "(unknown)"
	}
	return hostname
}

// hostTopology produces a host topology for this host. No need to do this
// more than once per published report.
func hostTopology(hostID, hostname string) report.Topology {
	var localCIDRs []string
	if localNets, err := net.InterfaceAddrs(); err == nil {
		// Not all networks are IP networks.
		for _, localNet := range localNets {
			if ipNet, ok := localNet.(*net.IPNet); ok {
				localCIDRs = append(localCIDRs, ipNet.String())
			}
		}
	}
	t := report.MakeTopology()
	t.NodeMetadatas[hostID] = report.NodeMetadata{
		"ts":             time.Now().UTC().Format(time.RFC3339Nano),
		"host_name":      hostname,
		"local_networks": strings.Join(localCIDRs, " "),
		"os":             runtime.GOOS,
		"load":           getLoad(),
	}
	return t
}

func interrupt() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}
