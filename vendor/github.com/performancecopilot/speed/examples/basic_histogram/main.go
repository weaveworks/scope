package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/performancecopilot/speed"
)

func main() {
	max := int64(100)

	c, err := speed.NewPCPClient("histogram_test")
	if err != nil {
		log.Fatal("Could not create client, error: ", err)
	}

	m, err := speed.NewPCPHistogram("hist", 0, max, 5, speed.OneUnit, "a sample histogram")
	if err != nil {
		log.Fatal("Could not create histogram, error: ", err)
	}

	c.MustRegister(m)

	c.MustStart()
	defer c.MustStop()

	for i := 0; i < 60; i++ {
		v := rand.Int63n(max)

		fmt.Println("recording", v)
		m.MustRecord(v)

		time.Sleep(time.Second)
	}
}
