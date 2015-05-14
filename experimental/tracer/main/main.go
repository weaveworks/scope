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
	go tracer.http(6060)
	handleSignals()
}

func handleSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT)
	buf := make([]byte, 1<<20)
	for {
		sig := <-sigs
		switch sig {
		case syscall.SIGQUIT:
			stacklen := runtime.Stack(buf, true)
			log.Printf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
		}
	}
}
