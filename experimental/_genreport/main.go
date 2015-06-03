package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

func main() {
	nodes := flag.Int("nodes", 10, "node count")
	flag.Parse()
	if err := json.NewEncoder(os.Stdout).Encode(DemoReport(*nodes)); err != nil {
		log.Print(err)
	}
}
