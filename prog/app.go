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
	"github.com/weaveworks/scope/common/xfer"
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
	handler := router(app.NewCollector(*window))
	go func() {
		log.Infof("listening on %s", *listen)
		log.Info(http.ListenAndServe(*listen, handler))
	}()

	common.SignalHandlerLoop()
}
