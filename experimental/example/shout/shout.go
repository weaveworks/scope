package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	SHOUTCLOUD "github.com/richo/GOSHOUT"
)

type requestResponse struct {
	Text string `json:"text"`
}

func main() {
	var (
		addr = flag.String("addr", ":8090", "HTTP listen address")
	)
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var req requestResponse
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusTeapot)
			return
		}

		req.Text, err = SHOUTCLOUD.UPCASE(req.Text)
		if err != nil {
			w.WriteHeader(http.StatusTeapot)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&req)
	})

	errc := make(chan error)

	go func() {
		log.Printf("listening on %s", *addr)
		errc <- http.ListenAndServe(*addr, nil)
	}()

	go func() {
		errc <- interrupt()
	}()

	log.Printf("%v", <-errc)
}

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}
