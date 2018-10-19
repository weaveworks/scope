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

	"github.com/armon/go-metrics"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/network"
	"github.com/weaveworks/common/sanitize"
	"github.com/weaveworks/common/signals"
	"github.com/weaveworks/common/tracing"
	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/scope/common/hostname"
	"github.com/weaveworks/scope/common/weave"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe"
	"github.com/weaveworks/scope/probe/appclient"
	"github.com/weaveworks/scope/probe/awsecs"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/cri"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/plugins"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

const (
	versionCheckPeriod = 6 * time.Hour
	defaultServiceHost = "https://cloud.weave.works.:443"

	kubernetesRoleHost    = "host"
	kubernetesRoleCluster = "cluster"
)

var (
	pluginAPIVersion = "1"
)

func checkNewScopeVersion(flags probeFlags) {
	checkpointFlags := makeBaseCheckpointFlags()
	if flags.kubernetesEnabled {
		checkpointFlags["kubernetes_enabled"] = "true"
	}
	if flags.ecsEnabled {
		checkpointFlags["ecs_enabled"] = "true"
	}

	go func() {
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
			Flags:   checkpointFlags,
		}
		resp, err := checkpoint.Check(&params)
		handleResponse(resp, err)
		checkpoint.CheckInterval(&params, versionCheckPeriod, handleResponse)
	}()
}

func maybeExportProfileData(flags probeFlags) {
	if flags.httpListen != "" {
		go func() {
			http.Handle("/metrics", prometheus.Handler())
			log.Infof("Profiling data being exported to %s", flags.httpListen)
			log.Infof("go tool pprof http://%s/debug/pprof/{profile,heap,block}", flags.httpListen)
			log.Infof("Profiling endpoint %s terminated: %v", flags.httpListen, http.ListenAndServe(flags.httpListen, nil))
		}()
	}
}

