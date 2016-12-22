package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func connect(url string, numConn int) {
	fmt.Printf("Establishing %d TCP connections to %s\n", numConn, url)
	for x := 0; x < numConn; x++ {
		_, err := net.Dial("tcp", url)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	// wait forever
	select {}
}

func listen(url string) {
	l, err := net.Listen("tcp", url)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening on " + url)
	for {
		_, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
	}

}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments")
		os.Exit(1)
	}

	verb := os.Args[1]

	if verb == "connect" {
		if len(os.Args) != 4 {
			fmt.Println("Not enough arguments")
			os.Exit(1)
		}

		url := os.Args[2]
		numConn, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("Error with second argument\n")
			os.Exit(1)
		}

		connect(url, numConn)
	}
	if verb == "listen" {
		if len(os.Args) != 3 {
			fmt.Println("Not enough arguments")
			os.Exit(1)
		}

		port := os.Args[2]
		listen(":" + port)
	}

}
