package app

import (
	"fmt"
	"math/rand"
	"sync"

	"golang.org/x/net/context"

	"github.com/weaveworks/scope/common/xfer"
)

// ControlRouter is a thing that can route control requests and responses
// between the UI and a probe.
type ControlRouter interface {
	Handle(ctx context.Context, probeID string, req xfer.Request) (xfer.Response, error)
	Register(ctx context.Context, probeID string, handler xfer.ControlHandlerFunc) (int64, error)
	Deregister(ctx context.Context, probeID string, id int64) error
}

// NewLocalControlRouter creates a new ControlRouter that does everything
// locally, in memory.
func NewLocalControlRouter() ControlRouter {
	return &localControlRouter{
		probes: map[string]probe{},
	}
}

type localControlRouter struct {
	sync.Mutex
	probes map[string]probe
}

type probe struct {
	id      int64
	handler xfer.ControlHandlerFunc
}

func (l *localControlRouter) Handle(_ context.Context, probeID string, req xfer.Request) (xfer.Response, error) {
	l.Lock()
	probe, ok := l.probes[probeID]
	l.Unlock()
	if !ok {
		return xfer.Response{}, fmt.Errorf("probe %s is not connected right now", probeID)
	}
	return probe.handler(req), nil
}

func (l *localControlRouter) Register(_ context.Context, probeID string, handler xfer.ControlHandlerFunc) (int64, error) {
	l.Lock()
	defer l.Unlock()
	id := rand.Int63()
	l.probes[probeID] = probe{
		id:      id,
		handler: handler,
	}
	return id, nil
}

func (l *localControlRouter) Deregister(_ context.Context, probeID string, id int64) error {
	l.Lock()
	defer l.Unlock()
	// NB probe might have reconnected in the mean time, need to ensure we do not
	// delete new connection!  Also, it might have connected then deleted itself!
	if l.probes[probeID].id == id {
		delete(l.probes, probeID)
	}
	return nil
}