// Main runs the probe
func probeMain(flags probeFlags, targets []appclient.Target) {
	setLogLevel(flags.logLevel)
	setLogFormatter(flags.logPrefix)

	traceCloser := tracing.NewFromEnv("scope-probe")
	defer traceCloser.Close()

	// Setup in memory metrics sink
	inm := metrics.NewInmemSink(time.Minute, 2*time.Minute)
	sig := metrics.DefaultInmemSignal(inm)
	defer sig.Stop()
	metrics.NewGlobal(metrics.DefaultConfig("scope-probe"), inm)
	logCensoredArgs()
	defer log.Info("probe exiting")

	switch flags.kubernetesRole {
	case "": // nothing special
	case kubernetesRoleHost:
		flags.kubernetesEnabled = true
	case kubernetesRoleCluster:
		flags.kubernetesKubeletPort = 0
		flags.kubernetesEnabled = true
		flags.spyProcs = false
		flags.procEnabled = false
		flags.useConntrack = false
		flags.useEbpfConn = false
	default:
		log.Warnf("unrecognized --probe.kubernetes.role: %s", flags.kubernetesRole)
	}

	if flags.spyProcs && os.Getegid() != 0 {
		log.Warn("--probe.proc.spy=true, but that requires root to find everything")
	}

	rand.Seed(time.Now().UnixNano())
	var (
		probeID  = strconv.FormatInt(rand.Int63(), 16)
		hostName = hostname.Get()
		hostID   = hostName // TODO(pb): we should sanitize the hostname
	)
	log.Infof("probe starting, version %s, ID %s", version, probeID)
	checkNewScopeVersion(flags)

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

	var clients interface {
		probe.ReportPublisher
		controls.PipeClient
	}
	if flags.printOnStdout {
		if len(targets) > 0 {
			log.Warnf("Dumping to stdout only: targets %v will be ignored", targets)
		}
		clients = new(struct {
			report.StdoutPublisher
			controls.DummyPipeClient
		})
	} else {
		multiClients := appclient.NewMultiAppClient(clientFactory, flags.noControls)
		defer multiClients.Stop()

		dnsLookupFn := net.LookupIP
		if flags.resolver != "" {
			dnsLookupFn = appclient.LookupUsing(flags.resolver)
		}
		resolver, err := appclient.NewResolver(appclient.ResolverConfig{
			Targets: targets,
			Lookup:  dnsLookupFn,
			Set:     multiClients.Set,
		})
		if err != nil {
			log.Fatalf("Failed to create resolver: %v", err)
			return
		}
		defer resolver.Stop()

		if flags.weaveEnabled && flags.weaveHostname != "" {
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
						Set:     multiClients.Set,
					})
					if err != nil {
						log.Errorf("Failed to create weave resolver: %v", err)
					} else {
						defer weaveResolver.Stop()
					}
				}
			}
		}
		clients = multiClients
	}

	p := probe.New(flags.spyInterval, flags.publishInterval, clients, flags.noControls)

	hostReporter := host.NewReporter(hostID, hostName, probeID, version, clients, handlerRegistry)
	defer hostReporter.Stop()
	p.AddReporter(hostReporter)
	p.AddTagger(probe.NewTopologyTagger(), host.NewTagger(hostID))

	var processCache *process.CachingWalker
	if flags.procEnabled {
		processCache = process.NewCachingWalker(process.NewWalker(flags.procRoot, false))
		p.AddTicker(processCache)
		p.AddReporter(process.NewReporter(processCache, hostID, process.GetDeltaTotalJiffies, flags.noCommandLineArguments))
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
		UseEbpfConn:  flags.useEbpfConn,
		ProcRoot:     flags.procRoot,
		BufferSize:   flags.conntrackBufferSize,
		ProcessCache: processCache,
		DNSSnooper:   dnsSnooper,
	})
	defer endpointReporter.Stop()
	p.AddReporter(endpointReporter)

	if flags.dockerEnabled {
		// Don't add the bridge in Kubernetes since container IPs are global and
		// shouldn't be scoped
		if flags.dockerBridge != "" && !flags.kubernetesEnabled {
			if err := report.AddLocalBridge(flags.dockerBridge); err != nil {
				log.Errorf("Docker: problem with bridge %s: %v", flags.dockerBridge, err)
			}
		}
		options := docker.RegistryOptions{
			Interval:               flags.dockerInterval,
			Pipes:                  clients,
			CollectStats:           true,
			HostID:                 hostID,
			HandlerRegistry:        handlerRegistry,
			NoCommandLineArguments: flags.noCommandLineArguments,
			NoEnvironmentVariables: flags.noEnvironmentVariables,
		}
		if registry, err := docker.NewRegistry(options); err == nil {
			defer registry.Stop()
			if flags.procEnabled {
				p.AddTagger(docker.NewTagger(registry, processCache))
			}
			p.AddReporter(docker.NewReporter(registry, hostID, probeID, p))
		} else {
			log.Errorf("Docker: failed to start registry: %v", err)
		}
	}

	if flags.criEnabled {
		client, err := cri.NewCRIClient(flags.criEndpoint)
		if err != nil {
			log.Errorf("CRI: failed to start registry: %v", err)
		} else {
			p.AddReporter(cri.NewReporter(client))
		}
	}

	if flags.kubernetesEnabled && flags.kubernetesRole != kubernetesRoleHost {
		if client, err := kubernetes.NewClient(flags.kubernetesClientConfig); err == nil {
			defer client.Stop()
			reporter := kubernetes.NewReporter(client, clients, probeID, hostID, p, handlerRegistry, flags.kubernetesNodeName, flags.kubernetesKubeletPort)
			defer reporter.Stop()
			p.AddReporter(reporter)
		} else {
			log.Errorf("Kubernetes: failed to start client: %v", err)
			log.Errorf("Kubernetes: make sure to run Scope inside a POD with a service account or provide valid probe.kubernetes.* flags")
		}
	}

	if flags.kubernetesEnabled {
		p.AddTagger(&kubernetes.Tagger{})
	}

	if flags.ecsEnabled {
		reporter := awsecs.Make(flags.ecsCacheSize, flags.ecsCacheExpiry, flags.ecsClusterRegion, handlerRegistry, probeID)
		defer reporter.Stop()
		p.AddReporter(reporter)
		p.AddTagger(reporter)
	}

	if flags.weaveEnabled {
		client := weave.NewClient(sanitize.URL("http://", 6784, "")(flags.weaveAddr))
		weave, err := overlay.NewWeave(hostID, client)
		if err != nil {
			log.Errorf("Weave: failed to start client: %v", err)
		} else {
			defer weave.Stop()
			p.AddTagger(weave)
			p.AddReporter(weave)
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

	maybeExportProfileData(flags)

	p.Start()
	signals.SignalHandlerLoop(
		logging.Logrus(log.StandardLogger()),
		p,
	)
}
