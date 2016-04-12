package app

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/weaveworks/scope/report"
)

// Raw report handler
func makeRawReportHandler(rep Reporter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		report, err := rep.Report(ctx)
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWith(w, http.StatusOK, report)
	}
}

// Probe handler
func makeProbeHandler(rep Reporter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		rpt, err := rep.Report(ctx)
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err.Error())
			return
		}
		result := []report.Probe{}
		for _, p := range rpt.Probes {
			result = append(result, p)
		}
		respondWith(w, http.StatusOK, result)
	}
}
