package xfer

import (
	"io"
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
	readers   chan io.Reader
	quit      chan struct{}
}

// NewBackgroundPublisher creates a new BackgroundPublisher with the given publisher
func NewBackgroundPublisher(p Publisher) *BackgroundPublisher {
	bp := &BackgroundPublisher{
		publisher: p,
		readers:   make(chan io.Reader),
		quit:      make(chan struct{}),
	}
	go bp.loop()
	return bp
}

func (bp *BackgroundPublisher) loop() {
	backoff := initialBackoff

	for r := range bp.readers {
		err := bp.publisher.Publish(r)
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
func (bp *BackgroundPublisher) Publish(r io.Reader) error {
	select {
	case bp.readers <- r:
	default:
	}
	return nil
}

// Stop implements Publisher
func (bp *BackgroundPublisher) Stop() {
	close(bp.readers)
	close(bp.quit)
	bp.publisher.Stop()
}
