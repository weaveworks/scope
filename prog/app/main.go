package app

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	// a blank import should be only in a main or test package, or have a comment justifying it
	_ "net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/weaveworks/weave/common"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/xfer"
)

// Router creates the mux for all the various app components.
func Router(c app.Collector) *mux.Router {
	router := mux.NewRouter()
	app.RegisterTopologyRoutes(c, router)
	app.RegisterReportPostHandler(c, router)
	app.RegisterControlRoutes(router)
	router.Methods("GET").PathPrefix("/").Handler(http.FileServer(FS(false)))
	return router
}

// Main runs the app
func Main() {
	var (
		window       = flag.Duration("window", 15*time.Second, "window")
		listen       = flag.String("http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
		logPrefix    = flag.String("log.prefix", "<app>", "prefix for each log line")
		printVersion = flag.Bool("version", false, "print version number and exit")
	)
	flag.Parse()

	if *printVersion {
		fmt.Println(app.Version)
		return
	}

	if !strings.HasSuffix(*logPrefix, " ") {
		*logPrefix += " "
	}
	log.SetPrefix(*logPrefix)

	defer log.Print("app exiting")

	rand.Seed(time.Now().UnixNano())
	app.UniqueID = strconv.FormatInt(rand.Int63(), 16)
	log.Printf("app starting, version %s, ID %s", app.Version, app.UniqueID)
	http.Handle("/", Router(app.NewCollector(*window)))
	go func() {
		log.Printf("listening on %s", *listen)
		log.Print(http.ListenAndServe(*listen, nil))
	}()

	common.SignalHandlerLoop()
}
