package main

import (
	"fmt"
	"os"

	"github.com/weaveworks/scope/prog/app"
	"github.com/weaveworks/scope/prog/probe"
)

func usage() {
	fmt.Printf("usage: %s (app|probe) args...\n", os.Args[0])
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
		app.Main()
	case "probe":
		probe.Main()
	default:
		usage()
	}
}
