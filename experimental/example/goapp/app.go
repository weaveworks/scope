package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	var (
		addr = flag.String("addr", ":8080", "HTTP listen address")
		rate = flag.Duration("rate", 3*time.Second, "request rate")
	)
	flag.Parse()

	var targets []string
	for _, s := range os.Args[1:] {
		target, err := normalize(s)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("target %s", target)
		targets = append(targets, target)
	}
	if len(targets) <= 0 {
		log.Fatal("no targets")
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = strings.ToLower(strings.Replace(hostname, " ", "-", -1)) // lol
	}

	var reads uint64

	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "%s %d", hostname, atomic.LoadUint64(&reads))
	})

	errc := make(chan error)

	go func() {
		log.Printf("listening on %s", *addr)
		errc <- http.ListenAndServe(*addr, nil)
	}()

	go func() {
		errc <- read(targets, *rate, &reads)
	}()

	go func() {
		errc <- interrupt()
	}()

	log.Printf("%s", <-errc)
}

func read(targets []string, rate time.Duration, reads *uint64) error {
	for range time.Tick(rate) {
		var wg sync.WaitGroup
		wg.Add(len(targets))
		for _, target := range targets {
			go func(target string) {
				get(target)
				atomic.AddUint64(reads, 1)
				wg.Done()
			}(target)
		}
		wg.Wait()
	}
	return nil
}

func get(target string) {
	resp, err := http.Get(target)
	if err != nil {
		log.Printf("%s: %v", target, err)
		return
	}
	log.Printf("%s: %s", target, resp.Status)
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}

func normalize(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" && u.Path != "" {
		u.Host, u.Path = u.Path, u.Host
	}
	if _, port, err := net.SplitHostPort(u.Host); err != nil || port == "" {
		u.Host = u.Host + ":9200"
	}
	return u.String(), nil
}
