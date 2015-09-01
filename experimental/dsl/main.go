package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/weaveworks/scope/report"
)

func main() {
	var (
		listen = flag.String("listen", ":8080", "HTTP listen address")
	)
	flag.Parse()

	// Decode report
	var rpt report.Report
	if err := json.NewDecoder(os.Stdin).Decode(&rpt); err != nil {
		log.Fatal(err)
	}

	// Flatten all the topologies
	tpy := report.MakeTopology()
	for _, t := range rpt.Topologies() {
		if err := conflict(tpy, t); err != nil {
			log.Fatal(err)
		}
		tpy = tpy.Merge(t)
	}
	log.Printf("topology with %d node(s)", len(tpy.NodeMetadatas))

	// Serve HTTP
	http.HandleFunc("/json", handleJSON(tpy))
	http.HandleFunc("/dot", handleDOT(tpy))
	http.HandleFunc("/svg", handleSVG(tpy))
	http.HandleFunc("/", handleHTML)
	log.Printf("listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

// Just a safety check.
func conflict(a, b report.Topology) error {
	var errs []string
	for id := range a.NodeMetadatas {
		if _, ok := b.NodeMetadatas[id]; ok {
			errs = append(errs, id)
		}
	}
	for id := range b.NodeMetadatas {
		if _, ok := a.NodeMetadatas[id]; ok {
			errs = append(errs, id)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, ", "))
	}
	return nil
}
