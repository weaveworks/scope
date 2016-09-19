package main

import (
	"fmt"
	"os"
	"time"

	"github.com/weaveworks/scope/probe/endpoint"
)

func main() {
	tr := endpoint.NewEbpfTracker("/home/asymmetric/code/kinvolk/bcc/examples/tracing/tcpv4tracer.py")

	if tr == nil {
		fmt.Fprintf(os.Stderr, "error creating tracker\n")
		os.Exit(1)
	}

	// create some http connection within these 10 seconds
	time.Sleep(10 * time.Second)

	tr.WalkEvents(func(e endpoint.ConnectionEvent) {
		fmt.Println(e)
	})

	fmt.Println("done")
}
