package app

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/common/xfer"
)

// RegisterPipeRoutes registers the pipe routes
func RegisterPipeRoutes(router *mux.Router, pr PipeRouter) {
	router.Methods("GET").
		Path("/api/pipe/{pipeID}/check").
		HandlerFunc(requestContextDecorator(checkPipe(pr, UIEnd)))

	router.Methods("GET").
		Path("/api/pipe/{pipeID}").
		HandlerFunc(requestContextDecorator(handlePipeWs(pr, UIEnd)))

	router.Methods("GET").
		Path("/api/pipe/{pipeID}/probe").
		HandlerFunc(requestContextDecorator(handlePipeWs(pr, ProbeEnd)))

	router.Methods("DELETE", "POST").
		Path("/api/pipe/{pipeID}").
		HandlerFunc(requestContextDecorator(deletePipe(pr)))
}

func checkPipe(pr PipeRouter, end End) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["pipeID"]
		exists, err := pr.Exists(ctx, id)
		if err != nil {
			respondWith(w, http.StatusInternalServerError, err)
		} else if exists {
			w.WriteHeader(http.StatusNoContent)
		} else {
			http.NotFound(w, r)
		}
	}
}

func handlePipeWs(pr PipeRouter, end End) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["pipeID"]
		pipe, endIO, err := pr.Get(ctx, id, end)
		if err != nil {
			log.Errorf("Error getting pipe %s: %v", id, err)
			http.NotFound(w, r)
			return
		}
		defer pr.Release(ctx, id, end)

		conn, err := xfer.Upgrade(w, r, nil)
		if err != nil {
			log.Errorf("Error upgrading pipe %s (%d) websocket: %v", id, end, err)
			return
		}
		defer conn.Close()

		log.Infof("Success got pipe %s:%s", id, end)
		if err := pipe.CopyToWebsocket(endIO, conn); err != nil && !xfer.IsExpectedWSCloseError(err) {
			log.Printf("Error copying to pipe %s (%d) websocket: %v", id, end, err)
		}
	}
}

func deletePipe(pr PipeRouter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		pipeID := mux.Vars(r)["pipeID"]
		log.Infof("Deleting pipe %s", pipeID)
		if err := pr.Delete(ctx, pipeID); err != nil {
			respondWith(w, http.StatusInternalServerError, err)
		}
	}
}
