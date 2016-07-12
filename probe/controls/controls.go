package controls

import (
	"sync"

	"github.com/weaveworks/scope/common/xfer"
)

// HandlerRegistryBackend is an interface for storing control request
// handlers.
type HandlerRegistryBackend interface {
	// Lock locks the backend, so the batch insertions or
	// removals can be performed.
	Lock()
	// Unlock unlocks the registry.
	Unlock()
	// Register a new control handler under a given
	// id. Implementations should not call Lock() or Unlock()
	// here, it will be done by HandlerRegistry.
	Register(control string, f xfer.ControlHandlerFunc)
	// Rm deletes the handler for a given name. Implementations
	// should not call Lock() or Unlock() here, it will be done by
	// HandlerRegistry.
	Rm(control string)
	// Handler gets the handler for a control. Implementations
	// should not call Lock() or Unlock() here, it will be done by
	// HandlerRegistry.
	Handler(control string) (xfer.ControlHandlerFunc, bool)
}

type defaultBackend struct {
	handlers map[string]xfer.ControlHandlerFunc
	mtx      sync.Mutex
}

// NewDefaultHandlerRegistryBackend creates a default backend for
// handler registry.
func NewDefaultHandlerRegistryBackend() HandlerRegistryBackend {
	return &defaultBackend{
		handlers: map[string]xfer.ControlHandlerFunc{},
	}
}

// Lock locks the registry, so the batch insertions or
// removals can be performed.
func (b *defaultBackend) Lock() {
	b.mtx.Lock()
}

// Unlock unlocks the registry.
func (b *defaultBackend) Unlock() {
	b.mtx.Unlock()
}

// Register a new control handler under a given id.
func (b *defaultBackend) Register(control string, f xfer.ControlHandlerFunc) {
	b.handlers[control] = f
}

// Rm deletes the handler for a given name.
func (b *defaultBackend) Rm(control string) {
	delete(b.handlers, control)
}

// Handler gets the handler for a control.
func (b *defaultBackend) Handler(control string) (xfer.ControlHandlerFunc, bool) {
	handler, ok := b.handlers[control]
	return handler, ok
}

// HandlerRegistry uses backend for storing and retrieving control
// requests handlers.
type HandlerRegistry struct {
	backend HandlerRegistryBackend
}

// NewDefaultHandlerRegistry creates a registry with a default
// backend.
func NewDefaultHandlerRegistry() *HandlerRegistry {
	return NewHandlerRegistry(NewDefaultHandlerRegistryBackend())
}

// NewHandlerRegistry creates a registry with a custom backend.
func NewHandlerRegistry(backend HandlerRegistryBackend) *HandlerRegistry {
	return &HandlerRegistry{
		backend: backend,
	}
}

// Register registers a new control handler under a given name.
func (r *HandlerRegistry) Register(control string, f xfer.ControlHandlerFunc) {
	r.backend.Lock()
	defer r.backend.Unlock()
	r.backend.Register(control, f)
}

// Rm deletes the handler for a given name.
func (r *HandlerRegistry) Rm(control string) {
	r.backend.Lock()
	defer r.backend.Unlock()
	r.backend.Rm(control)
}

// Batch first deletes handlers for given names in toRemove then
// registers new handlers for given names in toAdd.
func (r *HandlerRegistry) Batch(toRemove []string, toAdd map[string]xfer.ControlHandlerFunc) {
	r.backend.Lock()
	defer r.backend.Unlock()
	for _, control := range toRemove {
		r.backend.Rm(control)
	}
	for control, handler := range toAdd {
		r.backend.Register(control, handler)
	}
}

// HandleControlRequest performs a control request.
func (r *HandlerRegistry) HandleControlRequest(req xfer.Request) xfer.Response {
	h, ok := r.handler(req.Control)
	if !ok {
		return xfer.ResponseErrorf("Control %q not recognised", req.Control)
	}

	return h(req)
}

func (r *HandlerRegistry) handler(control string) (xfer.ControlHandlerFunc, bool) {
	r.backend.Lock()
	defer r.backend.Unlock()
	return r.backend.Handler(control)
}
