package app

import (
	"net/http"
)

// Raw report handler
func makeRawReportHandler(rep Reporter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// r.ParseForm()
		respondWith(w, http.StatusOK, rep.Report())
	}
}
