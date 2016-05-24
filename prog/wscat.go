package main

import (
	"flag"
	"io"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

func wscat(flags wscatFlags) {
	setLogFormatter("wscat")
	// Output to stderr instead of stdout
	log.SetOutput(os.Stderr)
	dialer := websocket.Dialer{}
	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("Only one url argument expected")
	}
	url := args[0]
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("Cannot dial %s: %s", url, err)
	}
	defer conn.Close()
	readQuit := make(chan int)
	writeQuit := make(chan int)

	// Read-from-UI loop
	go func() {
		status := 0
		for {
			_, buf, err := conn.ReadMessage() // TODO type should be binary message
			if err != nil {
				if err != io.EOF && err != io.ErrUnexpectedEOF {
					status = 1
					log.Errorf("Error reading websocket: %s", err)
				}
				break
			}

			if _, err := os.Stdout.Write(buf); err != nil {
				status = 1
				log.Errorf("Error writing to stdout: %s", err)
				break
			}
		}
		readQuit <- status
	}()

	// Write-to-UI loop
	go func() {
		status := 0
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err != io.EOF && err != io.ErrUnexpectedEOF {
					log.Errorf("Error reading stdin: %s", err)
				}
				status = 1
				break
			}

			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Errorf("Error writing websocket: %s", err)
				status = 1
				break
			}
		}
		writeQuit <- status
	}()

	// block until one (both when blockOnEOF) of the goroutines exit
	// this convoluted mechanism is to ensure we only close the websocket once.
	var (
		readStatus  = -1
		writeStatus = -1
	)
	for {
		select {
		case readStatus = <-readQuit:
		case writeStatus = <-writeQuit:
		}
		if !flags.blockOnEOF || (readStatus != -1 && writeStatus != -1) {
			break
		}
	}
	if readStatus > 0 || writeStatus > 0 {
		os.Exit(1)
	}

}
