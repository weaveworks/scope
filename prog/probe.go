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
	"github.com/weaveworks/scope/probe/plugins"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

const (
	versionCheckPeriod = 6 * time.Hour
	defaultServiceHost = "cloud.weave.works:443"
)

var pluginAPIVersion = "1"

func check(flags map[string]string) {
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
		Flags:   flags,
	}
	resp, err := checkpoint.Check(&params)
	handleResponse(resp, err)
	checkpoint.CheckInterval(&params, versionCheckPeriod, handleResponse)
}

// Main runs the probe
func probeMain(flags probeFlags) {
	setLogLevel(flags.logLevel)
	setLogFormatter(flags.logPrefix)

	// Setup in memory metrics sink
	inm := metrics.NewInmemSink(time.Minute, 2*time.Minute)
	sig := metrics.DefaultInmemSignal(inm)
	defer sig.Stop()
	metrics.NewGlobal(metrics.DefaultConfig("scope-probe"), inm)
	logCensoredArgs()
	defer log.Info("probe exiting")

	if flags.spyProcs && os.Getegid() != 0 {
		log.Warn("--probe.process=true, but that requires root to find everything")
	}

	rand.Seed(time.Now().UnixNano())
	var (
		probeID  = strconv.FormatInt(rand.Int63(), 16)
		hostName = hostname.Get()
		hostID   = hostName // TODO(pb): we should sanitize the hostname
	)
	log.Infof("probe starting, version %s, ID %s", version, probeID)
	log.Infof("command line: %v", os.Args)
	checkpointFlags := map[string]string{}
	if flags.kubernetesEnabled {
		checkpointFlags["kubernetes_enabled"] = "true"
	}
	go check(checkpointFlags)

	var targets = []string{}
	if flags.token != "" {
		// service mode
		if len(flag.Args()) == 0 {
			targets = append(targets, defaultServiceHost)
		}
	} else if !flags.noApp {
		targets = append(targets, fmt.Sprintf("localhost:%d", xfer.AppPort))
	}
	targets = append(targets, flag.Args()...)
	log.Infof("publishing to: %s", strings.Join(targets, ", "))

	probeConfig := appclient.ProbeConfig{
		Token:        flags.token,
		ProbeVersion: version,
		ProbeID:      probeID,
		Insecure:     flags.insecure,
	}
	handlerRegistry := controls.NewDefaultHandlerRegistry()
	clientFactory := func(hostname, endpoint string) (appclient.AppClient, error) {
		return appclient.NewAppClient(
			probeConfig, hostname, endpoint,
			xfer.ControlHandlerFunc(handlerRegistry.HandleControlRequest),
		)
	}
	clients := appclient.NewMultiAppClient(clientFactory, flags.noControls)
	defer clients.Stop()

	dnsLookupFn := net.LookupIP
	if flags.resolver != "" {
		dnsLookupFn = appclient.LookupUsing(flags.resolver)
	}
	resolver := appclient.NewResolver(targets, dnsLookupFn, clients.Set)
	defer resolver.Stop()

	p := probe.New(flags.spyInterval, flags.publishInterval, clients, flags.noControls)

	hostReporter := host.NewReporter(hostID, hostName, probeID, version, clients, handlerRegistry)
	defer hostReporter.Stop()
	p.AddReporter(hostReporter)
	p.AddTagger(probe.NewTopologyTagger(), host.NewTagger(hostID))

	var processCache *process.CachingWalker
	var scanner procspy.ConnectionScanner
	if flags.procEnabled {
		processCache = process.NewCachingWalker(process.NewWalker(flags.procRoot))
		scanner = procspy.NewConnectionScanner(processCache)
		p.AddTicker(processCache)
		p.AddReporter(process.NewReporter(processCache, hostID, process.GetDeltaTotalJiffies))
	}

	endpointReporter := endpoint.NewReporter(hostID, hostName, flags.spyProcs, flags.useConntrack, flags.procEnabled, scanner)
	defer endpointReporter.Stop()
	p.AddReporter(endpointReporter)

	if flags.dockerEnabled {
		// Don't add the bridge in Kubernetes since container IPs are global and
		// shouldn't be scoped
		if !flags.kubernetesEnabled {
			if err := report.AddLocalBridge(flags.dockerBridge); err != nil {
				log.Errorf("Docker: problem with bridge %s: %v", flags.dockerBridge, err)
			}
		}
		if registry, err := docker.NewRegistry(flags.dockerInterval, clients, true, hostID, handlerRegistry); err == nil {
			defer registry.Stop()
			if flags.procEnabled {
				p.AddTagger(docker.NewTagger(registry, processCache))
			}
			p.AddReporter(docker.NewReporter(registry, hostID, probeID, p))
		} else {
			log.Errorf("Docker: failed to start registry: %v", err)
		}
	}

	if flags.kubernetesEnabled {
		if client, err := kubernetes.NewClient(flags.kubernetesAPI, flags.kubernetesInterval); err == nil {
			defer client.Stop()
			reporter := kubernetes.NewReporter(client, clients, probeID, hostID, p, handlerRegistry)
			defer reporter.Stop()
			p.AddReporter(reporter)
			p.AddTagger(reporter)
		} else {
			log.Errorf("Kubernetes: failed to start client: %v", err)
			log.Errorf("Kubernetes: make sure to run Scope inside a POD with a service account or provide a valid kubernetes.api url")
		}
	}

	if flags.weaveEnabled {
		client := weave.NewClient(sanitize.URL("http://", 6784, "")(flags.weaveAddr))
		weave := overlay.NewWeave(hostID, client)
		defer weave.Stop()
		p.AddTagger(weave)
		p.AddReporter(weave)

		dockerBridgeIP, err := network.GetFirstAddressOf(flags.dockerBridge)
		if err != nil {
			log.Println("Error getting docker bridge ip:", err)
		} else {
			weaveDNSLookup := appclient.LookupUsing(dockerBridgeIP + ":53")
			weaveResolver := appclient.NewResolver([]string{flags.weaveHostname}, weaveDNSLookup, clients.Set)
			defer weaveResolver.Stop()
		}
	}

	pluginRegistry, err := plugins.NewRegistry(
		flags.pluginsRoot,
		pluginAPIVersion,
		map[string]string{
			"probe_id":    probeID,
			"api_version": pluginAPIVersion,
		},
	)
	if err != nil {
		log.Errorf("plugins: problem loading: %v", err)
	} else {
		defer pluginRegistry.Close()
		p.AddReporter(pluginRegistry)
	}

	if flags.httpListen != "" {
		go func() {
			log.Infof("Profiling data being exported to %s", flags.httpListen)
			log.Infof("go tool pprof http://%s/debug/pprof/{profile,heap,block}", flags.httpListen)
			log.Infof("Profiling endpoint %s terminated: %v", flags.httpListen, http.ListenAndServe(flags.httpListen, nil))
		}()
	}

	p.Start()
	defer p.Stop()

	common.SignalHandlerLoop()
}
