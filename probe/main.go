package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/scope/probe/proc"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/probe/sniff"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

var version = "dev" // set at build time

func main() {
	var (
		targets            = []string{fmt.Sprintf("localhost:%d", xfer.AppPort), fmt.Sprintf("scope.weave.local:%d", xfer.AppPort)}
		token              = flag.String("token", "default-token", "probe token")
		httpListen         = flag.String("http.listen", "", "listen address for HTTP profiling and instrumentation server")
		publishInterval    = flag.Duration("publish.interval", 3*time.Second, "publish (output) interval")
		spyInterval        = flag.Duration("spy.interval", time.Second, "spy (scan) interval")
		prometheusEndpoint = flag.String("prometheus.endpoint", "/metrics", "Prometheus metrics exposition endpoint (requires -http.listen)")
		spyProcs           = flag.Bool("processes", true, "report processes (needs root)")
		dockerEnabled      = flag.Bool("docker", false, "collect Docker-related attributes for processes")
		dockerInterval     = flag.Duration("docker.interval", 10*time.Second, "how often to update Docker attributes")
		dockerBridge       = flag.String("docker.bridge", "docker0", "the docker bridge name")
		weaveRouterAddr    = flag.String("weave.router.addr", "", "IP address or FQDN of the Weave router")
		procRoot           = flag.String("proc.root", "/proc", "location of the proc filesystem")
		captureEnabled     = flag.Bool("capture", false, "perform sampled packet capture")
		captureInterfaces  = flag.String("capture.interfaces", interfaces(), "packet capture on these interfaces")
		captureOn          = flag.Duration("capture.on", 1*time.Second, "packet capture duty cycle 'on'")
		captureOff         = flag.Duration("capture.off", 5*time.Second, "packet capture duty cycle 'off'")
		printVersion       = flag.Bool("version", false, "print version number and exit")
		useConntrack       = flag.Bool("conntrack", true, "also use conntrack to track connections")
	)
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	var (
		hostName = hostname()
		hostID   = hostName // TODO: we should sanitize the hostname
		probeID  = hostName // TODO: does this need to be a random string instead?
	)
	log.Printf("probe starting, version %s, ID %s", version, probeID)

	if len(flag.Args()) > 0 {
		targets = flag.Args()
	}
	log.Printf("publishing to: %s", strings.Join(targets, ", "))

	proc.SetProcRoot(*procRoot)

	if *httpListen != "" {
		log.Printf("profiling data being exported to %s", *httpListen)
		log.Printf("go tool pprof http://%s/debug/pprof/{profile,heap,block}", *httpListen)
		if *prometheusEndpoint != "" {
			log.Printf("exposing Prometheus endpoint at %s%s", *httpListen, *prometheusEndpoint)
			http.Handle(*prometheusEndpoint, makePrometheusHandler())
		}
		go func() {
			err := http.ListenAndServe(*httpListen, nil)
			log.Print(err)
		}()
	}

	if *spyProcs && os.Getegid() != 0 {
		log.Printf("warning: process reporting enabled, but that requires root to find everything")
	}

	publisherFactory := func(target string) (xfer.Publisher, error) {
		publisher, err := xfer.NewHTTPPublisher(target, *token, probeID)
		if err != nil {
			return nil, err
		}
		return xfer.NewBackgroundPublisher(publisher), nil
	}
	publishers := xfer.NewMultiPublisher(publisherFactory)
	defer publishers.Stop()
	resolver := newStaticResolver(targets, publishers.Add)
	defer resolver.Stop()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}
	localNets := report.Networks{}
	for _, addr := range addrs {
		// Not all addrs are IPNets.
		if ipNet, ok := addr.(*net.IPNet); ok {
			localNets = append(localNets, ipNet)
		}
	}

	var (
		endpointReporter = endpoint.NewReporter(hostID, hostName, *spyProcs, *useConntrack)
		processCache     = proc.NewCachingWalker(proc.NewWalker(*procRoot))
		tickers          = []Ticker{processCache}
		reporters        = []Reporter{
			endpointReporter,
			host.NewReporter(hostID, hostName, localNets),
			process.NewReporter(processCache, hostID),
		}
		taggers = []Tagger{newTopologyTagger(), host.NewTagger(hostID)}
	)
	defer endpointReporter.Stop()

	if *dockerEnabled {
		if err := report.AddLocalBridge(*dockerBridge); err != nil {
			log.Fatalf("failed to get docker bridge address: %v", err)
		}

		dockerRegistry, err := docker.NewRegistry(*dockerInterval)
		if err != nil {
			log.Fatalf("failed to start docker registry: %v", err)
		}
		defer dockerRegistry.Stop()

		taggers = append(taggers, docker.NewTagger(dockerRegistry, processCache))
		reporters = append(reporters, docker.NewReporter(dockerRegistry, hostID))
	}

	if *weaveRouterAddr != "" {
		weave, err := overlay.NewWeave(hostID, *weaveRouterAddr)
		if err != nil {
			log.Fatalf("failed to start Weave tagger: %v", err)
		}
		tickers = append(tickers, weave)
		taggers = append(taggers, weave)
		reporters = append(reporters, weave)
	}

	if *captureEnabled {
		var sniffers int
		for _, iface := range strings.Split(*captureInterfaces, ",") {
			source, err := sniff.NewSource(iface)
			if err != nil {
				log.Printf("warning: %v", err)
				continue
			}
			defer source.Close()
			log.Printf("capturing packets on %s", iface)
			reporters = append(reporters, sniff.New(hostID, localNets, source, *captureOn, *captureOff))
			sniffers++
		}
		// Packet capture can block OS threads on Linux, so we need to provide
		// sufficient overhead in GOMAXPROCS.
		if have, want := runtime.GOMAXPROCS(-1), (sniffers + 1); have < want {
			runtime.GOMAXPROCS(want)
		}
	}

	quit := make(chan struct{})
	defer close(quit)
	go func() {
		var (
			pubTick = time.Tick(*publishInterval)
			spyTick = time.Tick(*spyInterval)
			r       = report.MakeReport()
			p       = xfer.NewReportPublisher(publishers)
		)

		for {
			select {
			case <-pubTick:
				publishTicks.WithLabelValues().Add(1)
				r.Window = *publishInterval
				if err := p.Publish(r); err != nil {
					log.Printf("publish: %v", err)
				}
				r = report.MakeReport()

			case <-spyTick:
				start := time.Now()

				for _, ticker := range tickers {
					if err := ticker.Tick(); err != nil {
						log.Printf("error doing ticker: %v", err)
					}
				}

				r = r.Merge(doReport(reporters))
				r = Apply(r, taggers)

				if took := time.Since(start); took > *spyInterval {
					log.Printf("report generation took too long (%s)", took)
				}

			case <-quit:
				return
			}
		}
	}()
	log.Printf("%s", <-interrupt())
}

func doReport(reporters []Reporter) report.Report {
	reports := make(chan report.Report, len(reporters))
	for _, rep := range reporters {
		go func(rep Reporter) {
			newReport, err := rep.Report()
			if err != nil {
				log.Printf("error generating report: %v", err)
				newReport = report.MakeReport() // empty is OK to merge
			}
			reports <- newReport
		}(rep)
	}

	result := report.MakeReport()
	for i := 0; i < cap(reports); i++ {
		result = result.Merge(<-reports)
	}
	return result
}

func interrupt() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}

func interfaces() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print(err)
		return ""
	}
	a := make([]string, 0, len(ifaces))
	for _, iface := range ifaces {
		a = append(a, iface.Name)
	}
	return strings.Join(a, ",")
}
