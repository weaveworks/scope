package xfer

import (
	"fmt"
	"net/rpc"
	"sync"

	"github.com/gorilla/websocket"
)

// Request is the UI -> App -> Probe message type for control RPCs
type Request struct {
	ID      int64
	NodeID  string
	Control string
}

// Response is the Probe -> App -> UI message type for the control RPCs.
type Response struct {
	ID    int64
	Value interface{}
	Error string
}

// ControlHandler is interface used in the app and the probe to represent
// a control RPC.
type ControlHandler interface {
	Handle(req Request, res *Response) error
}

// ControlHandlerFunc is a adapter (ala golang's http RequestHandlerFunc)
// for ControlHandler
type ControlHandlerFunc func(Request) Response

// Handle is an adapter method to make ControlHandlers exposable via golang rpc
func (c ControlHandlerFunc) Handle(req Request, res *Response) error {
	*res = c(req)
	return nil
}

// ResponseErrorf creates a new Response with the given formatted error string.
func ResponseErrorf(format string, a ...interface{}) Response {
	return Response{
		Error: fmt.Sprintf(format, a...),
	}
}

// ResponseError creates a new Response with the given error.
func ResponseError(err error) Response {
	if err != nil {
		return Response{
			Error: err.Error(),
		}
	}
	return Response{}
}

// JSONWebsocketCodec is golang rpc compatible Server and Client Codec
// that transmits and receives RPC messages over a websocker, as JSON.
type JSONWebsocketCodec struct {
	sync.Mutex
	conn *websocket.Conn
	err  chan struct{}
}

// NewJSONWebsocketCodec makes a new JSONWebsocketCodec
func NewJSONWebsocketCodec(conn *websocket.Conn) *JSONWebsocketCodec {
	return &JSONWebsocketCodec{
		conn: conn,
		err:  make(chan struct{}),
	}
}

// WaitForReadError blocks until any read on this codec returns an error.
// This is useful to know when the server has disconnected from the client.
func (j *JSONWebsocketCodec) WaitForReadError() {
	<-j.err
}

// WriteRequest implements rpc.ClientCodec
func (j *JSONWebsocketCodec) WriteRequest(r *rpc.Request, v interface{}) error {
	j.Lock()
	defer j.Unlock()

	if err := j.conn.WriteJSON(r); err != nil {
		return err
	}
	return j.conn.WriteJSON(v)
}

// ReadResponseHeader implements rpc.ClientCodec
func (j *JSONWebsocketCodec) ReadResponseHeader(r *rpc.Response) error {
	err := j.conn.ReadJSON(r)
	if err != nil {
		close(j.err)
	}
	return err
}

// ReadResponseBody implements rpc.ClientCodec
func (j *JSONWebsocketCodec) ReadResponseBody(v interface{}) error {
	err := j.conn.ReadJSON(v)
	if err != nil {
		close(j.err)
	}
	return err
}

// Close implements rpc.ClientCodec and rpc.ServerCodec
func (j *JSONWebsocketCodec) Close() error {
	return j.conn.Close()
}

// ReadRequestHeader implements rpc.ServerCodec
func (j *JSONWebsocketCodec) ReadRequestHeader(r *rpc.Request) error {
	err := j.conn.ReadJSON(r)
	if err != nil {
		close(j.err)
	}
	return err
}

// ReadRequestBody implements rpc.ServerCodec
func (j *JSONWebsocketCodec) ReadRequestBody(v interface{}) error {
	err := j.conn.ReadJSON(v)
	if err != nil {
		close(j.err)
	}
	return err
}

// WriteResponse implements rpc.ServerCodec
func (j *JSONWebsocketCodec) WriteResponse(r *rpc.Response, v interface{}) error {
	j.Lock()
	defer j.Unlock()

	if err := j.conn.WriteJSON(r); err != nil {
		return err
	}
	return j.conn.WriteJSON(v)
}
