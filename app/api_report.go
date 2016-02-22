package app

import (
	"net/http"

	"golang.org/x/net/context"
)

// Raw report handler
func makeRawReportHandler(rep Reporter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		respondWith(w, http.StatusOK, rep.Report(ctx))
	}
}
