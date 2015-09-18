package xfer

import "io"

// Publisher is something which can send a stream of data somewhere, probably
// to a remote collector.
type Publisher interface {
	Publish(io.Reader) error
	Stop()
}
