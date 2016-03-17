package multitenant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	consul "github.com/hashicorp/consul/api"
	"golang.org/x/net/context"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/common/network"
	"github.com/weaveworks/scope/common/xfer"
)

const (
	gcInterval       = 30 * time.Second // we check all the pipes every 30s
	pipeTimeout      = 1 * time.Minute  // pipes are closed when a client hasn't been connected for 1 minute
	gcTimeout        = 10 * time.Minute // after another 10 minutes, tombstoned pipes are forgotten
	longPollDuration = 10 * time.Second

	privateAPIPort = 4444
)

var (
	queryOptions = &consul.QueryOptions{
		RequireConsistent: true,
	}
	writeOptions = &consul.WriteOptions{}
	wsDialer     = &websocket.Dialer{}
)

// TODO deal with garbage collection
type consulPipe struct {
	CreatedAt, DeletedAt time.Time
	UIEnd, ProbeEnd      string // Addrs where each end is connected
	UIRef, ProbeRef      int    // Ref counts
}

func (c *consulPipe) toBytes() ([]byte, error) {
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *consulPipe) fromBytes(bs []byte) error {
	return json.NewDecoder(bytes.NewReader(bs)).Decode(c)
}

func (c *consulPipe) setEnd(e app.End, addr string) {
	if e == app.UIEnd {
		c.UIEnd = addr
	} else {
		c.ProbeEnd = addr
	}
}

func (c *consulPipe) end(e app.End) string {
	if e == app.UIEnd {
		return c.UIEnd
	}
	return c.ProbeEnd
}

func (c *consulPipe) otherEnd(e app.End) string {
	if e == app.UIEnd {
		return c.ProbeEnd
	}
	return c.UIEnd
}

func (c *consulPipe) eitherEndFor(addr string) bool {
	return c.end(app.UIEnd) == addr || c.end(app.ProbeEnd) == addr
}

func (c *consulPipe) incr(e app.End) int {
	if e == app.UIEnd {
		c.UIRef++
		return c.UIRef
	}
	c.ProbeRef++
	return c.ProbeRef
}

func (c *consulPipe) decr(e app.End) int {
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
	client    *consul.Client
	userIDer  UserIDer

	pipes     map[string]xfer.Pipe // Active pipes
	bridges   map[string]*bridgeConnection
	actorChan chan func()

	// Used by Stop()
	quit chan struct{}
	wait sync.WaitGroup
}

// NewConsulPipeRouter returns a new consul based router
func NewConsulPipeRouter(addr, prefix, inf string, userIDer UserIDer) (app.PipeRouter, error) {
	advertise, err := network.GetFirstAddressOf(inf)
	if err != nil {
		return nil, err
	}
	client, err := consul.NewClient(&consul.Config{
		Address: addr,
		Scheme:  "http",
	})
	if err != nil {
		return nil, err
	}
	pipeRouter := &consulPipeRouter{
		prefix:    prefix,
		advertise: advertise,
		client:    client,
		userIDer:  userIDer,

		pipes:     map[string]xfer.Pipe{},
		bridges:   map[string]*bridgeConnection{},
		actorChan: make(chan func()),
		quit:      make(chan struct{}),
	}
	pipeRouter.wait.Add(2)
	go pipeRouter.watchAll()
	go pipeRouter.privateAPI()
	go pipeRouter.actor()
	return pipeRouter, nil
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
	var (
		index = uint64(0)
		kv    = pr.client.KV()
	)
	for {
		select {
		case <-pr.quit:
			pr.wait.Done()
			return
		default:
		}

		kvps, meta, err := kv.List(pr.prefix, &consul.QueryOptions{
			RequireConsistent: true,
			WaitIndex:         index,
			WaitTime:          longPollDuration,
		})
		if err != nil {
			log.Errorf("Error getting path %s: %v", pr.prefix, err)
			continue
		}
		if index == meta.LastIndex {
			continue
		}
		index = meta.LastIndex

		for _, kvp := range kvps {
			//log.Infof("Got background update to %s (%d)", kvp.Key, index)

			cp := consulPipe{}
			if err := cp.fromBytes(kvp.Value); err != nil {
				log.Errorf("Error deserialising pipe %s: %s", kvp.Key, err)
				continue
			}

			pr.actorChan <- func() { pr.handlePipeUpdate(kvp.Key, cp) }
		}
	}
}

