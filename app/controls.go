package app

import (
	"log"
	"math/rand"
	"net/http"
	"net/rpc"
	"sync"

	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/xfer"
)

// RegisterControlRoutes registers the various control routes with a http mux.
func RegisterControlRoutes(router *mux.Router) {
	controlRouter := &controlRouter{
		probes: map[string]controlHandler{},
	}
	router.Methods("GET").Path("/api/control/ws").HandlerFunc(controlRouter.handleProbeWS)
	router.Methods("POST").MatcherFunc(URLMatcher("/api/control/{probeID}/{nodeID}/{control}")).HandlerFunc(controlRouter.handleControl)
}

type controlHandler struct {
	id     int64
	client *rpc.Client
	codec  *xfer.JSONWebsocketCodec
}

type controlRouter struct {
	sync.Mutex
	probes map[string]controlHandler
}

func (ch *controlHandler) handle(req xfer.Request) xfer.Response {
	var res xfer.Response
	if err := ch.client.Call("control.Handle", req, &res); err != nil {
		return xfer.ResponseError(err)
	}
	return res
}

func (cr *controlRouter) get(probeID string) (controlHandler, bool) {
	cr.Lock()
	defer cr.Unlock()
	handler, ok := cr.probes[probeID]
	return handler, ok
}

func (cr *controlRouter) set(probeID string, handler controlHandler) {
	cr.Lock()
	defer cr.Unlock()
	cr.probes[probeID] = handler
}

func (cr *controlRouter) rm(probeID string, handler controlHandler) {
	cr.Lock()
	defer cr.Unlock()
	// NB probe might have reconnected in the mean time, need to ensure we do not
	// delete new connection!  Also, it might have connected then deleted itself!
	if cr.probes[probeID].id == handler.id {
		delete(cr.probes, probeID)
	}
}

// handleControl routes control requests from the client to the appropriate
// probe.  Its is blocking.
func (cr *controlRouter) handleControl(w http.ResponseWriter, r *http.Request) {
	var (
		vars    = mux.Vars(r)
		probeID = vars["probeID"]
		nodeID  = vars["nodeID"]
		control = vars["control"]
	)
	handler, ok := cr.get(probeID)
	if !ok {
		log.Printf("Probe %s is not connected right now...", probeID)
		http.NotFound(w, r)
		return
	}

	result := handler.handle(xfer.Request{
		AppID:   UniqueID,
		NodeID:  nodeID,
		Control: control,
	})
	if result.Error != "" {
		respondWith(w, http.StatusBadRequest, result.Error)
		return
	}
	respondWith(w, http.StatusOK, result)
}

// handleProbeWS accepts websocket connections from the probe and registers
// them in the control router, such that HandleControl calls can find them.
func (cr *controlRouter) handleProbeWS(w http.ResponseWriter, r *http.Request) {
	probeID := r.Header.Get(xfer.ScopeProbeIDHeader)
	if probeID == "" {
		respondWith(w, http.StatusBadRequest, xfer.ScopeProbeIDHeader)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to websocket: %v", err)
		return
	}
	defer conn.Close()

	codec := xfer.NewJSONWebsocketCodec(conn)
	client := rpc.NewClientWithCodec(codec)
	handler := controlHandler{
		id:     rand.Int63(),
		codec:  codec,
		client: client,
	}

	cr.set(probeID, handler)

	codec.WaitForReadError()

	cr.rm(probeID, handler)
	client.Close()
}
