package main

import (
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-metrics"
	"github.com/prometheus/client_golang/prometheus"
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
	defaultServiceHost = "https://cloud.weave.works:443"
)

var (
	pluginAPIVersion = "1"
	dockerEndpoint   = "unix:///var/run/docker.sock"
)

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
func probeMain(flags probeFlags, targets []appclient.Target) {
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
	checkpointFlags := map[string]string{}
	if flags.kubernetesEnabled {
		checkpointFlags["kubernetes_enabled"] = "true"
	}
	go check(checkpointFlags)

	handlerRegistry := controls.NewDefaultHandlerRegistry()
	clientFactory := func(hostname string, url url.URL) (appclient.AppClient, error) {
		token := flags.token
		if url.User != nil {
			token = url.User.Username()
			url.User = nil // erase credentials, as we use a special header
		}
		probeConfig := appclient.ProbeConfig{
			Token:        token,
			ProbeVersion: version,
			ProbeID:      probeID,
			Insecure:     flags.insecure,
		}
		return appclient.NewAppClient(
			probeConfig, hostname, url,
			xfer.ControlHandlerFunc(handlerRegistry.HandleControlRequest),
		)
	}
	clients := appclient.NewMultiAppClient(clientFactory, flags.noControls)
	defer clients.Stop()

	dnsLookupFn := net.LookupIP
	if flags.resolver != "" {
		dnsLookupFn = appclient.LookupUsing(flags.resolver)
	}
	resolver, err := appclient.NewResolver(appclient.ResolverConfig{
		Targets: targets,
		Lookup:  dnsLookupFn,
		Set:     clients.Set,
	})
	if err != nil {
		log.Fatalf("Failed to create resolver: %v", err)
		return
	}
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

	dnsSnooper, err := endpoint.NewDNSSnooper()
	if err != nil {
		log.Errorf("Failed to start DNS snooper: nodes for external services will be less accurate: %s", err)
	} else {
		defer dnsSnooper.Stop()
	}

	endpointReporter := endpoint.NewReporter(endpoint.ReporterConfig{
		HostID:       hostID,
		HostName:     hostName,
		SpyProcs:     flags.spyProcs,
		UseConntrack: flags.useConntrack,
		WalkProc:     flags.procEnabled,
		ProcRoot:     flags.procRoot,
		BufferSize:   flags.conntrackBufferSize,
		Scanner:      scanner,
		DNSSnooper:   dnsSnooper,
	})
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
		if registry, err := docker.NewRegistry(flags.dockerInterval, clients, true, hostID, handlerRegistry, dockerEndpoint); err == nil {
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
		if client, err := kubernetes.NewClient(flags.kubernetesConfig); err == nil {
			defer client.Stop()
			reporter := kubernetes.NewReporter(client, clients, probeID, hostID, p, handlerRegistry)
			defer reporter.Stop()
			p.AddReporter(reporter)
			p.AddTagger(reporter)
		} else {
			log.Errorf("Kubernetes: failed to start client: %v", err)
			log.Errorf("Kubernetes: make sure to run Scope inside a POD with a service account or provide valid probe.kubernetes.* flags")
		}
	}

	if flags.weaveEnabled {
		client := weave.NewClient(sanitize.URL("http://", 6784, "")(flags.weaveAddr))
		weave, err := overlay.NewWeave(hostID, client, dockerEndpoint)
		if err != nil {
			log.Errorf("Weave: failed to start client: %v", err)
		} else {
			defer weave.Stop()
			p.AddTagger(weave)
			p.AddReporter(weave)

			dockerBridgeIP, err := network.GetFirstAddressOf(flags.dockerBridge)
			if err != nil {
				log.Errorf("Error getting docker bridge ip: %v", err)
			} else {
				weaveDNSLookup := appclient.LookupUsing(dockerBridgeIP + ":53")
				weaveTargets, err := appclient.ParseTargets([]string{flags.weaveHostname})
				if err != nil {
					log.Errorf("Failed to parse weave targets: %v", err)
				} else {
					weaveResolver, err := appclient.NewResolver(appclient.ResolverConfig{
						Targets: weaveTargets,
						Lookup:  weaveDNSLookup,
						Set:     clients.Set,
					})
					if err != nil {
						log.Errorf("Failed to create weave resolver: %v", err)
					} else {
						defer weaveResolver.Stop()
					}
				}
			}
		}
	}

	pluginRegistry, err := plugins.NewRegistry(
		flags.pluginsRoot,
		pluginAPIVersion,
		map[string]string{
			"probe_id":    probeID,
			"api_version": pluginAPIVersion,
		},
		handlerRegistry,
		p,
	)
	if err != nil {
		log.Errorf("plugins: problem loading: %v", err)
	} else {
		defer pluginRegistry.Close()
		p.AddReporter(pluginRegistry)
	}

	if flags.httpListen != "" {
		go func() {
			http.Handle("/metrics", prometheus.Handler())
			log.Infof("Profiling data being exported to %s", flags.httpListen)
			log.Infof("go tool proof http://%s/debug/pprof/{profile,heap,block}", flags.httpListen)
			log.Infof("Profiling endpoint %s terminated: %v", flags.httpListen, http.ListenAndServe(flags.httpListen, nil))
		}()
	}

	p.Start()
	defer p.Stop()

	common.SignalHandlerLoop()
}
