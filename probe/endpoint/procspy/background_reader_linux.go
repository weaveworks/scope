package procspy

import (
	"bytes"
	"io"
	"math"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/scope/probe/process"
)

const (
	initialRateLimitPeriod = 50 * time.Millisecond  // Read 20 * fdBlockSize file descriptors (/proc/PID/fd/*) per namespace per second
	maxRateLimitPeriod     = 250 * time.Millisecond // Read at least 4 * fdBlockSize file descriptors per namespace per second
	fdBlockSize            = uint64(300)            // Maximum number of /proc/PID/fd/* files to stat per rate-limit period
	// (as a rule of thumb going through each block should be more expensive than reading /proc/PID/tcp{,6})
	targetWalkTime = 10 * time.Second // Aim at walking all files in 10 seconds
)

type backgroundReader struct {
	stopc         chan struct{}
	mtx           sync.Mutex
	latestBuf     *bytes.Buffer
	latestSockets map[uint64]*Proc
}

// starts a rate-limited background goroutine to read the expensive files from
// proc.
func newBackgroundReader(walker process.Walker) *backgroundReader {
	br := &backgroundReader{
		latestBuf: bytes.NewBuffer(make([]byte, 0, 5000)),
		stopc:     make(chan struct{}),
	}
	go br.loop(walker)
	return br
}

func (br *backgroundReader) stop() {
	close(br.stopc)
}

func (br *backgroundReader) getWalkedProcPid(buf *bytes.Buffer) (map[uint64]*Proc, error) {
	br.mtx.Lock()
	defer br.mtx.Unlock()

	_, err := io.Copy(buf, br.latestBuf)

	return br.latestSockets, err
}

func (br *backgroundReader) loop(walker process.Walker) {
	var (
		begin           time.Time                      // when we started the last performWalk
		tickc           = time.After(time.Millisecond) // fire immediately
		walkc           chan map[uint64]*Proc          // initially nil, i.e. off
		walkBuf         = bytes.NewBuffer(make([]byte, 0, 5000))
		rateLimitPeriod = initialRateLimitPeriod
		nextInterval    time.Duration
		ticker          = time.NewTicker(rateLimitPeriod)
		pWalker         = newPidWalker(walker, ticker.C, fdBlockSize)
	)

	for {
		select {
		case <-tickc:
			tickc = nil                             // turn off until the next loop
			walkc = make(chan map[uint64]*Proc, 1)  // turn on (need buffered so we don't leak performWalk)
			begin = time.Now()                      // reset counter
			go performWalk(pWalker, walkBuf, walkc) // do work

		case sockets := <-walkc:
			// Swap buffers
			br.mtx.Lock()
			br.latestBuf, walkBuf = walkBuf, br.latestBuf
			br.latestSockets = sockets
			br.mtx.Unlock()
			walkBuf.Reset()

			// Schedule next walk and adjust rate limit
			walkTime := time.Since(begin)
			rateLimitPeriod, nextInterval = scheduleNextWalk(rateLimitPeriod, walkTime)
			ticker.Stop()
			ticker = time.NewTicker(rateLimitPeriod)
			pWalker.ticker = ticker.C

			walkc = nil                      // turn off until the next loop
			tickc = time.After(nextInterval) // turn on

		case <-br.stopc:
			pWalker.stop()
			ticker.Stop()
			return // abort
		}
	}
}

// Adjust rate limit for next walk and calculate how long to wait until it should be started
func scheduleNextWalk(rateLimitPeriod time.Duration, took time.Duration) (time.Duration, time.Duration) {

	log.Debugf("background /proc reader: full pass took %s", took)
	if float64(took)/float64(targetWalkTime) > 1.5 {
		log.Warnf(
			"background /proc reader: full pass took %s: 50%% more than expected (%s)",
			took,
			targetWalkTime,
		)
	}

	// Adjust rate limit to more-accurately meet the target walk time in next iteration
	scaledRateLimitPeriod := float64(targetWalkTime) / float64(took) * float64(rateLimitPeriod)
	rateLimitPeriod = time.Duration(math.Min(scaledRateLimitPeriod, float64(maxRateLimitPeriod)))

	log.Debugf("background /proc reader: new rate limit %s", rateLimitPeriod)

	return rateLimitPeriod, targetWalkTime - took
}

func performWalk(w pidWalker, buf *bytes.Buffer, c chan<- map[uint64]*Proc) {
	sockets, err := w.walk(buf)
	if err != nil {
		log.Errorf("background /proc reader: error walking /proc: %s", err)
		buf.Reset()
		c <- nil
		return
	}
	c <- sockets
}
