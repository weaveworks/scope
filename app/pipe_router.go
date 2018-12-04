package app

import (
	"fmt"
	"io"
	"sync"
	"time"

	"context"
	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/common/xfer"
)

const (
	gcInterval  = 30 * time.Second // we check all the pipes every 30s
	pipeTimeout = 1 * time.Minute  // pipes are closed when a client hasn't been connected for 1 minute
	gcTimeout   = 10 * time.Minute // after another 10 minutes, tombstoned pipes are forgotten
)

// End is an enum for either end of the pipe.
type End int

// Valid values of type End
const (
	UIEnd = iota
	ProbeEnd
)

func (e End) String() string {
	if e == UIEnd {
		return "ui"
	}
	return "probe"
}

// PipeRouter stores pipes and allows you to connect to either end of them.
type PipeRouter interface {
	Exists(context.Context, string) (bool, error)
	Get(context.Context, string, End) (xfer.Pipe, io.ReadWriter, error)
	Release(context.Context, string, End) error
	Delete(context.Context, string) error
	Stop()
}

// PipeRouter connects incoming and outgoing pipes.
type localPipeRouter struct {
	sync.Mutex
	wait  sync.WaitGroup
	quit  chan struct{}
	pipes map[string]*pipe
}

// for each end of the pipe, we keep a reference count & lastUsedTIme,
// such that we can timeout pipes when either end is inactive.
type pipe struct {
	xfer.Pipe

	tombstoneTime time.Time

	ui, probe end
}

type end struct {
	refCount     int
	lastUsedTime time.Time
}

func (p *pipe) end(end End) (*end, io.ReadWriter) {
	ui, probe := p.Ends()
	if end == UIEnd {
		return &p.ui, ui
	}
	return &p.probe, probe
}

// NewLocalPipeRouter returns a new local (in-memory) pipe router.
func NewLocalPipeRouter() PipeRouter {
	pipeRouter := &localPipeRouter{
		quit:  make(chan struct{}),
		pipes: map[string]*pipe{},
	}
	pipeRouter.wait.Add(1)
	go pipeRouter.gcLoop()
	return pipeRouter
}

func (pr *localPipeRouter) Exists(_ context.Context, id string) (bool, error) {
	pr.Lock()
	defer pr.Unlock()
	p, ok := pr.pipes[id]
	if !ok {
		return true, nil
	}
	return !p.Closed(), nil
}

func (pr *localPipeRouter) Get(_ context.Context, id string, e End) (xfer.Pipe, io.ReadWriter, error) {
	pr.Lock()
	defer pr.Unlock()
	p, ok := pr.pipes[id]
	if !ok {
		log.Debugf("Creating pipe id %s", id)
		p = &pipe{
			ui:    end{lastUsedTime: mtime.Now()},
			probe: end{lastUsedTime: mtime.Now()},
			Pipe:  xfer.NewPipe(),
		}
		pr.pipes[id] = p
	}
	if p.Closed() {
		return nil, nil, fmt.Errorf("Pipe %s closed", id)
	}
	end, endIO := p.end(e)
	end.refCount++
	return p, endIO, nil
}

func (pr *localPipeRouter) Release(_ context.Context, id string, e End) error {
	pr.Lock()
	defer pr.Unlock()

	p, ok := pr.pipes[id]
	if !ok {
		return fmt.Errorf("Pipe %s not found", id)
	}

	end, _ := p.end(e)
	end.refCount--
	if end.refCount > 0 {
		return nil
	}

	if !p.Closed() {
		end.lastUsedTime = mtime.Now()
	}

	return nil
}

func (pr *localPipeRouter) Delete(_ context.Context, id string) error {
	pr.Lock()
	defer pr.Unlock()
	p, ok := pr.pipes[id]
	if !ok {
		return nil
	}
	p.Close()
	p.tombstoneTime = mtime.Now()
	return nil
}

func (pr *localPipeRouter) Stop() {
	close(pr.quit)
	pr.wait.Wait()
}

func (pr *localPipeRouter) gcLoop() {
	defer pr.wait.Done()
	ticker := time.Tick(gcInterval)
	for {
		select {
		case <-pr.quit:
			return
		case <-ticker:
		}

		pr.timeout()
		pr.garbageCollect()
	}
}

func (pr *localPipeRouter) timeout() {
	pr.Lock()
	defer pr.Unlock()
	now := mtime.Now()
	for id, pipe := range pr.pipes {
		if pipe.Closed() || (pipe.ui.refCount > 0 && pipe.probe.refCount > 0) {
			continue
		}

		if (pipe.ui.refCount == 0 && now.Sub(pipe.ui.lastUsedTime) >= pipeTimeout) ||
			(pipe.probe.refCount == 0 && now.Sub(pipe.probe.lastUsedTime) >= pipeTimeout) {
			log.Infof("Timing out pipe %s", id)
			pipe.Close()
			pipe.tombstoneTime = now
		}
	}
}

func (pr *localPipeRouter) garbageCollect() {
	pr.Lock()
	defer pr.Unlock()
	now := mtime.Now()
	for pipeID, pipe := range pr.pipes {
		if pipe.Closed() && now.Sub(pipe.tombstoneTime) >= gcTimeout {
			delete(pr.pipes, pipeID)
		}
	}
}
