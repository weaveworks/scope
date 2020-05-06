package multitenant

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"context"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/common/middleware"
	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/common/xfer"
)

const (
	gcInterval  = 30 * time.Second // we check all the pipes every 30s
	pipeTimeout = 1 * time.Minute  // pipes are closed when a client hasn't been connected for 1 minute
	gcTimeout   = 10 * time.Minute // after another 10 minutes, tombstoned pipes are forgotten
)

var (
	wsDialer = &websocket.Dialer{}
)

// TODO deal with garbage collection
type consulPipe struct {
	CreatedAt, DeletedAt time.Time
	UIAddr, ProbeAddr    string // Addrs where each end is connected
	UIRef, ProbeRef      int    // Ref counts
}

func (c *consulPipe) setAddrFor(e app.End, addr string) {
	if e == app.UIEnd {
		c.UIAddr = addr
	} else {
		c.ProbeAddr = addr
	}
}

func (c *consulPipe) addrFor(e app.End) string {
	if e == app.UIEnd {
		return c.UIAddr
	}
	return c.ProbeAddr
}

func (c *consulPipe) eitherEndFor(addr string) bool {
	return c.addrFor(app.UIEnd) == addr || c.addrFor(app.ProbeEnd) == addr
}

func (c *consulPipe) acquire(e app.End) int {
	if e == app.UIEnd {
		c.UIRef++
		return c.UIRef
	}
	c.ProbeRef++
	return c.ProbeRef
}

func (c *consulPipe) release(e app.End) int {
	if e == app.UIEnd {
		c.UIRef--
		return c.UIRef
	}
	c.ProbeRef--
	return c.ProbeRef
}

type consulPipeRouter struct {
	prefix    string
	advertise string // Address of this pipe router to advertise in consul
	client    ConsulClient
	userIDer  UserIDer

	activePipes map[string]xfer.Pipe
	bridges     map[string]*bridgeConnection
	actorChan   chan func()
	pipeWaiters map[string][]chan xfer.Pipe

	// Used by Stop()
	quit chan struct{}
	wait sync.WaitGroup
}

// NewConsulPipeRouter returns a new consul based router
func NewConsulPipeRouter(client ConsulClient, prefix, advertise string, userIDer UserIDer) app.PipeRouter {
	pipeRouter := &consulPipeRouter{
		prefix:    prefix,
		advertise: advertise,
		client:    client,
		userIDer:  userIDer,

		activePipes: map[string]xfer.Pipe{},
		bridges:     map[string]*bridgeConnection{},
		actorChan:   make(chan func()),
		pipeWaiters: map[string][]chan xfer.Pipe{},

		quit: make(chan struct{}),
	}
	pipeRouter.wait.Add(2)
	go pipeRouter.watchAll()
	go pipeRouter.actor()
	go pipeRouter.privateAPI()
	return pipeRouter
}

func (pr *consulPipeRouter) Stop() {
	close(pr.quit)
	pr.wait.Wait()
}

func (pr *consulPipeRouter) actor() {
	defer pr.wait.Done()
	for {
		select {
		case f := <-pr.actorChan:
			f()
		case <-pr.quit:
			return
		}
	}
}

// watchAll listens to all pipe updates from consul.
// This is effectively a distributed, consistent actor routine.
// All state changes for this pipe router happen in this loop,
// and all the methods are implemented as CAS's on consul, to
// trigger an event in this loop.
func (pr *consulPipeRouter) watchAll() {
	defer pr.wait.Done()
	pr.client.WatchPrefix(pr.prefix, &consulPipe{}, pr.quit, func(key string, value interface{}) bool {
		cp := *value.(*consulPipe)
		select {
		case pr.actorChan <- func() { pr.handlePipeUpdate(key, cp) }:
			return true
		case <-pr.quit:
			return false
		}
	})
}

