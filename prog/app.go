package main

import (
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/weave/common"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/app/multitenant"
	"github.com/weaveworks/scope/common/middleware"
	"github.com/weaveworks/scope/common/network"
	"github.com/weaveworks/scope/common/weave"
	"github.com/weaveworks/scope/probe/docker"
)

var (
	requestDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "scope",
		Name:      "request_duration_nanoseconds",
		Help:      "Time spent serving HTTP requests.",
	}, []string{"method", "route", "status_code", "ws"})
)

func init() {
	prometheus.MustRegister(requestDuration)
}

// Router creates the mux for all the various app components.
func router(collector app.Collector, controlRouter app.ControlRouter, pipeRouter app.PipeRouter) http.Handler {
	router := mux.NewRouter().SkipClean(true)

	// We pull in the http.DefaultServeMux to get the pprof routes
	router.PathPrefix("/debug/pprof").Handler(http.DefaultServeMux)
	router.Path("/metrics").Handler(prometheus.Handler())

	app.RegisterReportPostHandler(collector, router)
	app.RegisterControlRoutes(router, controlRouter)
	app.RegisterPipeRoutes(router, pipeRouter)
	app.RegisterTopologyRoutes(router, collector)

	router.PathPrefix("/").Name("static").Handler(http.FileServer(FS(false)))

	instrument := middleware.Instrument{
		RouteMatcher: router,
		Duration:     requestDuration,
	}
	return instrument.Wrap(router)
}

func awsConfigFromURL(url *url.URL) (*aws.Config, error) {
	if url.User == nil {
		return nil, fmt.Errorf("Must specify username & password in URL")
	}
	password, _ := url.User.Password()
	creds := credentials.NewStaticCredentials(url.User.Username(), password, "")
	config := aws.NewConfig().WithCredentials(creds)
	if strings.Contains(url.Host, ".") {
		config = config.WithEndpoint(fmt.Sprintf("http://%s", url.Host)).WithRegion("dummy")
	} else {
		config = config.WithRegion(url.Host)
	}
	return config, nil
}

func collectorFactory(userIDer multitenant.UserIDer, collectorURL, s3URL, natsHostname, memcachedHostname string, memcachedTimeout time.Duration, memcachedService string, window time.Duration, createTables bool) (app.Collector, error) {
	if collectorURL == "local" {
		return app.NewCollector(window), nil
	}

	parsed, err := url.Parse(collectorURL)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme == "dynamodb" {
		s3, err := url.Parse(s3URL)
		if err != nil {
			return nil, fmt.Errorf("Valid URL for s3 required: %v", err)
		}
		dynamoDBConfig, err := awsConfigFromURL(parsed)
		if err != nil {
			return nil, err
		}
		s3Config, err := awsConfigFromURL(s3)
		if err != nil {
			return nil, err
		}
		tableName := strings.TrimPrefix(parsed.Path, "/")
		bucketName := strings.TrimPrefix(s3.Path, "/")
		dynamoCollector, err := multitenant.NewDynamoDBCollector(
			userIDer, dynamoDBConfig, s3Config, tableName, bucketName, natsHostname,
			memcachedHostname, memcachedTimeout, memcachedService,
		)
		if err != nil {
			return nil, err
		}
		if createTables {
			if err := dynamoCollector.CreateTables(); err != nil {
				return nil, err
			}
		}
		return dynamoCollector, nil
	}

	return nil, fmt.Errorf("Invalid collector '%s'", collectorURL)
}

func controlRouterFactory(userIDer multitenant.UserIDer, controlRouterURL string) (app.ControlRouter, error) {
	if controlRouterURL == "local" {
		return app.NewLocalControlRouter(), nil
	}

	parsed, err := url.Parse(controlRouterURL)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme == "sqs" {
		prefix := strings.TrimPrefix(parsed.Path, "/")
		sqsConfig, err := awsConfigFromURL(parsed)
		if err != nil {
			return nil, err
		}
		return multitenant.NewSQSControlRouter(sqsConfig, userIDer, prefix), nil
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

	defer log.Info("app exiting")
	rand.Seed(time.Now().UnixNano())
	app.UniqueID = strconv.FormatInt(rand.Int63(), 16)
	app.Version = version
	log.Infof("app starting, version %s, ID %s", app.Version, app.UniqueID)
	log.Infof("command line: %v", os.Args)

	userIDer := multitenant.NoopUserIDer
	if flags.userIDHeader != "" {
		userIDer = multitenant.UserIDHeader(flags.userIDHeader)
	}

	collector, err := collectorFactory(
		userIDer, flags.collectorURL, flags.s3URL, flags.natsHostname, flags.memcachedHostname,
		flags.memcachedTimeout, flags.memcachedService, flags.window, flags.awsCreateTables)
	if err != nil {
		log.Fatalf("Error creating collector: %v", err)
		return
	}

	controlRouter, err := controlRouterFactory(userIDer, flags.controlRouterURL)
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
	}, versionCheckPeriod, func(r *checkpoint.CheckResponse, err error) {
		if err != nil {
			log.Errorf("Error checking version: %v", err)
		} else if r.Outdated {
			log.Infof("Scope version %s is available; please update at %s",
				r.CurrentVersion, r.CurrentDownloadURL)
			app.NewVersion(r.CurrentVersion, r.CurrentDownloadURL)
		}
	})

	// Periodically try and register out IP address in WeaveDNS.
	if flags.weaveEnabled {
		weave, err := newWeavePublisher(
			flags.dockerEndpoint, flags.weaveAddr,
			flags.weaveHostname, flags.containerName)
		if err != nil {
			log.Println("Failed to start weave integration:", err)
		} else {
			defer weave.Stop()
		}
	}

	handler := router(collector, controlRouter, pipeRouter)
	if flags.logHTTP {
		handler = middleware.Logging.Wrap(handler)
	}
	go func() {
		log.Infof("listening on %s", flags.listen)
		log.Info(http.ListenAndServe(flags.listen, handler))
	}()

	common.SignalHandlerLoop()
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
