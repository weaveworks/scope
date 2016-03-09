package main

import (
	"flag"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/gorilla/mux"
	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/weave/common"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/app/multitenant"
	"github.com/weaveworks/scope/common/weave"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/docker"
)

// Router creates the mux for all the various app components.
func router(collector app.Collector, controlRouter app.ControlRouter, pipeRouter app.PipeRouter) http.Handler {
	router := mux.NewRouter()
	app.RegisterReportPostHandler(collector, router)
	app.RegisterControlRoutes(router, controlRouter)
	app.RegisterPipeRoutes(router, pipeRouter)
	return app.TopologyHandler(collector, router, http.FileServer(FS(false)))
}

// Main runs the app
func appMain() {
	var (
		window    = flag.Duration("window", 15*time.Second, "window")
		listen    = flag.String("http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
		logLevel  = flag.String("log.level", "info", "logging threshold level: debug|info|warn|error|fatal|panic")
		logPrefix = flag.String("log.prefix", "<app>", "prefix for each log line")

		weaveAddr      = flag.String("weave.addr", app.DefaultWeaveURL, "Address on which to contact WeaveDNS")
		weaveHostname  = flag.String("weave.hostname", app.DefaultHostname, "Hostname to advertise in WeaveDNS")
		containerName  = flag.String("container.name", app.DefaultContainerName, "Name of this container (to lookup container ID)")
		dockerEndpoint = flag.String("docker", app.DefaultDockerEndpoint, "Location of docker endpoint (to lookup container ID)")

		collectorType     = flag.String("collector", "local", "Collector to use (local of dynamodb)")
		controlRouterType = flag.String("control.router", "local", "Control router to use (local or sqs)")
		pipeRouterType    = flag.String("pipe.router", "local", "Pipe router to use (local)")
		userIDHeader      = flag.String("userid.header", "", "HTTP header to use as userid")

		awsDynamoDB     = flag.String("aws.dynamodb", "", "URL of DynamoDB instance")
		awsCreateTables = flag.Bool("aws.create.tables", false, "Create the tables in DynamoDB")
		awsSQS          = flag.String("aws.sqs", "", "URL of SQS instance")
		awsRegion       = flag.String("aws.region", "", "AWS Region")
		awsID           = flag.String("aws.id", "", "AWS Account ID")
		awsSecret       = flag.String("aws.secret", "", "AWS Account Secret")
		awsToken        = flag.String("aws.token", "", "AWS Account Token")

		consulPrefix = flag.String("consul.prefix", "", "Prefix for keys in consul")
		consulAddr   = flag.String("consul.addr", "", "Address of consul instance")
		consulInf    = flag.String("consul.inf", "", "The interface who's address I should advertise myself under in consul")
	)
	flag.Parse()

	setLogLevel(*logLevel)
	setLogFormatter(*logPrefix)

	// Do we need a user IDer?
	var userIDer = multitenant.NoopUserIDer
	if *userIDHeader != "" {
		userIDer = multitenant.UserIDHeader(*userIDHeader)
	}

	// Create a collector
	var collector app.Collector
	if *collectorType == "local" {
		collector = app.NewCollector(*window)
	} else if *collectorType == "dynamodb" {
		creds := credentials.NewStaticCredentials(*awsID, *awsSecret, *awsToken)
		dynamoCollector := multitenant.NewDynamoDBCollector(*awsDynamoDB, *awsRegion, creds, userIDer)
		collector = dynamoCollector
		if *awsCreateTables {
			if err := dynamoCollector.CreateTables(); err != nil {
				log.Fatalf("Error createing DynamoDB tables: %v", err)
			}
		}
	} else {
		log.Fatalf("Invalid collector '%s'", *collectorType)
		return
	}

	// Create a control router
	var controlRouter app.ControlRouter
	if *controlRouterType == "local" {
		controlRouter = app.NewLocalControlRouter()
	} else if *controlRouterType == "sqs" {
		creds := credentials.NewStaticCredentials(*awsID, *awsSecret, *awsToken)
		controlRouter = multitenant.NewSQSControlRouter(*awsSQS, *awsRegion, creds, userIDer)
	} else {
		log.Fatalf("Invalid control router '%s'", *controlRouterType)
		return
	}

	// Create a pipe router
	var pipeRouter app.PipeRouter
	if *pipeRouterType == "local" {
		pipeRouter = app.NewLocalPipeRouter()
	} else if *pipeRouterType == "consul" {
		var err error
		pipeRouter, err = multitenant.NewConsulPipeRouter(*consulAddr, *consulPrefix, *consulInf, userIDer)
		if err != nil {
			log.Fatalf("Error createing consul pipe router: %v", err)
		}
	} else {
		log.Fatalf("Invalid pipe router '%s'", *pipeRouterType)
		return
	}

	defer log.Info("app exiting")
	rand.Seed(time.Now().UnixNano())
	app.UniqueID = strconv.FormatInt(rand.Int63(), 16)
	app.Version = version
	log.Infof("app starting, version %s, ID %s", app.Version, app.UniqueID)

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
