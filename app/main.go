package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/weaveworks/weave/common"

	"github.com/weaveworks/scope/xfer"
)

var (
	// Set at buildtime.
	version = "dev"

	// Set at runtime.
	uniqueID = "0"
)

func registerStatic(router *mux.Router) {
	router.Methods("GET").PathPrefix("/").Handler(http.FileServer(FS(false)))
}

// Router creates the mux for all the various app components.
func Router(c collector) *mux.Router {
	router := mux.NewRouter()
	registerTopologyRoutes(c, router)
	registerControlRoutes(router)
	registerStatic(router)
	return router
}

func main() {
	var (
		window       = flag.Duration("window", 15*time.Second, "window")
		listen       = flag.String("http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
		logPrefix    = flag.String("log.prefix", "<app>", "prefix for each log line")
		printVersion = flag.Bool("version", false, "print version number and exit")
	)
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	if !strings.HasSuffix(*logPrefix, " ") {
		*logPrefix += " "
	}
	log.SetPrefix(*logPrefix)

	defer log.Print("app exiting")

	rand.Seed(time.Now().UnixNano())
	uniqueID = strconv.FormatInt(rand.Int63(), 16)
	log.Printf("app starting, version %s, ID %s", version, uniqueID)

	c := NewCollector(*window)
	http.Handle("/", Router(c))
	go func() {
		log.Printf("listening on %s", *listen)
		log.Print(http.ListenAndServe(*listen, nil))
	}()

	common.SignalHandlerLoop()
}
