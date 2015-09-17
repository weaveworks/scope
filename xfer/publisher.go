package xfer

import "bytes"

// Publisher is something which can send a buffered set of data somewhere,
// probably to a remote collector.
type Publisher interface {
	Publish(*bytes.Buffer) error
	Stop()
}
