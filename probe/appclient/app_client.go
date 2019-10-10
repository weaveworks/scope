package appclient

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"sync"
	"time"

	"github.com/armon/go-metrics"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-cleanhttp"
	log "github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/common/xfer"
)

const (
	httpClientTimeout = 12 * time.Second // a bit less than default app.window
	initialBackoff    = 1 * time.Second
	maxBackoff        = 60 * time.Second
)

// AppClient is a client to an app, dealing with report publishing, controls and pipes.
type AppClient interface {
	Details() (xfer.Details, error)
	ControlConnection()
	PipeConnection(string, xfer.Pipe)
	PipeClose(string) error
	Publish(io.Reader, bool) error
	Target() url.URL
	ReTarget(url.URL)
	Stop()
}

// appClient is a client to an app, dealing with report publishing, controls and pipes.
type appClient struct {
	ProbeConfig

	quit     chan struct{}
	mtx      sync.Mutex
	client   *http.Client
	wsDialer websocket.Dialer
	appID    string
	hostname string
	target   url.URL

	// Track all the background goroutines, ensure they all stop
	backgroundWait sync.WaitGroup

	// Track ongoing websocket connections
	conns map[string]xfer.Websocket

	// For publish
	publishLoop sync.Once
	readers     chan io.Reader

	// For controls
	control xfer.ControlHandler
}

// NewAppClient makes a new appClient.
func NewAppClient(pc ProbeConfig, hostname string, target url.URL, control xfer.ControlHandler) (AppClient, error) {
	httpTransport := pc.getHTTPTransport(hostname)
	httpClient := cleanhttp.DefaultClient()
	httpClient.Transport = httpTransport
	httpClient.Timeout = httpClientTimeout

	return &appClient{
		ProbeConfig: pc,
		quit:        make(chan struct{}),
		hostname:    hostname,
		target:      target,
		client:      httpClient,
		wsDialer: websocket.Dialer{
			TLSClientConfig:  httpTransport.TLSClientConfig,
			HandshakeTimeout: httpClientTimeout,
		},
		conns:   map[string]xfer.Websocket{},
		readers: make(chan io.Reader, 2),
		control: control,
	}, nil
}

func (c *appClient) url(path string) string {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.target.String() + path
}

func (c *appClient) wsURL(path string) string {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	output := c.target //copy the url
	if output.Scheme == "https" {
		output.Scheme = "wss"
	} else {
		output.Scheme = "ws"
	}
	return output.String() + path
}

func (c *appClient) hasQuit() bool {
	select {
	case <-c.quit:
		return true
	default:
		return false
	}
}

func (c *appClient) registerConn(id string, conn xfer.Websocket) bool {
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

func (c *appClient) Target() url.URL {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.target
}

// Re-target the appClient, publishing to a new URL. Note that control
// and pipe websocket connections are left untouched since we don't
// want to disrupt them just because there's some load-balancing going
// on. They *will* however pick up the new URL when terminating
// (e.g. due to errors or when the connection drops).
func (c *appClient) ReTarget(target url.URL) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.target = target
}

// Stop stops the appClient.
func (c *appClient) Stop() {
	c.mtx.Lock()
	close(c.readers)
	close(c.quit)
	for _, conn := range c.conns {
		conn.Close()
	}
	c.conns = map[string]xfer.Websocket{}
	c.mtx.Unlock()

	c.backgroundWait.Wait()
	c.client.Transport.(*http.Transport).CloseIdleConnections()
	return
}

