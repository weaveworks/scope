package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/weaveworks/weave/common"

	"github.com/weaveworks/scope/common/hostname"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/probe/appclient"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Main runs the probe
func probeMain() {
	var (
		targets            = []string{fmt.Sprintf("localhost:%d", xfer.AppPort), fmt.Sprintf("scope.weave.local:%d", xfer.AppPort)}
		token              = flag.String("token", "default-token", "probe token")
		httpListen         = flag.String("http.listen", "", "listen address for HTTP profiling and instrumentation server")
		publishInterval    = flag.Duration("publish.interval", 3*time.Second, "publish (output) interval")
		spyInterval        = flag.Duration("spy.interval", time.Second, "spy (scan) interval")
		spyProcs           = flag.Bool("processes", true, "report processes (needs root)")
		dockerEnabled      = flag.Bool("docker", false, "collect Docker-related attributes for processes")
		dockerInterval     = flag.Duration("docker.interval", 10*time.Second, "how often to update Docker attributes")
		dockerBridge       = flag.String("docker.bridge", "docker0", "the docker bridge name")
		kubernetesEnabled  = flag.Bool("kubernetes", false, "collect kubernetes-related attributes for containers, should only be enabled on the master node")
		kubernetesAPI      = flag.String("kubernetes.api", "", "Address of kubernetes master api")
		kubernetesInterval = flag.Duration("kubernetes.interval", 10*time.Second, "how often to do a full resync of the kubernetes data")
		weaveRouterAddr    = flag.String("weave.router.addr", "", "IP address or FQDN of the Weave router")
		procRoot           = flag.String("proc.root", "/proc", "location of the proc filesystem")
		useConntrack       = flag.Bool("conntrack", true, "also use conntrack to track connections")
		insecure           = flag.Bool("insecure", false, "(SSL) explicitly allow \"insecure\" SSL connections and transfers")
		logPrefix          = flag.String("log.prefix", "<probe>", "prefix for each log line")
	)
	flag.Parse()

	// Setup in memory metrics sink
	inm := metrics.NewInmemSink(time.Minute, 2*time.Minute)
	sig := metrics.DefaultInmemSignal(inm)
	defer sig.Stop()
	metrics.NewGlobal(metrics.DefaultConfig("scope-probe"), inm)

	if !strings.HasSuffix(*logPrefix, " ") {
		*logPrefix += " "
	}
	log.SetPrefix(*logPrefix)

	defer log.Print("probe exiting")

	if *spyProcs && os.Getegid() != 0 {
		log.Printf("warning: -process=true, but that requires root to find everything")
	}

	rand.Seed(time.Now().UnixNano())
	probeID := strconv.FormatInt(rand.Int63(), 16)
	var (
		hostName = hostname.Get()
		hostID   = hostName // TODO(pb): we should sanitize the hostname
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

	probeConfig := appclient.ProbeConfig{
		Token:    *token,
		ProbeID:  probeID,
		Insecure: *insecure,
	}
	clients := appclient.NewMultiAppClient(func(hostname, endpoint string) (appclient.AppClient, error) {
		return appclient.NewAppClient(
			probeConfig, hostname, endpoint,
			xfer.ControlHandlerFunc(controls.HandleControlRequest),
		)
	})
	defer clients.Stop()

	resolver := appclient.NewStaticResolver(targets, clients.Set)
	defer resolver.Stop()

	processCache := process.NewCachingWalker(process.NewWalker(*procRoot))

	endpointReporter := endpoint.NewReporter(hostID, hostName, *spyProcs, *useConntrack, processCache)
	defer endpointReporter.Stop()

	p := probe.New(*spyInterval, *publishInterval, clients)
	p.AddTicker(processCache)
	p.AddReporter(
		endpointReporter,
		host.NewReporter(hostID, hostName, localNets),
		process.NewReporter(processCache, hostID, process.GetDeltaTotalJiffies),
	)
	p.AddTagger(probe.NewTopologyTagger(), host.NewTagger(hostID, probeID))

	if *dockerEnabled {
		if err := report.AddLocalBridge(*dockerBridge); err != nil {
			log.Printf("Docker: problem with bridge %s: %v", *dockerBridge, err)
		}
		if registry, err := docker.NewRegistry(*dockerInterval, clients); err == nil {
			defer registry.Stop()
			p.AddTagger(docker.NewTagger(registry, processCache))
			p.AddReporter(docker.NewReporter(registry, hostID, p))
		} else {
			log.Printf("Docker: failed to start registry: %v", err)
		}
	}

	if *kubernetesEnabled {
		if client, err := kubernetes.NewClient(*kubernetesAPI, *kubernetesInterval); err == nil {
			defer client.Stop()
			p.AddReporter(kubernetes.NewReporter(client))
		} else {
			log.Printf("Kubernetes: failed to start client: %v", err)
			log.Printf("Kubernetes: make sure to run Scope inside a POD with a service account or provide a valid kubernetes.api url")
		}
	}

	if *weaveRouterAddr != "" {
		weave := overlay.NewWeave(hostID, *weaveRouterAddr)
		defer weave.Stop()
		p.AddTicker(weave)
		p.AddTagger(weave)
		p.AddReporter(weave)
	}

	if *httpListen != "" {
		go func() {
			log.Printf("Profiling data being exported to %s", *httpListen)
			log.Printf("go tool pprof http://%s/debug/pprof/{profile,heap,block}", *httpListen)
			log.Printf("Profiling endpoint %s terminated: %v", *httpListen, http.ListenAndServe(*httpListen, nil))
		}()
	}

	p.Start()
	defer p.Stop()

	common.SignalHandlerLoop()
}
