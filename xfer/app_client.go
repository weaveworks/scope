package xfer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/weaveworks/scope/common/sanitize"
)

const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 60 * time.Second
)

// Details are some generic details that can be fetched from /api
type Details struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Hostname string `json:"hostname"`
}

// AppClient is a client to an app for dealing with controls.
type AppClient interface {
	Details() (Details, error)
	ControlConnection(handler ControlHandler)
	Publish(r io.Reader) error
	Stop()
}

type appClient struct {
	ProbeConfig

	quit     chan struct{}
	target   string
	insecure bool
	client   http.Client

	// For publish
	publishLoop sync.Once
	readers     chan io.Reader

	// For controls
	controlServerCodecMtx sync.Mutex
	controlServerCodec    rpc.ServerCodec
}

// NewAppClient makes a new AppClient.
func NewAppClient(pc ProbeConfig, hostname, target string) (AppClient, error) {
	httpTransport, err := pc.getHTTPTransport(hostname)
	if err != nil {
		return nil, err
	}

	appClient := &appClient{
		ProbeConfig: pc,
		quit:        make(chan struct{}),
		readers:     make(chan io.Reader),
		target:      target,
		client: http.Client{
			Transport: httpTransport,
		},
	}

	return appClient, nil
}

// Stop stops the appClient.
func (c *appClient) Stop() {
	c.controlServerCodecMtx.Lock()
	defer c.controlServerCodecMtx.Unlock()
	close(c.readers)
	close(c.quit)
	if c.controlServerCodec != nil {
		c.controlServerCodec.Close()
	}
	c.client.Transport.(*http.Transport).CloseIdleConnections()
}

// Details fetches the details (version, id) of the app.
func (c *appClient) Details() (Details, error) {
	result := Details{}
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

		log.Printf("Error doing %s for %s, backing off %s: %v", msg, c.target, backoff, err)
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

func (c *appClient) controlConnection(handler ControlHandler) error {
	dialer := websocket.Dialer{}
	headers := http.Header{}
	c.ProbeConfig.authorizeHeaders(headers)
	// TODO(twilkie) need to update sanitize to work with wss
	url := sanitize.URL("ws://", 0, "/api/control/ws")(c.target)
	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return err
	}
	defer func() {
		log.Printf("Closing control connection to %s", c.target)
		conn.Close()
	}()

	codec := NewJSONWebsocketCodec(conn)
	server := rpc.NewServer()
	if err := server.RegisterName("control", handler); err != nil {
		return err
	}

	c.controlServerCodecMtx.Lock()
	c.controlServerCodec = codec
	// At this point we may have tried to quit earlier, so check to see if the
	// quit channel has been closed, non-blocking.
	select {
	default:
	case <-c.quit:
		codec.Close()
		return nil
	}
	c.controlServerCodecMtx.Unlock()

	server.ServeCodec(codec)

	c.controlServerCodecMtx.Lock()
	c.controlServerCodec = nil
	c.controlServerCodecMtx.Unlock()
	return nil
}

func (c *appClient) ControlConnection(handler ControlHandler) {
	go func() {
		log.Printf("Control connection to %s starting", c.target)
		defer log.Printf("Control connection to %s exiting", c.target)
		c.doWithBackoff("controls", func() (bool, error) {
			return false, c.controlConnection(handler)
		})
	}()
}

func (c *appClient) publish(r io.Reader) error {
	url := sanitize.URL("", 0, "/api/report")(c.target)
	req, err := c.ProbeConfig.authorizedRequest("POST", url, r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Encoding", "gzip")
	// req.Header.Set("Content-Type", "application/binary") // TODO: we should use http.DetectContentType(..) on the gob'ed

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func (c *appClient) startPublishing() {
	go func() {
		log.Printf("Publish loop for %s starting", c.target)
		defer log.Printf("Publish loop for %s exiting", c.target)
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
func (c *appClient) Publish(r io.Reader) error {
	// Lazily start the background publishing loop.
	c.publishLoop.Do(c.startPublishing)
	select {
	case c.readers <- r:
	default:
	}
	return nil
}
