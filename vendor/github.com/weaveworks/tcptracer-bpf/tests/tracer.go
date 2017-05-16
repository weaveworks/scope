package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/weaveworks/tcptracer-bpf/pkg/tracer"
)

var lastTimestampV4 uint64
var lastTimestampV6 uint64

func tcpEventCbV4(e tracer.TcpV4) {
	fmt.Printf("%v cpu#%d %s %v %s %v:%v %v:%v %v\n",
		e.Timestamp, e.CPU, e.Type, e.Pid, e.Comm, e.SAddr, e.SPort, e.DAddr, e.DPort, e.NetNS)

	if lastTimestampV4 > e.Timestamp {
		fmt.Printf("ERROR: late event!\n")
		os.Exit(1)
	}

	lastTimestampV4 = e.Timestamp
}

func tcpEventCbV6(e tracer.TcpV6) {
	fmt.Printf("%v cpu#%d %s %v %s %v:%v %v:%v %v\n",
		e.Timestamp, e.CPU, e.Type, e.Pid, e.Comm, e.SAddr, e.SPort, e.DAddr, e.DPort, e.NetNS)

	if lastTimestampV6 > e.Timestamp {
		fmt.Printf("ERROR: late event!\n")
		os.Exit(1)
	}

	lastTimestampV6 = e.Timestamp
}

func lostCb(count uint64) {
	fmt.Printf("ERROR: lost %d events!\n", count)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
		os.Exit(1)
	}

	t, err := tracer.NewTracer(tcpEventCbV4, tcpEventCbV6, lostCb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Ready\n")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	<-sig
	t.Stop()
}