func (pr *consulPipeRouter) handlePipeUpdate(key string, cp consulPipe) {
	log.Infof("Got update to pipe %s", key)

	// 1. If this pipe is closed, or we're not one of the ends, we
	//    should ensure our local pipe (and bridge) is closed.
	if !cp.DeletedAt.IsZero() || !cp.eitherEndFor(pr.advertise) {
		log.Infof("Pipe %s not in use on this node.", key)
		pipe, ok := pr.pipes[key]
		delete(pr.pipes, key)
		if ok {
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
	pipe, ok := pr.pipes[key]
	if !ok {
		pipe = xfer.NewPipe()
		pr.pipes[key] = pipe
	}

	// 3. Ensure there is a bridging connection for this pipe.
	//    Semantics are the owner of the UIEnd connects to the owner of the ProbeEnd
	shouldBridge := cp.DeletedAt.IsZero() &&
		cp.end(app.UIEnd) != cp.end(app.ProbeEnd) &&
		cp.end(app.UIEnd) == pr.advertise &&
		cp.end(app.ProbeEnd) != ""
	bridge, ok := pr.bridges[key]

	// If we shouldn't be bridging but are, or we should be bridging but are pointing
	// at the wrong place, stop the current bridge.
	if (!shouldBridge && ok) || (shouldBridge && ok && bridge.addr != cp.end(app.ProbeEnd)) {
		log.Infof("Stopping bridge connection for %s", key)
		delete(pr.bridges, key)
		bridge.stop()
		ok = false
	}

	// If we should be bridging and are not, start a new bridge
	if shouldBridge && !ok {
		log.Infof("Starting bridge connection for %s", key)
		bridge = newBridgeConnection(key, cp.end(app.ProbeEnd), pipe)
		pr.bridges[key] = bridge
	}
}

func (pr *consulPipeRouter) privateAPI() {
	router := mux.NewRouter()
	router.Methods("GET").
		MatcherFunc(app.URLMatcher("/private/api/pipe/{key}")).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				key = mux.Vars(r)["key"]
				pc  = make(chan xfer.Pipe)
			)
			pr.actorChan <- func() {
				pc <- pr.pipes[key]
			}
			pipe := <-pc
			if pipe == nil {
				http.NotFound(w, r)
				return
			}

			conn, err := xfer.Upgrade(w, r, nil)
			if err != nil {
				log.Errorf("Error upgrading pipe %s websocket: %v", key, err)
				return
			}
			defer conn.Close()

			end, _ := pipe.Ends()
			if err := pipe.CopyToWebsocket(end, conn); err != nil && !xfer.IsExpectedWSCloseError(err) {
				log.Printf("Error copying to pipe %s websocket: %v", key, err)
			}
		})

	addr := fmt.Sprintf("%s:%d", pr.advertise, privateAPIPort)
	log.Infof("Serving private API on endpoint %s.", addr)
	log.Infof("Private API terminated: %v", http.ListenAndServe(addr, router))
}

// Atomically modify a pipe in a callback.
// If pipe doesn't exist you'll get nil in callback.
func (pr *consulPipeRouter) get(key string) (*consulPipe, error) {
	var (
		kv   = pr.client.KV()
		pipe consulPipe
	)
	kvp, _, err := kv.Get(key, queryOptions)
	if err != nil {
		return nil, err
	}
	if kvp == nil {
		return nil, nil
	}
	if err := pipe.fromBytes(kvp.Value); err != nil {
		return nil, err
	}
	return &pipe, nil
}

// Atomically modify a pipe in a callback.
// If pipe doesn't exist you'll get nil in callback.
func (pr *consulPipeRouter) cas(key string, f func(*consulPipe) (*consulPipe, bool, error)) (*consulPipe, error) {
	var (
		index   = uint64(0)
		kv      = pr.client.KV()
		pipe    *consulPipe
		retries = 10
		retry   = true
	)
	for i := 0; i < retries; i++ {
		kvp, _, err := kv.Get(key, queryOptions)
		if err != nil {
			log.Errorf("Error getting %s: %v", key, err)
			continue
		}
		if kvp != nil {
			pipe = &consulPipe{}
			if err := pipe.fromBytes(kvp.Value); err != nil {
				log.Errorf("Error deserialising pipe %s: %v", key, err)
				continue
			}
			index = kvp.ModifyIndex // if it doesn't exist, it will be 0
		}

		if pipe, retry, err = f(pipe); err != nil {
			log.Errorf("Error CASing pipe %s: %v", key, err)
			if !retry {
				return nil, err
			}
			continue
		}

		if pipe == nil {
			panic("Callback must instantiate pipe!")
		}

		value, err := pipe.toBytes()
		if err != nil {
			log.Errorf("Error serialising pipe %s: %v", key, err)
			continue
		}
		ok, _, err := kv.CAS(&consul.KVPair{
			Key:         key,
			Value:       value,
			ModifyIndex: index,
		}, writeOptions)
		if err != nil {
			log.Errorf("Error CASing pipe %s: %v", key, err)
			continue
		}
		if !ok {
			log.Errorf("Error CASing pipe %s, trying again %d", key, index)
			continue
		}
		return pipe, nil
	}
	return nil, fmt.Errorf("Failed to aquire pipe")
}