// Details fetches the details (version, id) of the app.
func (c *appClient) Details() (xfer.Details, error) {
	result := xfer.Details{}
	req, err := c.ProbeConfig.authorizedRequest("GET", c.url("/api"), nil)
	if err != nil {
		return result, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return result, err
	}
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("Error response from %s: %s", c.url("/api"), resp.Status)
	}
	defer resp.Body.Close()
	if err := codec.NewDecoder(resp.Body, &codec.JsonHandle{}).Decode(&result); err != nil {
		return result, err
	}
	c.appID = result.ID
	return result, nil
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
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// The timeout period itself serves as a backoff that
			// prevents thrashing. Hence there is no need to introduce
			// further delays. Moreover, any delays between publishing
			// reports that exceed the app.window (defaults to 15s)
			// cause the UI to display no data, which is debilitating.
			log.Errorf("Error doing %s for %s: %v", msg, c.hostname, err)
			backoff = initialBackoff
			continue
		}
		log.Errorf("Error doing %s for %s, backing off %s: %v", msg, c.hostname, backoff, err)
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
	url := c.wsURL("/api/control/ws")
	conn, _, err := xfer.DialWS(&c.wsDialer, url, headers)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	doControl := func(req xfer.Request) xfer.Response {
		req.AppID = c.appID
		var res xfer.Response
		c.control.Handle(req, &res)
		return res
	}

	codec := xfer.NewJSONWebsocketCodec(conn)
	server := rpc.NewServer()
	if err := server.RegisterName("control", xfer.ControlHandlerFunc(doControl)); err != nil {
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
		log.Infof("Control connection to %s starting", c.hostname)
		defer log.Infof("Control connection to %s exiting", c.hostname)
		c.doWithBackoff("controls", c.controlConnection)
	}()
}

func (c *appClient) publish(r io.Reader) error {
	url := c.url("/api/report")
	req, err := c.ProbeConfig.authorizedRequest("POST", url, r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/msgpack")
	// req.Header.Set("Content-Type", "application/binary") // TODO: we should use http.DetectContentType(..) on the gob'ed

	// Make sure this request is cancelled when we stop the client
	req.Cancel = c.quit

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	metrics.IncrCounterWithLabels([]string{"publishes"}, 1, []metrics.Label{
		{Name: "destination", Value: req.Host},
		{Name: "status", Value: fmt.Sprint(resp.StatusCode)},
	})
	if resp.StatusCode != http.StatusOK {
		text, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf(resp.Status + ": " + string(text))
	}
	return nil
}

func (c *appClient) startPublishing() {
	go func() {
		log.Infof("Publish loop for %s starting", c.hostname)
		defer log.Infof("Publish loop for %s exiting", c.hostname)
		c.doWithBackoff("publish", func() (bool, error) {
			r := <-c.readers
			if r == nil {
				return true, nil
			}
			return false, c.publish(r)
		})
	}()
}

// Publish implements Publisher
func (c *appClient) Publish(r io.Reader, shortcut bool) error {
	// Lazily start the background publishing loop.
	c.publishLoop.Do(c.startPublishing)
	// enqueue report
	select {
	case c.readers <- r:
	default:
		log.Warnf("Dropping report to %s", c.hostname)
		if shortcut {
			return nil
		}
		// drop an old report to make way for new one
		c.mtx.Lock()
		defer c.mtx.Unlock()
		select {
		case <-c.readers:
		default:
		}
		c.readers <- r
	}
	return nil
}

func (c *appClient) pipeConnection(id string, pipe xfer.Pipe) (bool, error) {
	headers := http.Header{}
	c.ProbeConfig.authorizeHeaders(headers)
	url := c.wsURL(fmt.Sprintf("/api/pipe/%s/probe", id))
	conn, resp, err := xfer.DialWS(&c.wsDialer, url, headers)
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
	if err := pipe.CopyToWebsocket(remote, conn); err != nil && !xfer.IsExpectedWSCloseError(err) {
		return false, err
	}
	return false, nil
}

func (c *appClient) PipeConnection(id string, pipe xfer.Pipe) {
	go func() {
		log.Infof("Pipe %s connection to %s starting", id, c.hostname)
		defer log.Infof("Pipe %s connection to %s exiting", id, c.hostname)
		c.doWithBackoff(id, func() (bool, error) {
			return c.pipeConnection(id, pipe)
		})
	}()
}

// PipeClose closes the given pipe id on the app.
func (c *appClient) PipeClose(id string) error {
	url := c.url(fmt.Sprintf("/api/pipe/%s", id))
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
