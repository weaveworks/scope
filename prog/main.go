package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s (app|probe) args...\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	module := os.Args[1]
	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)

	switch module {
	case "app":
		appMain()
	case "probe":
		probeMain()
	default:
		usage()
	}
}