func (pr *consulPipeRouter) handlePipeUpdate(key string, cp consulPipe) {
	// 1. If this pipe is closed, or we're not one of the ends, we
	//    should ensure our local pipe (and bridge) is closed.
	if !cp.DeletedAt.IsZero() || !cp.eitherEndFor(pr.advertise) {
		pipe, ok := pr.activePipes[key]
		delete(pr.activePipes, key)
		if ok {
			log.Infof("Deleting pipe %s", key)
			pipe.Close()
		}

		bridge, ok := pr.bridges[key]
		delete(pr.bridges, key)
		if ok {
			bridge.stop()
		}
		return
	}

	if !cp.eitherEndFor(pr.advertise) {
		return
	}

	// 2. If this pipe if for us, we should have a pipe for it.
	pipe, ok := pr.activePipes[key]
	if !ok {
		log.Infof("Creating pipe %s", key)
		pipe = xfer.NewPipe()
		pr.activePipes[key] = pipe
		for _, pw := range pr.pipeWaiters[key] {
			pw <- pipe
		}
		delete(pr.pipeWaiters, key)
	}

	// 3. Ensure there is a bridging connection for this pipe.
	//    Semantics are the owner of the UIEnd connects to the owner of the ProbeEnd
	shouldBridge := cp.DeletedAt.IsZero() &&
		cp.addrFor(app.UIEnd) != cp.addrFor(app.ProbeEnd) &&
		cp.addrFor(app.UIEnd) == pr.advertise &&
		cp.addrFor(app.ProbeEnd) != ""
	bridge, ok := pr.bridges[key]

	// If we shouldn't be bridging but are, or we should be bridging but are pointing
	// at the wrong place, stop the current bridge.
	if (!shouldBridge && ok) || (shouldBridge && ok && bridge.addr != cp.addrFor(app.ProbeEnd)) {
		delete(pr.bridges, key)
		bridge.stop()
		ok = false
	}

	// If we should be bridging and are not, start a new bridge
	if shouldBridge && !ok {
		bridge = newBridgeConnection(key, cp.addrFor(app.ProbeEnd), pipe)
		pr.bridges[key] = bridge
	}
}

func (pr *consulPipeRouter) getPipe(key string) xfer.Pipe {
	pc := make(chan xfer.Pipe)
	select {
	case pr.actorChan <- func() { pc <- pr.activePipes[key] }:
		return <-pc
	case <-pr.quit:
		return nil
	}
}

func (pr *consulPipeRouter) waitForPipe(key string) xfer.Pipe {
	pc := make(chan xfer.Pipe)
	select {
	case pr.actorChan <- func() {
		pipe, ok := pr.activePipes[key]
		if ok {
			pc <- pipe
		} else {
			pr.pipeWaiters[key] = append(pr.pipeWaiters[key], pc)
		}
	}:
		return <-pc
	case <-pr.quit:
		return nil
	}
}

func (pr *consulPipeRouter) privateAPI() {
	router := mux.NewRouter()
	router.Methods("GET").
		MatcherFunc(app.URLMatcher("/private/api/pipe/{key}")).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := mux.Vars(r)["key"]
			log.Infof("%s: Server bridge connection started", key)
			defer log.Infof("%s: Server bridge connection stopped", key)

			pipe := pr.getPipe(key)
			if pipe == nil {
				log.Errorf("%s: Server bridge connection; Unknown pipe!", key)
				w.WriteHeader(http.StatusNotFound)
				return
			}

			conn, err := xfer.Upgrade(w, r, nil)
			if err != nil {
				log.Errorf("%s: Server bridge connection; Error upgrading to websocket: %v", key, err)
				return
			}
			defer conn.Close()

			end, _ := pipe.Ends()
			if _, err := pipe.CopyToWebsocket(end, conn); err != nil {
				log.Errorf("%s: Server bridge connection; Error copying pipe to websocket: %v", key, err)
			}
		})

	handler := middleware.Tracer{RouteMatcher: router}.Wrap(router)
	log.Infof("Serving private API on endpoint %s.", pr.advertise)
	log.Infof("Private API terminated: %v", http.ListenAndServe(pr.advertise, handler))
}

func (pr *consulPipeRouter) Exists(ctx context.Context, id string) (bool, error) {
	userID, err := pr.userIDer(ctx)
	if err != nil {
		return false, err
	}
	key := fmt.Sprintf("%s%s-%s", pr.prefix, userID, id)
	consulPipe := consulPipe{}
	err = pr.client.Get(ctx, key, &consulPipe)
	if err == ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return consulPipe.DeletedAt.IsZero(), nil
}

