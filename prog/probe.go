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

	"$GITHUB_URI/common/hostname"
	"$GITHUB_URI/common/network"
	"$GITHUB_URI/common/sanitize"
	"$GITHUB_URI/common/weave"
	"$GITHUB_URI/common/xfer"
	"$GITHUB_URI/probe"
	"$GITHUB_URI/probe/appclient"
	"$GITHUB_URI/probe/controls"
	"$GITHUB_URI/probe/docker"
	"$GITHUB_URI/probe/endpoint"
	"$GITHUB_URI/probe/endpoint/procspy"
	"$GITHUB_URI/probe/host"
	"$GITHUB_URI/probe/kubernetes"
	"$GITHUB_URI/probe/overlay"
	"$GITHUB_URI/probe/plugins"
	"$GITHUB_URI/probe/process"
	"$GITHUB_URI/report"
)

const (
	versionCheckPeriod = 6 * time.Hour
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
	if !flags.noApp {
		targets = append(targets, fmt.Sprintf("localhost:%d", xfer.AppPort))
	}
	if len(flag.Args()) > 0 {
		targets = append(targets, flag.Args()...)
	}
	log.Infof("publishing to: %s", strings.Join(targets, ", "))

	probeConfig := appclient.ProbeConfig{
		Token:    flags.token,
		ProbeID:  probeID,
		Insecure: flags.insecure,
	}
	clients := appclient.NewMultiAppClient(func(hostname, endpoint string) (appclient.AppClient, error) {
		return appclient.NewAppClient(
			probeConfig, hostname, endpoint,
			xfer.ControlHandlerFunc(controls.HandleControlRequest),
		)
	})
	defer clients.Stop()

	dnsLookupFn := net.LookupIP
	if flags.resolver != "" {
		dnsLookupFn = appclient.LookupUsing(flags.resolver)
	}
	resolver := appclient.NewResolver(targets, dnsLookupFn, clients.Set)
	defer resolver.Stop()

	processCache := process.NewCachingWalker(process.NewWalker(flags.procRoot))
	scanner := procspy.NewConnectionScanner(processCache)

	endpointReporter := endpoint.NewReporter(hostID, hostName, flags.spyProcs, flags.useConntrack, scanner)
	defer endpointReporter.Stop()

	p := probe.New(flags.spyInterval, flags.publishInterval, clients)
	p.AddTicker(processCache)
	hostReporter := host.NewReporter(hostID, hostName, probeID, version, clients)
	defer hostReporter.Stop()
	p.AddReporter(
		endpointReporter,
		hostReporter,
		process.NewReporter(processCache, hostID, process.GetDeltaTotalJiffies),
	)
	p.AddTagger(probe.NewTopologyTagger(), host.NewTagger(hostID))

	if flags.dockerEnabled {
		// Don't add the bridge in Kubernetes since container IPs are global and
		// shouldn't be scoped
		if !flags.kubernetesEnabled {
			if err := report.AddLocalBridge(flags.dockerBridge); err != nil {
				log.Errorf("Docker: problem with bridge %s: %v", flags.dockerBridge, err)
			}
		}
		if registry, err := docker.NewRegistry(flags.dockerInterval, clients, true, hostID); err == nil {
			defer registry.Stop()
			p.AddTagger(docker.NewTagger(registry, processCache))
			p.AddReporter(docker.NewReporter(registry, hostID, probeID, p))
		} else {
			log.Errorf("Docker: failed to start registry: %v", err)
		}
	}

	if flags.kubernetesEnabled {
		if client, err := kubernetes.NewClient(flags.kubernetesAPI, flags.kubernetesInterval); err == nil {
			defer client.Stop()
			reporter := kubernetes.NewReporter(client, clients, probeID, hostID, p)
			defer reporter.Stop()
			p.AddReporter(reporter)
			p.AddTagger(reporter)
		} else {
			log.Errorf("Kubernetes: failed to start client: %v", err)
			log.Errorf("Kubernetes: make sure to run Scope inside a POD with a service account or provide a valid kubernetes.api url")
		}
	}

	if flags.weaveAddr != "" {
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
