package xfer

import (
	"encoding/json"
	"log"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/weaveworks/scope/common/sanitize"
)

// Details are some generic details that can be fetched from /api
type Details struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// AppClient is a client to an app for dealing with controls.
type AppClient interface {
	Details() (Details, error)
	ControlConnection(handler ControlHandler)
	Stop()
}

type appClient struct {
	ProbeConfig

	quit     chan struct{}
	target   string
	insecure bool
	client   http.Client

	controlServerCodecMtx sync.Mutex
	controlServerCodec    rpc.ServerCodec
}

// NewAppClient makes a new AppClient.
func NewAppClient(pc ProbeConfig, hostname, target string) (AppClient, error) {
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
	}, nil
}

// Stop stops the appClient.
func (c *appClient) Stop() {
	c.controlServerCodecMtx.Lock()
	defer c.controlServerCodecMtx.Unlock()
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

func (c *appClient) controlConnectionLoop(handler ControlHandler) {
	defer log.Printf("Control connection to %s exiting", c.target)
	backoff := initialBackoff

	for {
		err := c.controlConnection(handler)
		if err == nil {
			backoff = initialBackoff
			continue
		}

		log.Printf("Error doing controls for %s, backing off %s: %v", c.target, backoff, err)
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

func (c *appClient) ControlConnection(handler ControlHandler) {
	go c.controlConnectionLoop(handler)
}
