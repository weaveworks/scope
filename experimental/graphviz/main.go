package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"$GITHUB_URI/report"
)

func main() {
	var (
		listen = flag.String("listen", ":8080", "HTTP listen address")
	)
	flag.Parse()

	log.Printf("reading /api/report from stdin...")
	var rpt report.Report
	if err := json.NewDecoder(os.Stdin).Decode(&rpt); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handleHTML(rpt))
	http.HandleFunc("/dot", handleDot(rpt))
	http.HandleFunc("/svg", handleSVG(rpt))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "Stop it", http.StatusTeapot) })
	log.Printf("listening on %s", *listen)
	http.ListenAndServe(*listen, nil)
}
