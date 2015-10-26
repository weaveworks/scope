package main

import (
	"fmt"

	"github.com/weaveworks/procspy"
)

func main() {
	cs, err := procspy.Connections(true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("TCP Connections:\n")
	for c := cs.Next(); c != nil; c = cs.Next() {
		fmt.Printf(" - %+v\n", c)
	}
}
