package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
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
		logPrefix = flag.String("log.prefix", "<app>", "prefix for each log line")
	)
	flag.Parse()

	if !strings.HasSuffix(*logPrefix, " ") {
		*logPrefix += " "
	}
	log.SetPrefix(*logPrefix)

	defer log.Print("app exiting")

	rand.Seed(time.Now().UnixNano())
	app.UniqueID = strconv.FormatInt(rand.Int63(), 16)
	app.Version = version
	log.Printf("app starting, version %s, ID %s", app.Version, app.UniqueID)
	handler := router(app.NewCollector(*window))
	go func() {
		log.Printf("listening on %s", *listen)
		log.Print(http.ListenAndServe(*listen, handler))
	}()

	common.SignalHandlerLoop()
}
