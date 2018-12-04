package main

// Example usage

import (
	"fmt"
	"time"

	"github.com/typetypetype/conntrack"
)

func main() {
	cs, err := conntrack.Established()
	if err != nil {
		panic(fmt.Sprintf("Established: %s", err))
	}
	fmt.Printf("Established on start:\n")
	for _, cn := range cs {
		fmt.Printf(" - %s\n", cn)
	}
	fmt.Println("")

	c, err := conntrack.New()
	if err != nil {
		panic(err)
	}
	for range time.Tick(1 * time.Second) {
		fmt.Printf("Connections:\n")
		for _, cn := range c.Connections() {
			fmt.Printf(" - %s\n", cn)
		}
	}
}
