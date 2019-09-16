package main

import (
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/tylerb/graceful"

	billing "github.com/weaveworks/billing-client"
	"github.com/weaveworks/common/aws"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/middleware"
	"github.com/weaveworks/common/network"
	"github.com/weaveworks/common/signals"
	"github.com/weaveworks/common/tracing"
	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/app/multitenant"
	"github.com/weaveworks/scope/common/weave"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/docker"
)

const (
	memcacheUpdateInterval = 1 * time.Minute
	httpTimeout            = 90 * time.Second
)

var (
	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "scope",
		Name:      "request_duration_seconds",
		Help:      "Time in seconds spent serving HTTP requests.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "route", "status_code", "ws"})
)

func registerAppMetrics() {
	prometheus.MustRegister(requestDuration)
	billing.MustRegisterMetrics()
}

var registerAppMetricsOnce sync.Once

// Router creates the mux for all the various app components.
func router(collector app.Collector, controlRouter app.ControlRouter, pipeRouter app.PipeRouter, externalUI bool, capabilities map[string]bool, metricsGraphURL string) http.Handler {
	router := mux.NewRouter().SkipClean(true)

	// We pull in the http.DefaultServeMux to get the pprof routes
	router.PathPrefix("/debug/pprof").Handler(http.DefaultServeMux)
	router.Path("/metrics").Handler(prometheus.Handler())

	app.RegisterReportPostHandler(collector, router)
	app.RegisterControlRoutes(router, controlRouter)
	app.RegisterPipeRoutes(router, pipeRouter)
	app.RegisterTopologyRoutes(router, app.WebReporter{Reporter: collector, MetricsGraphURL: metricsGraphURL}, capabilities)
	app.RegisterAdminRoutes(router, collector)

	uiHandler := http.FileServer(GetFS(externalUI))
	router.PathPrefix("/ui").Name("static").Handler(
		middleware.PathRewrite(regexp.MustCompile("^/ui"), "").Wrap(
			uiHandler))
	router.PathPrefix("/").Name("static").Handler(uiHandler)

	middlewares := middleware.Merge(
		middleware.Instrument{
			RouteMatcher: router,
			Duration:     requestDuration,
		},
		middleware.Tracer{},
	)

	return middlewares.Wrap(router)
}

func collectorFactory(userIDer multitenant.UserIDer, collectorURL, s3URL, natsHostname string,
	memcacheConfig multitenant.MemcacheConfig, window time.Duration, maxTopNodes int, createTables bool) (app.Collector, error) {
	if collectorURL == "local" {
		return app.NewCollector(window), nil
	}

	parsed, err := url.Parse(collectorURL)
	if err != nil {
		return nil, err
	}

	switch parsed.Scheme {
	case "file":
		return app.NewFileCollector(parsed.Path, window)
	case "dynamodb":
		s3, err := url.Parse(s3URL)
		if err != nil {
			return nil, fmt.Errorf("Valid URL for s3 required: %v", err)
		}
		dynamoDBConfig, err := aws.ConfigFromURL(parsed)
		if err != nil {
			return nil, err
		}
		s3Config, err := aws.ConfigFromURL(s3)
		if err != nil {
			return nil, err
		}
		bucketName := strings.TrimPrefix(s3.Path, "/")
		tableName := strings.TrimPrefix(parsed.Path, "/")
		s3Store := multitenant.NewS3Client(s3Config, bucketName)
		var memcacheClient *multitenant.MemcacheClient
		if memcacheConfig.Host != "" {
			memcacheClient = multitenant.NewMemcacheClient(memcacheConfig)
		}
		awsCollector, err := multitenant.NewAWSCollector(
			multitenant.AWSCollectorConfig{
				UserIDer:       userIDer,
				DynamoDBConfig: dynamoDBConfig,
				DynamoTable:    tableName,
				S3Store:        &s3Store,
				NatsHost:       natsHostname,
				MemcacheClient: memcacheClient,
				Window:         window,
				MaxTopNodes:    maxTopNodes,
			},
		)
		if err != nil {
			return nil, err
		}
		if createTables {
			if err := awsCollector.CreateTables(); err != nil {
				return nil, err
			}
		}
		return awsCollector, nil
	}

	return nil, fmt.Errorf("Invalid collector '%s'", collectorURL)
}

