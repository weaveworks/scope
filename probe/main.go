package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/weaveworks/procspy"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

// Set during buildtime.
var version = "unknown"

func main() {
	var (
		httpListen         = flag.String("http.listen", "", "listen address for HTTP profiling and instrumentation server")
		publishInterval    = flag.Duration("publish.interval", 3*time.Second, "publish (output) interval")
		spyInterval        = flag.Duration("spy.interval", time.Second, "spy (scan) interval")
		listen             = flag.String("listen", ":"+strconv.Itoa(xfer.ProbePort), "listen address")
		prometheusEndpoint = flag.String("prometheus.endpoint", "/metrics", "Prometheus metrics exposition endpoint (requires -http.listen)")
		spyProcs           = flag.Bool("processes", true, "report processes (needs root)")
		cgroupsRoot        = flag.String("cgroups.root", "", "if provided, enrich -processes with cgroup names from this root (e.g. /mnt/cgroups)")
		cgroupsInterval    = flag.Duration("cgroups.interval", 10*time.Second, "how often to update cgroup names")
		dockerMapper       = flag.Bool("docker", true, "collect Docker-related attributes for processes")
		dockerInterval     = flag.Duration("docker.interval", 10*time.Second, "how often to update Docker attributes")
		procRoot           = flag.String("proc.root", "/proc", "location of the proc filesystem")
	)
	flag.Parse()

	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("probe starting, version %s", version)

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
		log.Printf("warning: process reporting enabled, but that requires root to find everything")
	}

	publisher, err := xfer.NewTCPPublisher(*listen)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	pms := []processMapper{identityMapper{}}

	if *cgroupsRoot != "" {
		if fi, err := os.Stat(*cgroupsRoot); err == nil && fi.IsDir() {
			log.Printf("enriching -processes with cgroup names from %s", *cgroupsRoot)
			pms = append(pms, newCgroupMapper(*cgroupsRoot, *cgroupsInterval))
		} else {
			log.Printf("-cgroups.root=%s: %v", *cgroupsRoot, err)
		}
	}

	if *dockerMapper {
		docker := newDockerMapper(*procRoot, *dockerInterval)
		pms = append(pms,
			docker.idMapper(),
			docker.nameMapper(),
			docker.imageIDMapper(),
			docker.imageNameMapper(),
		)
	}

	log.Printf("listening on %s", *listen)

	go func() {
		var (
			hostname = hostname()
			nodeID   = hostname // TODO: we should sanitize the hostname
			pubTick  = time.Tick(*publishInterval)
			spyTick  = time.Tick(*spyInterval)
			r        = report.NewReport()
		)

		for {
			select {
			case <-pubTick:
				publishTicks.WithLabelValues().Add(1)
				r.HostMetadatas[nodeID] = hostMetadata(hostname)
				publisher.Publish(r)
				r = report.NewReport()

			case <-spyTick:
				r.Merge(spy(hostname, hostname, *spyProcs, pms))
				// log.Printf("merged report:\n%#v\n", r)
			}
		}
	}()

	log.Printf("%s", <-interrupt())
}

func interrupt() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}

// hostMetadata produces an instantaneous HostMetadata for this host. No need
// to do this more than once per published report.
func hostMetadata(hostname string) report.HostMetadata {
	loadOne, loadFive, loadFifteen := getLoads()

	host := report.HostMetadata{
		Timestamp:   time.Now().UTC(),
		Hostname:    hostname,
		OS:          runtime.GOOS,
		LoadOne:     loadOne,
		LoadFive:    loadFive,
		LoadFifteen: loadFifteen,
	}

	if localNets, err := net.InterfaceAddrs(); err == nil {
		// Not all networks are IP networks.
		for _, localNet := range localNets {
			if net, ok := localNet.(*net.IPNet); ok {
				host.LocalNets = append(host.LocalNets, net)
			}
		}
	}

	return host
}
