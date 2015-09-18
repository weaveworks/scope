package xfer

import (
	"bytes"
	"log"
	"time"
)

const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 60 * time.Second
)

// BackgroundPublisher is a publisher which does the publish asynchronously.
// It will only do one publish at once; if there is an ongoing publish,
// concurrent publishes are dropped.
type BackgroundPublisher struct {
	publisher Publisher
	buffers   chan *bytes.Buffer
	quit      chan struct{}
}

// NewBackgroundPublisher creates a new BackgroundPublisher with the given publisher
func NewBackgroundPublisher(p Publisher) *BackgroundPublisher {
	bp := &BackgroundPublisher{
		publisher: p,
		buffers:   make(chan *bytes.Buffer),
		quit:      make(chan struct{}),
	}
	go bp.loop()
	return bp
}

func (bp *BackgroundPublisher) loop() {
	backoff := initialBackoff

	for buf := range bp.buffers {
		err := bp.publisher.Publish(buf)
		if err == nil {
			backoff = initialBackoff
			continue
		}

		log.Printf("Error publishing to %s, backing off %s: %v", bp.publisher, backoff, err)
		select {
		case <-time.After(backoff):
		case <-bp.quit:
		}
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// Publish implements Publisher
func (bp *BackgroundPublisher) Publish(buf *bytes.Buffer) error {
	select {
	case bp.buffers <- buf:
	default:
	}
	return nil
}

// Stop implements Publisher
func (bp *BackgroundPublisher) Stop() {
	close(bp.buffers)
	close(bp.quit)
	bp.publisher.Stop()
}
