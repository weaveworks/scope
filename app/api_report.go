package app

import (
	"net/http"

	"golang.org/x/net/context"
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
