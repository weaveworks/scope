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
	reports   chan *bytes.Buffer
	quit      chan struct{}
}

// NewBackgroundPublisher creates a new BackgroundPublisher with the given publisher
func NewBackgroundPublisher(p Publisher) *BackgroundPublisher {
	result := &BackgroundPublisher{
		publisher: p,
		reports:   make(chan *bytes.Buffer),
		quit:      make(chan struct{}),
	}
	go result.loop()
	return result
}

func (b *BackgroundPublisher) loop() {
	backoff := initialBackoff

	for r := range b.reports {
		err := b.publisher.Publish(r)
		if err == nil {
			backoff = initialBackoff
			continue
		}

		log.Printf("Error publishing to %s, backing off %s: %v", b.publisher, backoff, err)
		select {
		case <-time.After(backoff):
		case <-b.quit:
		}
		backoff = backoff * 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// Publish implements Publisher
func (b *BackgroundPublisher) Publish(buf *bytes.Buffer) error {
	select {
	case b.reports <- buf:
	default:
	}
	return nil
}

// Stop implements Publisher
func (b *BackgroundPublisher) Stop() {
	close(b.reports)
	close(b.quit)
	b.publisher.Stop()
}
