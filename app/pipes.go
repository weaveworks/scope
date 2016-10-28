package app

import (
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/common/xfer"
)

// RegisterPipeRoutes registers the pipe routes
func RegisterPipeRoutes(router *mux.Router, pr PipeRouter) {
	router.Methods("GET").
		Name("api_pipe_pipeid_check").
		Path("/api/pipe/{pipeID}/check").
		HandlerFunc(requestContextDecorator(checkPipe(pr)))

	router.Methods("POST").
		Name("api_pipe_pipeid_resize_tty").
		Path("/api/pipe/{pipeID}/resize_tty/{height}/{width}").
		HandlerFunc(requestContextDecorator(resizeTTY(pr)))

	router.Methods("GET").
		Name("api_pipe_pipeid").
		Path("/api/pipe/{pipeID}").
		HandlerFunc(requestContextDecorator(handlePipeWs(pr, UIEnd)))

	router.Methods("GET").
		Name("api_pipe_pipeid_probe").
		Path("/api/pipe/{pipeID}/probe").
		HandlerFunc(requestContextDecorator(handlePipeWs(pr, ProbeEnd)))

	router.Methods("DELETE", "POST").
		Name("api_pipe_pipeid").
		Path("/api/pipe/{pipeID}").
		HandlerFunc(requestContextDecorator(deletePipe(pr)))
}

func checkPipe(pr PipeRouter) CtxHandlerFunc {
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

func resizeTTY(pr PipeRouter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var (
			height, width uint64
			err           error
		)
		id := mux.Vars(r)["pipeID"]
		height, err = strconv.ParseUint(mux.Vars(r)["height"], 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		width, err = strconv.ParseUint(mux.Vars(r)["width"], 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		pipe, _, err := pr.Get(ctx, id, ProbeEnd)
		if err != nil {
			log.Debugf("Error getting pipe %s: %v", id, err)
			http.NotFound(w, r)
			return
		}

		tty := pipe.GetTTY()
		if tty == nil {
			// The pipe doesn't have a tty, nothing to do
			log.Debugf("Error getting terminal for pipe %s", id)
			http.NotFound(w, r)
			return
		}

		if err = tty.SetSize(uint(height), uint(width)); err != nil {
			log.Errorf("Error setting terminal size (%d, %d) of pipe %s: %v", height, width, id, err)
			respondWith(w, http.StatusInternalServerError, err)
		}

	}
}

func handlePipeWs(pr PipeRouter, end End) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["pipeID"]
		pipe, endIO, err := pr.Get(ctx, id, end)
		if err != nil {
			// this usually means the pipe has been closed
			log.Debugf("Error getting pipe %s: %v", id, err)
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

		if err := pipe.CopyToWebsocket(endIO, conn); err != nil && !xfer.IsExpectedWSCloseError(err) {
			log.Errorf("Error copying to pipe %s (%d) websocket: %v", id, end, err)
		}
	}
}

func deletePipe(pr PipeRouter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		pipeID := mux.Vars(r)["pipeID"]
		log.Debugf("Deleting pipe %s", pipeID)
		if err := pr.Delete(ctx, pipeID); err != nil {
			respondWith(w, http.StatusInternalServerError, err)
		}
	}
}
