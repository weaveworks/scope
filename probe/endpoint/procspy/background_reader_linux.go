package procspy

import (
	"bytes"
	"io"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/scope/probe/process"
)

const (
	initialRateLimitPeriod = 50 * time.Millisecond  // Read 20 * fdBlockSize file descriptors (/proc/PID/fd/*) per namespace per second
	maxRateLimitPeriod     = 500 * time.Millisecond // Read at least 2 * fdBlockSize file descriptors per namespace per second
	minRateLimitPeriod     = initialRateLimitPeriod
	fdBlockSize            = uint64(300) // Maximum number of /proc/PID/fd/* files to stat per rate-limit period
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
		stopc:         make(chan struct{}),
		latestSockets: map[uint64]*Proc{},
	}
	go br.loop(walker)
	return br
}

func newForegroundReader(walker process.Walker) *backgroundReader {
	br := &backgroundReader{
		stopc:         make(chan struct{}),
		latestSockets: map[uint64]*Proc{},
	}
	var (
		walkc   = make(chan walkResult)
		ticker  = time.NewTicker(time.Millisecond) // fire every millisecond
		pWalker = newPidWalker(walker, ticker.C, fdBlockSize)
	)

	go performWalk(pWalker, walkc)

	result := <-walkc
	br.mtx.Lock()
	br.latestBuf = result.buf
	br.latestSockets = result.sockets
	br.mtx.Unlock()

	return br
}

func (br *backgroundReader) stop() {
	close(br.stopc)
}

func (br *backgroundReader) getWalkedProcPid(buf *bytes.Buffer) (map[uint64]*Proc, error) {
	br.mtx.Lock()
	defer br.mtx.Unlock()

	// Don't access latestBuf directly but create a reader. In this way,
	// the buffer will not be empty in the next call of getWalkedProcPid
	// and it can be copied again.
	_, err := io.Copy(buf, bytes.NewReader(br.latestBuf.Bytes()))

	return br.latestSockets, err
}

type walkResult struct {
	buf     *bytes.Buffer
	sockets map[uint64]*Proc
}

func performWalk(w pidWalker, c chan<- walkResult) {
	var (
		err    error
		result = walkResult{
			// TODO should we increase buf size?
			buf: bytes.NewBuffer(make([]byte, 0, 5000)),
		}
	)

	result.sockets, err = w.walk(result.buf)
	if err != nil {
		log.Errorf("background /proc reader: error walking /proc: %s", err)
		result.buf.Reset()
		result.sockets = nil
	}
	c <- result
}

func (br *backgroundReader) loop(walker process.Walker) {
	var (
		begin           time.Time                      // when we started the last performWalk
		tickc           = time.After(time.Millisecond) // fire immediately
		walkc           chan walkResult                // initially nil, i.e. off
		rateLimitPeriod = initialRateLimitPeriod
		restInterval    time.Duration
		ticker          = time.NewTicker(rateLimitPeriod)
		pWalker         = newPidWalker(walker, ticker.C, fdBlockSize)
	)

	for {
		select {
		case <-tickc:
			tickc = nil                      // turn off until the next loop
			walkc = make(chan walkResult, 1) // turn on (need buffered so we don't leak performWalk)
			begin = time.Now()               // reset counter
			go performWalk(pWalker, walkc)   // do work

		case result := <-walkc:
			// Expose results
			br.mtx.Lock()
			br.latestBuf = result.buf
			br.latestSockets = result.sockets
			br.mtx.Unlock()

			// Schedule next walk and adjust its rate limit
			walkTime := time.Since(begin)
			rateLimitPeriod, restInterval = scheduleNextWalk(rateLimitPeriod, walkTime)
			ticker.Stop()
			ticker = time.NewTicker(rateLimitPeriod)
			pWalker.tickc = ticker.C

			walkc = nil                      // turn off until the next loop
			tickc = time.After(restInterval) // turn on

		case <-br.stopc:
			pWalker.stop()
			ticker.Stop()
			return // abort
		}
	}
}

// Adjust rate limit for next walk and calculate when it should be started
func scheduleNextWalk(rateLimitPeriod time.Duration, took time.Duration) (newRateLimitPeriod time.Duration, restInterval time.Duration) {
	log.Debugf("background /proc reader: full pass took %s", took)
	if float64(took)/float64(targetWalkTime) > 1.5 {
		log.Warnf(
			"background /proc reader: full pass took %s: 50%% more than expected (%s)",
			took,
			targetWalkTime,
		)
	}

	// Adjust rate limit to more-accurately meet the target walk time in next iteration
	newRateLimitPeriod = time.Duration(float64(targetWalkTime) / float64(took) * float64(rateLimitPeriod))
	if newRateLimitPeriod > maxRateLimitPeriod {
		newRateLimitPeriod = maxRateLimitPeriod
	} else if newRateLimitPeriod < minRateLimitPeriod {
		newRateLimitPeriod = minRateLimitPeriod
	}
	log.Debugf("background /proc reader: new rate limit period %s", newRateLimitPeriod)

	return newRateLimitPeriod, targetWalkTime - took
}
