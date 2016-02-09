package main

import (
	"flag"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/weaveworks/go-checkpoint"
	"github.com/weaveworks/weave/common"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/common/weave"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/docker"
)

// Router creates the mux for all the various app components.
func router(c app.Collector) http.Handler {
	router := mux.NewRouter()
	app.RegisterReportPostHandler(c, router)
	app.RegisterControlRoutes(router)
	app.RegisterPipeRoutes(router)
	return app.TopologyHandler(c, router, http.FileServer(FS(false)))
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
	)
	flag.Parse()

	setLogLevel(*logLevel)
	setLogFormatter(*logPrefix)

	defer log.Info("app exiting")

	// Start background version checking
	checkpoint.CheckInterval(&checkpoint.CheckParams{
		Product:       "scope-app",
		Version:       app.Version,
		SignatureFile: signatureFile,
	}, versionCheckPeriod, func(r *checkpoint.CheckResponse, err error) {
		if r.Outdated {
			log.Infof("Scope version %s is available; please update at %s",
				r.CurrentVersion, r.CurrentDownloadURL)
		}
	})

	rand.Seed(time.Now().UnixNano())
	app.UniqueID = strconv.FormatInt(rand.Int63(), 16)
	app.Version = version
	log.Infof("app starting, version %s, ID %s", app.Version, app.UniqueID)

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

	handler := router(app.NewCollector(*window))
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
