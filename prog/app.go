package main

import (
	"flag"
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
	"github.com/weaveworks/scope/common/weave"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/docker"
)

var (
	requestDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "scope",
		Name:      "request_duration_nanoseconds",
		Help:      "Time spent serving HTTP requests.",
	}, []string{"method", "route", "status_code"})
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

	router.PathPrefix("/").Handler(http.FileServer(FS(false)))

	instrument := middleware.Instrument{
		RouteMatcher: router,
		Duration:     requestDuration,
	}
	return instrument.Wrap(router)
}

func awsConfigFromURL(url *url.URL) *aws.Config {
	password, _ := url.User.Password()
	creds := credentials.NewStaticCredentials(url.User.Username(), password, "")
	config := aws.NewConfig().WithCredentials(creds)
	if strings.Contains(url.Host, ".") {
		config = config.WithEndpoint(fmt.Sprintf("http://%s", url.Host)).WithRegion("dummy")
	} else {
		config = config.WithRegion(url.Host)
	}
	return config
}

func collectorFactory(userIDer multitenant.UserIDer, collectorURL string, window time.Duration, createTables bool) (app.Collector, error) {
	if collectorURL == "local" {
		return app.NewCollector(window), nil
	}

	parsed, err := url.Parse(collectorURL)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme == "dynamodb" {
		dynamoCollector := multitenant.NewDynamoDBCollector(awsConfigFromURL(parsed), userIDer)
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
		return multitenant.NewSQSControlRouter(awsConfigFromURL(parsed), userIDer), nil
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
		return multitenant.NewConsulPipeRouter(consulClient, strings.TrimPrefix(parsed.Path, "/"), consulInf, userIDer)
	}

	return nil, fmt.Errorf("Invalid pipe router '%s'", pipeRouterURL)
}

// Main runs the app
func appMain() {
	var (
		window    = flag.Duration("window", 15*time.Second, "window")
		listen    = flag.String("http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
		logLevel  = flag.String("log.level", "info", "logging threshold level: debug|info|warn|error|fatal|panic")
		logPrefix = flag.String("log.prefix", "<app>", "prefix for each log line")
		logHTTP   = flag.Bool("log.http", false, "Log individual HTTP requests")

		weaveAddr      = flag.String("weave.addr", app.DefaultWeaveURL, "Address on which to contact WeaveDNS")
		weaveHostname  = flag.String("weave.hostname", app.DefaultHostname, "Hostname to advertise in WeaveDNS")
		containerName  = flag.String("container.name", app.DefaultContainerName, "Name of this container (to lookup container ID)")
		dockerEndpoint = flag.String("docker", app.DefaultDockerEndpoint, "Location of docker endpoint (to lookup container ID)")

		collectorURL     = flag.String("collector", "local", "Collector to use (local of dynamodb)")
		controlRouterURL = flag.String("control.router", "local", "Control router to use (local or sqs)")
		pipeRouterURL    = flag.String("pipe.router", "local", "Pipe router to use (local)")
		userIDHeader     = flag.String("userid.header", "", "HTTP header to use as userid")

		awsCreateTables = flag.Bool("aws.create.tables", false, "Create the tables in DynamoDB")
		consulInf       = flag.String("consul.inf", "", "The interface who's address I should advertise myself under in consul")
	)
	flag.Parse()

	setLogLevel(*logLevel)
	setLogFormatter(*logPrefix)

	userIDer := multitenant.NoopUserIDer
	if *userIDHeader != "" {
		userIDer = multitenant.UserIDHeader(*userIDHeader)
	}

	collector, err := collectorFactory(userIDer, *collectorURL, *window, *awsCreateTables)
	if err != nil {
		log.Fatalf("Error creating collector: %v", err)
		return
	}

	controlRouter, err := controlRouterFactory(userIDer, *controlRouterURL)
	if err != nil {
		log.Fatalf("Error creating control router: %v", err)
		return
	}

	pipeRouter, err := pipeRouterFactory(userIDer, *pipeRouterURL, *consulInf)
	if err != nil {
		log.Fatalf("Error creating pipe router: %v", err)
		return
	}

	defer log.Info("app exiting")
	rand.Seed(time.Now().UnixNano())
	app.UniqueID = strconv.FormatInt(rand.Int63(), 16)
	app.Version = version
	log.Infof("app starting, version %s, ID %s", app.Version, app.UniqueID)
	log.Infof("command line: %v", os.Args)

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
		}
	})

	// If user supplied a weave router address, periodically try and register
	// out IP address in WeaveDNS.
	if *weaveAddr != "" {
		weave, err := newWeavePublisher(
			*dockerEndpoint, *weaveAddr,
			*weaveHostname, *containerName)
		if err != nil {
			log.Println("Failed to start weave integration:", err)
		} else {
			defer weave.Stop()
		}
	}

	handler := router(collector, controlRouter, pipeRouter)
	if *logHTTP {
		handler = middleware.Logging.Wrap(handler)
	}
	go func() {
		log.Infof("listening on %s", *listen)
		log.Info(http.ListenAndServe(*listen, handler))
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
