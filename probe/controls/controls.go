package controls

import (
	"sync"

	"github.com/weaveworks/scope/xfer"
)

var (
	mtx      = sync.Mutex{}
	handlers = map[string]xfer.ControlHandlerFunc{}
)

// HandleControlRequest performs a control request.
func HandleControlRequest(req xfer.Request) xfer.Response {
	mtx.Lock()
	handler, ok := handlers[req.Control]
	mtx.Unlock()
	if !ok {
		return xfer.ResponseErrorf("Control %q not recognised", req.Control)
	}

	return handler(req)
}

// Register a new control handler under a given id.
func Register(control string, f xfer.ControlHandlerFunc) {
	mtx.Lock()
	defer mtx.Unlock()
	handlers[control] = f
}

// Rm deletes the handler for a given name
func Rm(control string) {
	mtx.Lock()
	defer mtx.Unlock()
	delete(handlers, control)
}
