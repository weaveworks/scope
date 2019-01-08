package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/performancecopilot/speed"
)

var metric speed.Counter

func main() {
	var err error
	metric, err = speed.NewPCPCounter(
		0,
		"http.requests",
		"Number of Requests",
	)
	if err != nil {
		log.Fatal("Could not create counter, error: ", err)
	}

	client, err := speed.NewPCPClient("example")
	if err != nil {
		log.Fatal("Could not create client, error: ", err)
	}

	client.MustRegister(metric)

	client.MustStart()
	defer client.MustStop()

	http.HandleFunc("/increment", handleIncrement)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Could not listen on port, error: ", err)
		}
	}()

	fmt.Println("To stop the server press enter")
	_, _ = os.Stdin.Read(make([]byte, 1))
	os.Exit(0)
}

func handleIncrement(w http.ResponseWriter, r *http.Request) {
	metric.Up()
	fmt.Fprintf(w, "incremented\n")
}
