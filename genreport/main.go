package main

import (
	"encoding/json"
	"flag"
	"os"
)

func main() {
	nodes := flag.Int("nodes", 10, "node count")
	flag.Parse()
	json.NewEncoder(os.Stdout).Encode(DemoReport(*nodes))
}