func emitterFactory(collector app.Collector, clientCfg billing.Config, userIDer multitenant.UserIDer, emitterCfg multitenant.BillingEmitterConfig) (*multitenant.BillingEmitter, error) {
	billingClient, err := billing.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}
	emitterCfg.UserIDer = userIDer
	return multitenant.NewBillingEmitter(
		collector,
		billingClient,
		emitterCfg,
	)
}

func controlRouterFactory(userIDer multitenant.UserIDer, controlRouterURL string, controlRPCTimeout time.Duration) (app.ControlRouter, error) {
	if controlRouterURL == "local" {
		return app.NewLocalControlRouter(), nil
	}

	parsed, err := url.Parse(controlRouterURL)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme == "sqs" {
		prefix := strings.TrimPrefix(parsed.Path, "/")
		sqsConfig, err := aws.ConfigFromURL(parsed)
		if err != nil {
			return nil, err
		}
		return multitenant.NewSQSControlRouter(sqsConfig, userIDer, prefix, controlRPCTimeout), nil
	}

	return nil, fmt.Errorf("Invalid control router '%s'", controlRouterURL)
}

func pipeRouterFactory(userIDer multitenant.UserIDer, pipeRouterURL, consulInf string) (app.PipeRouter, error) {
	if pipeRouterURL == "local" {
		return app.NewLocalPipeRouter(), nil
	}

	parsed, err := url.Parse(pipeRouterURL)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme == "consul" {
		consulClient, err := multitenant.NewConsulClient(parsed.Host)
		if err != nil {
			return nil, err
		}
		advertise, err := network.GetFirstAddressOf(consulInf)
		if err != nil {
			return nil, err
		}
		addr := fmt.Sprintf("%s:4444", advertise)
		return multitenant.NewConsulPipeRouter(consulClient, strings.TrimPrefix(parsed.Path, "/"), addr, userIDer), nil
	}

	return nil, fmt.Errorf("Invalid pipe router '%s'", pipeRouterURL)
}

