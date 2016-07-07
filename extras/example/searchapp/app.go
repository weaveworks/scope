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
)

func main() {
	var (
		addr   = flag.String("addr", ":80", "HTTP listen address")
		target = flag.String("target", "http://elasticsearch.weave.local:9200", "Elastic hostname")
	)
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = strings.ToLower(strings.Replace(hostname, " ", "-", -1)) // lol
	}

	var reads uint64

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling %v", r)
		read([]string{*target}, &reads)
		fmt.Fprintf(w, "%s %d", hostname, atomic.LoadUint64(&reads))
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

func read(targets []string, reads *uint64) {
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

func get(target string) {
	resp, err := http.Get(target)
	if err != nil {
		log.Printf("%s: %v", target, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("%s: %s", target, resp.Status)
	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		log.Printf("%s: %v", target, err)
		return
	}
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
