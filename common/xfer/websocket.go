package xfer

import (
	"io"

	"github.com/gorilla/websocket"
	"github.com/ugorji/go/codec"
)

// IsExpectedWSCloseError returns boolean indicating whether the error is a
// clean disconnection.
func IsExpectedWSCloseError(err error) bool {
	return websocket.IsCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
	)
}

// WriteJSONtoWS writes the JSON encoding of v to the connection.
func WriteJSONtoWS(c *websocket.Conn, v interface{}) error {
	w, err := c.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	err1 := codec.NewEncoder(w, &codec.JsonHandle{}).Encode(v)
	err2 := w.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// ReadJSONfromWS reads the next JSON-encoded message from the connection and stores
// it in the value pointed to by v.
func ReadJSONfromWS(c *websocket.Conn, v interface{}) error {
	_, r, err := c.NextReader()
	if err != nil {
		return err
	}
	err = codec.NewDecoder(r, &codec.JsonHandle{}).Decode(v)
	if err == io.EOF {
		// One value is expected in the message.
		err = io.ErrUnexpectedEOF
	}
	return err
}