func (pr *consulPipeRouter) Get(ctx context.Context, id string, e app.End) (xfer.Pipe, io.ReadWriter, error) {
	userID, err := pr.userIDer(ctx)
	if err != nil {
		return nil, nil, err
	}
	key := fmt.Sprintf("%s%s-%s", pr.prefix, userID, id)
	log.Infof("Get %s:%s", key, e)
	span, ctx := opentracing.StartSpanFromContext(ctx, "PipeRouter Get", opentracing.Tag{"key", key})
	defer span.Finish()

	// Try to ensure the given end of the given pipe
	// is 'owned' by this pipe service replica in consul.
	err = pr.client.CAS(ctx, key, &consulPipe{}, func(in interface{}) (interface{}, bool, error) {
		var pipe *consulPipe
		if in == nil {
			pipe = &consulPipe{
				CreatedAt: mtime.Now(),
			}
		} else {
			pipe = in.(*consulPipe)
		}
		if !pipe.DeletedAt.IsZero() {
			return nil, false, fmt.Errorf("Pipe %s has been deleted", key)
		}
		end := pipe.addrFor(e)
		if end != "" && end != pr.advertise {
			return nil, true, fmt.Errorf("Error: Pipe %s has existing connection to %s", key, end)
		}
		pipe.setAddrFor(e, pr.advertise)
		pipe.acquire(e)
		return pipe, false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	pipe := pr.waitForPipe(key)
	myEnd, _ := pipe.Ends()
	if e == app.ProbeEnd {
		_, myEnd = pipe.Ends()
	}
	return pipe, myEnd, nil
}

func (pr *consulPipeRouter) Release(ctx context.Context, id string, e app.End) error {
	userID, err := pr.userIDer(ctx)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s%s-%s", pr.prefix, userID, id)
	log.Infof("Release %s:%s", key, e)
	span, ctx := opentracing.StartSpanFromContext(ctx, "PipeRouter Release", opentracing.Tag{"key", key})
	defer span.Finish()

	// atomically clear my end of the pipe in consul
	return pr.client.CAS(ctx, key, &consulPipe{}, func(in interface{}) (interface{}, bool, error) {
		if in == nil {
			return nil, false, fmt.Errorf("pipe %s not found", id)
		}
		p := in.(*consulPipe)
		if p.addrFor(e) != pr.advertise {
			return nil, false, fmt.Errorf("pipe %s not owned by us", id)
		}
		refs := p.release(e)
		if refs == 0 {
			p.setAddrFor(e, "")
		}
		return p, true, nil
	})
}

func (pr *consulPipeRouter) Delete(ctx context.Context, id string) error {
	userID, err := pr.userIDer(ctx)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s%s-%s", pr.prefix, userID, id)
	log.Infof("Delete %s", key)
	span, ctx := opentracing.StartSpanFromContext(ctx, "PipeRouter Delete", opentracing.Tag{"key", key})
	defer span.Finish()

	return pr.client.CAS(ctx, key, &consulPipe{}, func(in interface{}) (interface{}, bool, error) {
		if in == nil {
			return nil, false, fmt.Errorf("Pipe %s not found", id)
		}
		p := in.(*consulPipe)
		p.DeletedAt = mtime.Now()
		return p, false, nil
	})
}

// A bridgeConnection represents a connection between two pipe router replicas.
// They are created & destroyed in response to events from consul, which in turn
// are triggered when UIs or Probes connect to various pipe routers.
type bridgeConnection struct {
	key  string
	addr string // address to connect to
	pipe xfer.Pipe

	mtx     sync.Mutex
	conn    xfer.Websocket
	stopped bool
	wait    sync.WaitGroup
}

func newBridgeConnection(key, addr string, pipe xfer.Pipe) *bridgeConnection {
	log.Infof("%s: Starting client bridge connection", key)
	result := &bridgeConnection{
		key:  key,
		addr: addr,
		pipe: pipe,
	}
	result.wait.Add(1)
	go result.loop()
	return result
}

func (bc *bridgeConnection) stop() {
	log.Infof("%s: Stopping client bridge connection", bc.key)
	bc.mtx.Lock()
	bc.stopped = true
	if bc.conn != nil {
		bc.conn.Close()
		end, _ := bc.pipe.Ends()
		end.Write(nil) // this will cause the other end of wake up and exit
	}
	bc.mtx.Unlock()
	bc.wait.Wait()
}

func (bc *bridgeConnection) loop() {
	log.Infof("%s: Client bridge connection started", bc.key)
	defer bc.wait.Done()
	defer log.Infof("%s: Client bridge connection stopped", bc.key)

	_, end := bc.pipe.Ends()
	url := fmt.Sprintf("ws://%s/private/api/pipe/%s", bc.addr, url.QueryEscape(bc.key))

	for {
		bc.mtx.Lock()
		bc.conn = nil
		if bc.stopped {
			bc.mtx.Unlock()
			return
		}
		bc.mtx.Unlock()

		// connect to other pipes instance
		conn, _, err := xfer.DialWS(wsDialer, url, http.Header{})
		if err != nil {
			log.Errorf("%s: Client bridge connection; Error connecting to %s: %v", bc.key, url, err)
			time.Sleep(time.Second) // TODO backoff
			continue
		}

		bc.mtx.Lock()
		if bc.stopped {
			bc.mtx.Unlock()
			conn.Close()
			return
		}
		bc.conn = conn
		bc.mtx.Unlock()

		if _, err := bc.pipe.CopyToWebsocket(end, conn); err != nil {
			log.Errorf("%s: Client bridge connection; Error copying pipe to websocket: %v", bc.key, err)
		}
		conn.Close()
	}
}
