package main

import (
	"log"
	"os"

	"github.com/gorilla/websocket"
	//"golang.org/x/crypto/ssh/terminal"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("must specify url")
	}
	url := os.Args[1]
	log.Printf("Connecting to %s", url)

	//oldState, err := terminal.MakeRaw(0)
	//if err != nil {
	//	panic(err)
	//}
	//defer terminal.Restore(0, oldState)

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	readQuit := make(chan struct{})
	writeQuit := make(chan struct{})

	// Read-from-UI loop
	go func() {
		defer close(readQuit)
		for {
			_, buf, err := conn.ReadMessage() // TODO type should be binary message
			if err != nil {
				log.Printf("Error reading websocket: %v", err)
				return
			}

			spew.Dump(buf)

			if _, err := os.Stdout.Write(buf); err != nil {
				log.Printf("Error writing stdout: %v", err)
				return
			}
		}
	}()

	// Write-to-UI loop
	go func() {
		defer close(writeQuit)
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				log.Printf("Error reading stdin: %v", err)
				return
			}

			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				log.Printf("Error writing websocket: %v", err)
				return
			}
		}
	}()

	// block until one of the goroutines exits
	// this convoluted mechanism is to ensure we only close the websocket once.
	select {
	case <-readQuit:
	case <-writeQuit:
	}
}