// Watch a given pipe and trigger a callback when it changes.
// if callback returns false or error, exit (with the error).
func (pr *consulPipeRouter) watch(key string, deadline time.Time, f func(*consulPipe) (bool, error)) (*consulPipe, error) {
	var (
		index = uint64(0)
		kv    = pr.client.KV()
	)
	for deadline.After(mtime.Now()) {
		// Poll waiting for the entry to get updated
		kvp, meta, err := kv.Get(key, &consul.QueryOptions{
			RequireConsistent: true,
			WaitIndex:         index,
			WaitTime:          longPollDuration,
		})
		if err != nil {
			return nil, fmt.Errorf("Error getting %s: %v", key, err)
		}
		if kvp == nil {
			return nil, fmt.Errorf("Pipe %s unexpectedly deleted!", key)
		}

		pipe := &consulPipe{}
		pipe.fromBytes(kvp.Value)
		if ok, err := f(pipe); !ok {
			return pipe, nil
		} else if err != nil {
			return pipe, err
		}

		index = meta.LastIndex
	}
	return nil, fmt.Errorf("Timed out waiting on %s", key)
}

func (pr *consulPipeRouter) Exists(ctx context.Context, id string) (bool, error) {
	userID, err := pr.userIDer(ctx)
	if err != nil {
		return false, err
	}
	key := fmt.Sprintf("%s%s-%s", pr.prefix, userID, id)
	consulPipe, err := pr.get(key)
	if err != nil {
		return false, err
	}
	return consulPipe == nil || consulPipe.DeletedAt.IsZero(), nil
}

func (pr *consulPipeRouter) Get(ctx context.Context, id string, e app.End) (xfer.Pipe, io.ReadWriter, error) {
	userID, err := pr.userIDer(ctx)
	if err != nil {
		return nil, nil, err
	}
	key := fmt.Sprintf("%s%s-%s", pr.prefix, userID, id)
	log.Infof("Get %s:%s", key, e)

	// Try to ensure the given end of the given pipe
	// is 'owned' by this pipe service replica in consul.
	_, err = pr.cas(key, func(p *consulPipe) (*consulPipe, bool, error) {
		if p == nil {
			p = &consulPipe{
				CreatedAt: mtime.Now(),
			}
		}
		if !p.DeletedAt.IsZero() {
			return nil, false, fmt.Errorf("Pipe %s has been deleted", key)
		}
		end := p.end(e)
		if end != "" && end != pr.advertise {
			return nil, true, fmt.Errorf("Error: Pipe %s has existing connection to %s", key, end)
		}
		p.setEnd(e, pr.advertise)
		p.incr(e)
		return p, false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	// next see if we already have a active pipe
	pc := make(chan xfer.Pipe)
	pr.actorChan <- func() {
		pipe, ok := pr.pipes[key]
		if !ok {
			pipe = xfer.NewPipe()
			pr.pipes[key] = pipe
		}
		pc <- pipe
	}
	pipe := <-pc

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

	// atomically clear my end of the pipe in consul
	_, err = pr.cas(key, func(p *consulPipe) (*consulPipe, bool, error) {
		if p == nil {
			return nil, false, fmt.Errorf("Pipe %s not found", id)
		}
		if p.end(e) != pr.advertise {
			return nil, false, fmt.Errorf("Pipe %s not owned by us!", id)
		}
		refs := p.decr(e)
		if refs == 0 {
			p.setEnd(e, "")
		}
		return p, true, nil
	})
	return err
}

func (pr *consulPipeRouter) Delete(ctx context.Context, id string) error {
	userID, err := pr.userIDer(ctx)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s%s-%s", pr.prefix, userID, id)
	log.Infof("Delete %s", key)

	_, err = pr.cas(key, func(p *consulPipe) (*consulPipe, bool, error) {
		if p == nil {
			return nil, false, fmt.Errorf("Pipe %s not found", id)
		}
		p.DeletedAt = mtime.Now()
		return p, false, nil
	})
	return err
}

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
	bc.mtx.Lock()
	bc.stopped = true
	if bc.conn != nil {
		bc.conn.Close()
	}
	bc.mtx.Unlock()
	bc.wait.Wait()
}

func (bc *bridgeConnection) loop() {
	log.Infof("Making bridge connection for pipe %s to %s", bc.key, bc.addr)
	defer bc.wait.Done()
	defer log.Infof("Stopping bridge connection for pipe %s to %s", bc.key, bc.addr)

	_, end := bc.pipe.Ends()
	url := fmt.Sprintf("ws://%s:%d/private/api/pipe/%s", bc.addr, privateAPIPort, url.QueryEscape(bc.key))

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
			log.Errorf("Error connecting to %s: %v", url, err)
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

		if err := bc.pipe.CopyToWebsocket(end, conn); err != nil && !xfer.IsExpectedWSCloseError(err) {
			log.Printf("Error copying to pipe %s websocket: %v", bc.key, err)
		}
	}
}
