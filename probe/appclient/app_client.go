package appclient

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"

	"github.com/weaveworks/scope/common/sanitize"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/report"
)

const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 60 * time.Second
)

// AppClient is a client to an app for dealing with controls.
type AppClient interface {
	Details() (xfer.Details, error)
	ControlConnection()
	PipeConnection(string, xfer.Pipe)
	PipeClose(string) error
	Publish(r report.Report) error
	Stop()
}

// appClient is a client to an app for dealing with controls.
type appClient struct {
	ProbeConfig

	quit     chan struct{}
	mtx      sync.Mutex
	target   string
	client   http.Client
	wsDialer websocket.Dialer

	// Track all the background goroutines, ensure they all stop
	backgroundWait sync.WaitGroup

	// Track ongoing websocket connections
	conns map[string]*websocket.Conn

	// For publish
	publishLoop sync.Once
	reports     chan report.Report

	// For controls
	control xfer.ControlHandler
}

// NewAppClient makes a new appClient.
func NewAppClient(pc ProbeConfig, hostname, target string, control xfer.ControlHandler) (AppClient, error) {
	httpTransport, err := pc.getHTTPTransport(hostname)
	if err != nil {
		return nil, err
	}

	return &appClient{
		ProbeConfig: pc,
		quit:        make(chan struct{}),
		target:      target,
		client: http.Client{
			Transport: httpTransport,
		},
		wsDialer: websocket.Dialer{
			TLSClientConfig: httpTransport.TLSClientConfig,
		},
		conns:   map[string]*websocket.Conn{},
		reports: make(chan report.Report),
		control: control,
	}, nil
}

func (c *appClient) hasQuit() bool {
	select {
	case <-c.quit:
		return true
	default:
		return false
	}
}

func (c *appClient) registerConn(id string, conn *websocket.Conn) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.hasQuit() {
		conn.Close()
		return false
	}
	c.conns[id] = conn
	return true
}

func (c *appClient) closeConn(id string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if conn, ok := c.conns[id]; ok {
		conn.Close()
		delete(c.conns, id)
	}
}

func (c *appClient) retainGoroutine() bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.hasQuit() {
		return false
	}
	c.backgroundWait.Add(1)
	return true
}

func (c *appClient) releaseGoroutine() {
	c.backgroundWait.Done()
}

// Stop stops the appClient.
func (c *appClient) Stop() {
	c.mtx.Lock()
	close(c.reports)
	close(c.quit)
	for _, conn := range c.conns {
		conn.Close()
	}
	c.conns = map[string]*websocket.Conn{}
	c.mtx.Unlock()

	c.backgroundWait.Wait()
	c.client.Transport.(*http.Transport).CloseIdleConnections()
	return
}

// Details fetches the details (version, id) of the app.
func (c *appClient) Details() (xfer.Details, error) {
	result := xfer.Details{}
	req, err := c.ProbeConfig.authorizedRequest("GET", sanitize.URL("", 0, "/api")(c.target), nil)
	if err != nil {
		return result, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	return result, json.NewDecoder(resp.Body).Decode(&result)
}

func (c *appClient) doWithBackoff(msg string, f func() (bool, error)) {
	if !c.retainGoroutine() {
		return
	}
	defer c.releaseGoroutine()

	backoff := initialBackoff

	for {
		done, err := f()
		if done {
			return
		}
		if err == nil {
			backoff = initialBackoff
			continue
		}

		log.Errorf("Error doing %s for %s, backing off %s: %v", msg, c.target, backoff, err)
		select {
		case <-time.After(backoff):
		case <-c.quit:
			return
		}
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (c *appClient) controlConnection() (bool, error) {
	headers := http.Header{}
	c.ProbeConfig.authorizeHeaders(headers)
	url := sanitize.URL("ws://", 0, "/api/control/ws")(c.target)
	conn, _, err := c.wsDialer.Dial(url, headers)
	if err != nil {
		return false, err
	}
	defer func() {
		conn.Close()
	}()

	codec := xfer.NewJSONWebsocketCodec(conn)
	server := rpc.NewServer()
	if err := server.RegisterName("control", c.control); err != nil {
		return false, err
	}

	// Will return false if we are exiting
	if !c.registerConn("control", conn) {
		return true, nil
	}
	defer c.closeConn("control")

	server.ServeCodec(codec)
	return false, nil
}

func (c *appClient) ControlConnection() {
	go func() {
		log.Infof("Control connection to %s starting", c.target)
		defer log.Infof("Control connection to %s exiting", c.target)
		c.doWithBackoff("controls", c.controlConnection)
	}()
}

func (c *appClient) publishConnection() (*websocket.Conn, error) {
	headers := http.Header{}
	c.ProbeConfig.authorizeHeaders(headers)
	url := sanitize.URL("ws://", 0, "/api/report")(c.target)
	// TODO: connection leak
	conn, _, err := c.wsDialer.Dial(url, headers)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (c *appClient) startPublishing() {
	go func() {
		log.Infof("Publish loop for %s starting", c.target)
		defer log.Infof("Publish loop for %s exiting", c.target)
		// TODO: handle websocket reconnections
		conn, err := c.publishConnection()
		if err != nil {
			log.Errorf("Publish connection error: %s", err)
			return
		}
		defer conn.Close()
		rawConn := conn.UnderlyingConn()
		// No gzip for now
		// gzwriter := gzip.NewWriter(rawConn)
		// defer gzwriter.Close()
		encoder := gob.NewEncoder(rawConn)

		c.doWithBackoff("publish", func() (bool, error) {
			r, more := <-c.reports
			if !more {
				return true, nil
			}
			return false, encoder.Encode(r)
		})
	}()
}

// Publish implements Publisher
func (c *appClient) Publish(r report.Report) error {
	// Lazily start the background publishing loop.
	c.publishLoop.Do(c.startPublishing)
	select {
	case c.reports <- r:
	default:
	}
	return nil
}

func (c *appClient) pipeConnection(id string, pipe xfer.Pipe) (bool, error) {
	headers := http.Header{}
	c.ProbeConfig.authorizeHeaders(headers)
	url := sanitize.URL("ws://", 0, fmt.Sprintf("/api/pipe/%s/probe", id))(c.target)
	conn, resp, err := c.wsDialer.Dial(url, headers)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		// Special handling - 404 means the app/user has closed the pipe
		pipe.Close()
		return true, nil
	}
	if err != nil {
		return false, err
	}

	// Will return false if we are exiting
	if !c.registerConn(id, conn) {
		return true, nil
	}
	defer c.closeConn(id)

	_, remote := pipe.Ends()
	return false, pipe.CopyToWebsocket(remote, conn)
}

func (c *appClient) PipeConnection(id string, pipe xfer.Pipe) {
	go func() {
		log.Infof("Pipe %s connection to %s starting", id, c.target)
		defer log.Infof("Pipe %s connection to %s exiting", id, c.target)
		c.doWithBackoff(id, func() (bool, error) {
			return c.pipeConnection(id, pipe)
		})
	}()
}

// PipeClose closes the given pipe id on the app.
func (c *appClient) PipeClose(id string) error {
	url := sanitize.URL("", 0, fmt.Sprintf("/api/pipe/%s", id))(c.target)
	req, err := c.ProbeConfig.authorizedRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
