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

	time.Sleep(100 * time.Second)

	fmt.Println("done")
}
