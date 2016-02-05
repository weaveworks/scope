package app

import (
	"io"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/common/xfer"
)

const (
	gcInterval  = 30 * time.Second // we check all the pipes every 30s
	pipeTimeout = 1 * time.Minute  // pipes are closed when a client hasn't been connected for 1 minute
	gcTimeout   = 10 * time.Minute // after another 10 minutes, tombstoned pipes are forgotten
)

// PipeRouter connects incoming and outgoing pipes.
type PipeRouter struct {
	sync.Mutex
	wait  sync.WaitGroup
	quit  chan struct{}
	pipes map[string]*pipe
}

// for each end of the pipe, we keep a reference count & lastUsedTIme,
// such that we can timeout pipes when either end is inactive.
type end struct {
	refCount     int
	lastUsedTime time.Time
}

type pipe struct {
	ui, probe     end
	tombstoneTime time.Time

	xfer.Pipe
}

// RegisterPipeRoutes registers the pipe routes
func RegisterPipeRoutes(router *mux.Router) *PipeRouter {
	pipeRouter := &PipeRouter{
		quit:  make(chan struct{}),
		pipes: map[string]*pipe{},
	}
	pipeRouter.wait.Add(1)
	go pipeRouter.gcLoop()
	router.Methods("GET").
		Path("/api/pipe/{pipeID}").
		HandlerFunc(pipeRouter.handleWs(func(p *pipe) (*end, io.ReadWriter) {
		uiEnd, _ := p.Ends()
		return &p.ui, uiEnd
	}))
	router.Methods("GET").
		Path("/api/pipe/{pipeID}/probe").
		HandlerFunc(pipeRouter.handleWs(func(p *pipe) (*end, io.ReadWriter) {
		_, probeEnd := p.Ends()
		return &p.probe, probeEnd
	}))
	router.Methods("DELETE", "POST").
		Path("/api/pipe/{pipeID}").
		HandlerFunc(pipeRouter.delete)
	return pipeRouter
}

// Stop stops the pipeRouter
func (pr *PipeRouter) Stop() {
	close(pr.quit)
	pr.wait.Wait()
}

func (pr *PipeRouter) gcLoop() {
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

func (pr *PipeRouter) timeout() {
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

func (pr *PipeRouter) garbageCollect() {
	pr.Lock()
	defer pr.Unlock()
	now := mtime.Now()
	for pipeID, pipe := range pr.pipes {
		if pipe.Closed() && now.Sub(pipe.tombstoneTime) >= gcTimeout {
			delete(pr.pipes, pipeID)
		}
	}
}

func (pr *PipeRouter) getOrCreate(id string) (*pipe, bool) {
	pr.Lock()
	defer pr.Unlock()
	p, ok := pr.pipes[id]
	if !ok {
		log.Infof("Creating pipe id %s", id)
		p = &pipe{
			ui:    end{lastUsedTime: mtime.Now()},
			probe: end{lastUsedTime: mtime.Now()},
			Pipe:  xfer.NewPipe(),
		}
		pr.pipes[id] = p
	}
	if p.Closed() {
		return nil, false
	}
	return p, true
}

func (pr *PipeRouter) retain(id string, pipe *pipe, end *end) bool {
	pr.Lock()
	defer pr.Unlock()
	if pipe.Closed() {
		return false
	}
	end.refCount++
	return true
}

func (pr *PipeRouter) release(id string, pipe *pipe, end *end) {
	pr.Lock()
	defer pr.Unlock()

	end.refCount--
	if end.refCount != 0 {
		return
	}

	if !pipe.Closed() {
		end.lastUsedTime = mtime.Now()
	}
}

func (pr *PipeRouter) handleWs(endSelector func(*pipe) (*end, io.ReadWriter)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pipeID := mux.Vars(r)["pipeID"]
		pipe, ok := pr.getOrCreate(pipeID)
		if !ok {
			http.NotFound(w, r)
			return
		}

		endRef, endIO := endSelector(pipe)
		if !pr.retain(pipeID, pipe, endRef) {
			http.NotFound(w, r)
			return
		}
		defer pr.release(pipeID, pipe, endRef)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Errorf("Error upgrading to websocket: %v", err)
			return
		}
		defer conn.Close()

		pipe.CopyToWebsocket(endIO, conn)
	}
}

func (pr *PipeRouter) delete(w http.ResponseWriter, r *http.Request) {
	pipeID := mux.Vars(r)["pipeID"]
	pipe, ok := pr.getOrCreate(pipeID)
	if ok && pr.retain(pipeID, pipe, &pipe.ui) {
		log.Infof("Closing pipe %s", pipeID)
		pipe.Close()
		pipe.tombstoneTime = mtime.Now()
		pr.release(pipeID, pipe, &pipe.ui)
	}
}
