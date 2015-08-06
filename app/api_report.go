package main

import (
	"net/http"

	"github.com/weaveworks/scope/xfer"
)

// Raw report handler
func makeRawReportHandler(rep xfer.Reporter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// r.ParseForm()
		respondWith(w, http.StatusOK, rep.Report())
	}
}
