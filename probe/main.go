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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/process"
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
		kubernetesEnabled  = flag.Bool("kubernetes", false, "collect kubernetes-related attributes for containers, should only be enabled on the master node")
		kubernetesAPI      = flag.String("kubernetes.api", "http://localhost:8080", "Address of kubernetes master api")
		kubernetesInterval = flag.Duration("kubernetes.interval", 10*time.Second, "how often to do a full resync of the kubernetes data")
		weaveRouterAddr    = flag.String("weave.router.addr", "", "IP address or FQDN of the Weave router")
		procRoot           = flag.String("proc.root", "/proc", "location of the proc filesystem")
		printVersion       = flag.Bool("version", false, "print version number and exit")
		useConntrack       = flag.Bool("conntrack", true, "also use conntrack to track connections")
		insecure           = flag.Bool("insecure", false, "(SSL) explicitly allow \"insecure\" SSL connections and transfers")
		logPrefix          = flag.String("log.prefix", "<probe>", "prefix for each log line")
	)
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	if !strings.HasSuffix(*logPrefix, " ") {
		*logPrefix += " "
	}
	log.SetPrefix(*logPrefix)

	defer log.Print("probe exiting")

	if *spyProcs && os.Getegid() != 0 {
		log.Printf("warning: -process=true, but that requires root to find everything")
	}

	var (
		hostName = hostname()
		hostID   = hostName // TODO(pb): we should sanitize the hostname
		probeID  = hostName // TODO(pb): does this need to be a random string instead?
	)
	log.Printf("probe starting, version %s, ID %s", version, probeID)

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

	if len(flag.Args()) > 0 {
		targets = flag.Args()
	}
	log.Printf("publishing to: %s", strings.Join(targets, ", "))

	factory := func(hostname, endpoint string) (string, xfer.Publisher, error) {
		id, publisher, err := xfer.NewHTTPPublisher(hostname, endpoint, *token, probeID, *insecure)
		if err != nil {
			return "", nil, err
		}
		return id, xfer.NewBackgroundPublisher(publisher), nil
	}

	publishers := xfer.NewMultiPublisher(factory)
	defer publishers.Stop()

	resolver := newStaticResolver(targets, publishers.Set)
	defer resolver.Stop()

	endpointReporter := endpoint.NewReporter(hostID, hostName, *spyProcs, *useConntrack)
	defer endpointReporter.Stop()

	processCache := process.NewCachingWalker(process.NewWalker(*procRoot))

	var (
		tickers   = []Ticker{processCache}
		reporters = []Reporter{endpointReporter, host.NewReporter(hostID, hostName, localNets), process.NewReporter(processCache, hostID)}
		taggers   = []Tagger{newTopologyTagger(), host.NewTagger(hostID)}
	)

	dockerTagger, dockerReporter, dockerRegistry := func() (*docker.Tagger, *docker.Reporter, docker.Registry) {
		if !*dockerEnabled {
			return nil, nil, nil
		}
		if err := report.AddLocalBridge(*dockerBridge); err != nil {
			log.Printf("Docker: problem with bridge %s: %v", *dockerBridge, err)
			return nil, nil, nil
		}
		registry, err := docker.NewRegistry(*dockerInterval)
		if err != nil {
			log.Printf("Docker: failed to start registry: %v", err)
			return nil, nil, nil
		}
		return docker.NewTagger(registry, processCache), docker.NewReporter(registry, hostID), registry
	}()
	if dockerTagger != nil {
		taggers = append(taggers, dockerTagger)
	}
	if dockerReporter != nil {
		reporters = append(reporters, dockerReporter)
	}
	if dockerRegistry != nil {
		defer dockerRegistry.Stop()
	}

	if *kubernetesEnabled {
		if client, err := kubernetes.NewClient(*kubernetesAPI, *kubernetesInterval); err == nil {
			defer client.Stop()
			reporters = append(reporters, kubernetes.NewReporter(client))
		} else {
			log.Printf("Kubernetes: failed to start client: %v", err)
		}
	}

	if *weaveRouterAddr != "" {
		weave := overlay.NewWeave(hostID, *weaveRouterAddr)
		tickers = append(tickers, weave)
		taggers = append(taggers, weave)
		reporters = append(reporters, weave)
	}

	if *httpListen != "" {
		go func() {
			log.Printf("Profiling data being exported to %s", *httpListen)
			log.Printf("go tool pprof http://%s/debug/pprof/{profile,heap,block}", *httpListen)
			if *prometheusEndpoint != "" {
				log.Printf("exposing Prometheus endpoint at %s%s", *httpListen, *prometheusEndpoint)
				http.Handle(*prometheusEndpoint, makePrometheusHandler())
			}
			log.Printf("Profiling endpoint %s terminated: %v", *httpListen, http.ListenAndServe(*httpListen, nil))
		}()
	}

	quit, done := make(chan struct{}), sync.WaitGroup{}
	done.Add(2)
	defer func() { done.Wait() }() // second, wait for the main loops to be killed
	defer close(quit)              // first, kill the main loops

	var rpt syncReport
	rpt.swap(report.MakeReport())

	go func() {
		defer done.Done()
		spyTick := time.Tick(*spyInterval)

		for {
			select {
			case <-spyTick:
				start := time.Now()
				for _, ticker := range tickers {
					if err := ticker.Tick(); err != nil {
						log.Printf("error doing ticker: %v", err)
					}
				}

				localReport := rpt.copy()
				localReport = localReport.Merge(doReport(reporters))
				localReport = Apply(localReport, taggers)
				rpt.swap(localReport)

				if took := time.Since(start); took > *spyInterval {
					log.Printf("report generation took too long (%s)", took)
				}

			case <-quit:
				return
			}
		}
	}()

	go func() {
		defer done.Done()
		var (
			pubTick = time.Tick(*publishInterval)
			p       = xfer.NewReportPublisher(publishers)
		)

		for {
			select {
			case <-pubTick:
				publishTicks.WithLabelValues().Add(1)
				localReport := rpt.swap(report.MakeReport())
				localReport.Window = *publishInterval
				if err := p.Publish(localReport); err != nil {
					log.Printf("publish: %v", err)
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

func interrupt() <-chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}

type syncReport struct {
	mtx sync.RWMutex
	rpt report.Report
}

func (r *syncReport) swap(other report.Report) report.Report {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	old := r.rpt
	r.rpt = other
	return old
}

func (r *syncReport) copy() report.Report {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return r.rpt.Copy()
}
