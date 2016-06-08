package app

import (
	"io"
	"net/http"
	"net/rpc"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/common/xfer"
)

// RegisterControlRoutes registers the various control routes with a http mux.
func RegisterControlRoutes(router *mux.Router, cr ControlRouter) {
	router.Methods("GET").Path("/api/control/ws").
		HandlerFunc(requestContextDecorator(handleProbeWS(cr)))
	router.Methods("POST").MatcherFunc(URLMatcher("/api/control/{probeID}/{nodeID}/{control}")).
		HandlerFunc(requestContextDecorator(handleControl(cr)))
}

// handleControl routes control requests from the client to the appropriate
// probe.  Its is blocking.
func handleControl(cr ControlRouter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var (
			vars    = mux.Vars(r)
			probeID = vars["probeID"]
			nodeID  = vars["nodeID"]
			control = vars["control"]
		)
		result, err := cr.Handle(ctx, probeID, xfer.Request{
			NodeID:  nodeID,
			Control: control,
		})
		if err != nil {
			respondWith(w, http.StatusBadRequest, err.Error())
			return
		}
		if result.Error != "" {
			respondWith(w, http.StatusBadRequest, result.Error)
			return
		}
		respondWith(w, http.StatusOK, result)
	}
}

// handleProbeWS accepts websocket connections from the probe and registers
// them in the control router, such that HandleControl calls can find them.
func handleProbeWS(cr ControlRouter) CtxHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		probeID := r.Header.Get(xfer.ScopeProbeIDHeader)
		if probeID == "" {
			respondWith(w, http.StatusBadRequest, xfer.ScopeProbeIDHeader)
			return
		}

		conn, err := xfer.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error upgrading control websocket: %v", err)
			return
		}
		defer conn.Close()

		codec := xfer.NewJSONWebsocketCodec(conn)
		client := rpc.NewClientWithCodec(codec)
		defer client.Close()

		id, err := cr.Register(ctx, probeID, func(req xfer.Request) xfer.Response {
			var res xfer.Response
			if err := client.Call("control.Handle", req, &res); err != nil {
				return xfer.ResponseError(err)
			}
			return res
		})
		if err != nil {
			respondWith(w, http.StatusBadRequest, err)
			return
		}
		defer cr.Deregister(ctx, probeID, id)
		if err := codec.WaitForReadError(); err != nil && err != io.EOF && !xfer.IsExpectedWSCloseError(err) {
			log.Errorf("Error on websocket: %v", err)
		}
	}
}
