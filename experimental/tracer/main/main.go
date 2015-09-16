package main

import (
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/weaveworks/scope/experimental/tracer/ptrace"
	"github.com/weaveworks/scope/probe/docker"
)

const (
	procRoot     = "/proc"
	pollInterval = 10 * time.Second
)

type tracer struct {
	ptrace ptrace.PTracer
	store  *store
	docker docker.Registry
}

func (t *tracer) Stop() {
	log.Printf("Shutting down...")
	t.ptrace.Stop()
	t.docker.Stop()
	log.Printf("Done.")
}

func main() {
	dockerRegistry, err := docker.NewRegistry(pollInterval)
	if err != nil {
		log.Fatalf("Could start docker watcher: %v", err)
	}

	tracer := tracer{
		ptrace: ptrace.NewPTracer(),
		store:  newStore(),
		docker: dockerRegistry,
	}
	defer tracer.Stop()

	go tracer.http(6060)
	<-handleSignals()
}

func handleSignals() chan struct{} {
	quit := make(chan struct{}, 10)
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		buf := make([]byte, 1<<20)
		for {
			sig := <-sigs
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				quit <- struct{}{}
			case syscall.SIGQUIT:
				stacklen := runtime.Stack(buf, true)
				log.Printf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
			}
		}
	} ()
	return quit
}