// Main runs the app
func appMain(flags appFlags) {
	setLogLevel(flags.logLevel)
	setLogFormatter(flags.logPrefix)
	runtime.SetBlockProfileRate(flags.blockProfileRate)

	registerAppMetricsOnce.Do(registerAppMetrics)

	traceCloser := tracing.NewFromEnv(fmt.Sprintf("scope-%s", flags.serviceName))
	defer traceCloser.Close()

	defer log.Info("app exiting")
	rand.Seed(time.Now().UnixNano())
	app.UniqueID = strconv.FormatInt(rand.Int63(), 16)
	app.Version = version
	log.Infof("app starting, version %s, ID %s", app.Version, app.UniqueID)
	logCensoredArgs()

	userIDer := multitenant.NoopUserIDer
	if flags.userIDHeader != "" {
		userIDer = multitenant.UserIDHeader(flags.userIDHeader)
	}

	collector, err := collectorFactory(
		userIDer, flags.collectorURL, flags.s3URL, flags.natsHostname,
		multitenant.MemcacheConfig{
			Host:             flags.memcachedHostname,
			Timeout:          flags.memcachedTimeout,
			Expiration:       flags.memcachedExpiration,
			UpdateInterval:   memcacheUpdateInterval,
			Service:          flags.memcachedService,
			CompressionLevel: flags.memcachedCompressionLevel,
		},
		flags.window, flags.maxTopNodes, flags.awsCreateTables)
	if err != nil {
		log.Fatalf("Error creating collector: %v", err)
		return
	}

	if flags.BillingEmitterConfig.Enabled {
		billingEmitter, err := emitterFactory(collector, flags.BillingClientConfig, userIDer, flags.BillingEmitterConfig)
		if err != nil {
			log.Fatalf("Error creating emitter: %v", err)
			return
		}
		defer billingEmitter.Close()
		collector = billingEmitter
	}

	controlRouter, err := controlRouterFactory(userIDer, flags.controlRouterURL, flags.controlRPCTimeout)
	if err != nil {
		log.Fatalf("Error creating control router: %v", err)
		return
	}

	pipeRouter, err := pipeRouterFactory(userIDer, flags.pipeRouterURL, flags.consulInf)
	if err != nil {
		log.Fatalf("Error creating pipe router: %v", err)
		return
	}

	// Start background version checking
	checkpoint.CheckInterval(&checkpoint.CheckParams{
		Product: "scope-app",
		Version: app.Version,
		Flags:   makeBaseCheckpointFlags(),
	}, versionCheckPeriod, func(r *checkpoint.CheckResponse, err error) {
		if err != nil {
			log.Errorf("Error checking version: %v", err)
		} else if r.Outdated {
			log.Infof("Scope version %s is available; please update at %s",
				r.CurrentVersion, r.CurrentDownloadURL)
			app.NewVersion(r.CurrentVersion, r.CurrentDownloadURL)
		}
	})

	// Periodically try and register our IP address in WeaveDNS.
	if flags.weaveEnabled && flags.weaveHostname != "" {
		weave, err := newWeavePublisher(
			flags.dockerEndpoint, flags.weaveAddr,
			flags.weaveHostname, flags.containerName)
		if err != nil {
			log.Println("Failed to start weave integration:", err)
		} else {
			defer weave.Stop()
		}
	}

	capabilities := map[string]bool{
		xfer.HistoricReportsCapability: collector.HasHistoricReports(),
	}
	logger := logging.Logrus(log.StandardLogger())
	handler := router(collector, controlRouter, pipeRouter, flags.externalUI, capabilities, flags.metricsGraphURL)
	if flags.logHTTP {
		handler = middleware.Log{
			Log:               logger,
			LogRequestHeaders: flags.logHTTPHeaders,
		}.Wrap(handler)
	}

	if flags.basicAuth {
		log.Infof("Basic authentication enabled")
		handler = httpauth.SimpleBasicAuth(flags.username, flags.password)(handler)
	} else {
		log.Infof("Basic authentication disabled")
	}

	server := &graceful.Server{
		// we want to manage the stop condition ourselves below
		NoSignalHandling: true,
		Server: &http.Server{
			Addr:           flags.listen,
			Handler:        handler,
			ReadTimeout:    httpTimeout,
			WriteTimeout:   httpTimeout,
			MaxHeaderBytes: 1 << 20,
		},
	}
	go func() {
		log.Infof("listening on %s", flags.listen)
		if err := server.ListenAndServe(); err != nil {
			log.Error(err)
		}
	}()

	// block until INT/TERM
	signals.SignalHandlerLoop(
		logger,
		stopper{
			Server:      server,
			StopTimeout: flags.stopTimeout,
		},
	)
}

// stopper adapts graceful.Server's interface to signals.SignalReceiver's interface.
type stopper struct {
	Server      *graceful.Server
	StopTimeout time.Duration
}

// Stop implements signals.SignalReceiver's Stop method.
func (c stopper) Stop() error {
	// stop listening, wait for any active connections to finish
	c.Server.Stop(c.StopTimeout)
	<-c.Server.StopChan()
	return nil
}

func newWeavePublisher(dockerEndpoint, weaveAddr, weaveHostname, containerName string) (*app.WeavePublisher, error) {
	dockerClient, err := docker.NewDockerClientStub(dockerEndpoint)
	if err != nil {
		return nil, err
	}
	weaveClient := weave.NewClient(weaveAddr)
	return app.NewWeavePublisher(
		weaveClient,
		dockerClient,
		app.Interfaces,
		weaveHostname,
		containerName,
	), nil
}
