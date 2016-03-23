package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-metrics"
	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/weave/common"

	"github.com/weaveworks/scope/common/hostname"
	"github.com/weaveworks/scope/common/network"
	"github.com/weaveworks/scope/common/sanitize"
	"github.com/weaveworks/scope/common/weave"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/probe/appclient"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/endpoint/procspy"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

const (
	versionCheckPeriod = 6 * time.Hour
)

func check() {
	handleResponse := func(r *checkpoint.CheckResponse, err error) {
		if err != nil {
			log.Errorf("Error checking version: %v", err)
		} else if r.Outdated {
			log.Infof("Scope version %s is available; please update at %s",
				r.CurrentVersion, r.CurrentDownloadURL)
		}
	}

	// Start background version checking
	params := checkpoint.CheckParams{
		Product: "scope-probe",
		Version: version,
	}
	resp, err := checkpoint.Check(&params)
	handleResponse(resp, err)
	checkpoint.CheckInterval(&params, versionCheckPeriod, handleResponse)
}

// Main runs the probe
func probeMain() {
	var (
		targets         = []string{fmt.Sprintf("localhost:%d", xfer.AppPort)}
		token           = flag.String("token", "default-token", "probe token")
		httpListen      = flag.String("http.listen", "", "listen address for HTTP profiling and instrumentation server")
		publishInterval = flag.Duration("publish.interval", 3*time.Second, "publish (output) interval")
		spyInterval     = flag.Duration("spy.interval", time.Second, "spy (scan) interval")
		spyProcs        = flag.Bool("processes", true, "report processes (needs root)")
		procRoot        = flag.String("proc.root", "/proc", "location of the proc filesystem")
		useConntrack    = flag.Bool("conntrack", true, "also use conntrack to track connections")
		insecure        = flag.Bool("insecure", false, "(SSL) explicitly allow \"insecure\" SSL connections and transfers")
		logPrefix       = flag.String("log.prefix", "<probe>", "prefix for each log line")
		logLevel        = flag.String("log.level", "info", "logging threshold level: debug|info|warn|error|fatal|panic")

		dockerEnabled  = flag.Bool("docker", false, "collect Docker-related attributes for processes")
		dockerInterval = flag.Duration("docker.interval", 10*time.Second, "how often to update Docker attributes")
		dockerBridge   = flag.String("docker.bridge", "docker0", "the docker bridge name")

		kubernetesEnabled  = flag.Bool("kubernetes", false, "collect kubernetes-related attributes for containers, should only be enabled on the master node")
		kubernetesAPI      = flag.String("kubernetes.api", "", "Address of kubernetes master api")
		kubernetesInterval = flag.Duration("kubernetes.interval", 10*time.Second, "how often to do a full resync of the kubernetes data")

		weaveAddr      = flag.String("weave.addr", "127.0.0.1:6784", "IP address & port of the Weave router")
		weaveDNSTarget = flag.String("weave.hostname", fmt.Sprintf("scope.weave.local:%d", xfer.AppPort), "Hostname to lookup in weaveDNS")
	)
	flag.Parse()

	setLogLevel(*logLevel)
	setLogFormatter(*logPrefix)

	// Setup in memory metrics sink
	inm := metrics.NewInmemSink(time.Minute, 2*time.Minute)
	sig := metrics.DefaultInmemSignal(inm)
	defer sig.Stop()
	metrics.NewGlobal(metrics.DefaultConfig("scope-probe"), inm)

	defer log.Info("probe exiting")

	if *spyProcs && os.Getegid() != 0 {
		log.Warn("-process=true, but that requires root to find everything")
	}

	rand.Seed(time.Now().UnixNano())
	probeID := strconv.FormatInt(rand.Int63(), 16)
	var (
		hostName = hostname.Get()
		hostID   = hostName // TODO(pb): we should sanitize the hostname
	)
	log.Infof("probe starting, version %s, ID %s", version, probeID)
	go check()

	if len(flag.Args()) > 0 {
		targets = flag.Args()
	}
	log.Infof("publishing to: %s", strings.Join(targets, ", "))

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

	hostControls := host.NewControls(clients)
	defer hostControls.Stop()

	resolver := appclient.NewResolver(targets, net.LookupIP, clients.Set)
	defer resolver.Stop()

	processCache := process.NewCachingWalker(process.NewWalker(*procRoot))
	scanner := procspy.NewConnectionScanner(processCache)

	endpointReporter := endpoint.NewReporter(hostID, hostName, *spyProcs, *useConntrack, scanner)
	defer endpointReporter.Stop()

	p := probe.New(*spyInterval, *publishInterval, clients)
	p.AddTicker(processCache)
	p.AddReporter(
		endpointReporter,
		host.NewReporter(hostID, hostName, probeID),
		process.NewReporter(processCache, hostID, process.GetDeltaTotalJiffies),
	)
	p.AddTagger(probe.NewTopologyTagger(), host.NewTagger(hostID))

	if *dockerEnabled {
		if err := report.AddLocalBridge(*dockerBridge); err != nil {
			log.Errorf("Docker: problem with bridge %s: %v", *dockerBridge, err)
		}
		if registry, err := docker.NewRegistry(*dockerInterval, clients, true); err == nil {
			defer registry.Stop()
			p.AddTagger(docker.NewTagger(registry, processCache))
			p.AddReporter(docker.NewReporter(registry, hostID, probeID, p))
		} else {
			log.Errorf("Docker: failed to start registry: %v", err)
		}
	}

	if *kubernetesEnabled {
		if client, err := kubernetes.NewClient(*kubernetesAPI, *kubernetesInterval); err == nil {
			defer client.Stop()
			p.AddReporter(kubernetes.NewReporter(client))
		} else {
			log.Errorf("Kubernetes: failed to start client: %v", err)
			log.Errorf("Kubernetes: make sure to run Scope inside a POD with a service account or provide a valid kubernetes.api url")
		}
	}

	if *weaveAddr != "" {
		client := weave.NewClient(sanitize.URL("http://", 6784, "")(*weaveAddr))
		weave := overlay.NewWeave(hostID, client)
		defer weave.Stop()
		p.AddTagger(weave)
		p.AddReporter(weave)

		dockerBridgeIP, err := network.GetFirstAddressOf(*dockerBridge)
		if err != nil {
			log.Println("Error getting docker bridge ip:", err)
		} else {
			weaveDNSLookup := appclient.LookupUsing(dockerBridgeIP + ":53")
			weaveResolver := appclient.NewResolver([]string{*weaveDNSTarget}, weaveDNSLookup, clients.Set)
			defer weaveResolver.Stop()
		}
	}

	if *httpListen != "" {
		go func() {
			log.Infof("Profiling data being exported to %s", *httpListen)
			log.Infof("go tool pprof http://%s/debug/pprof/{profile,heap,block}", *httpListen)
			log.Infof("Profiling endpoint %s terminated: %v", *httpListen, http.ListenAndServe(*httpListen, nil))
		}()
	}

	p.Start()
	defer p.Stop()

	common.SignalHandlerLoop()
}
