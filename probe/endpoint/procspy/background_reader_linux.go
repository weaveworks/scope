package procspy

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/weaveworks/scope/probe/process"
)

const (
	initialRateLimit = 50 * time.Millisecond  // Read 20 * fdBlockSize file descriptors (/proc/PID/fd/*) per namespace per second
	maxRateLimit     = 250 * time.Millisecond // Read at least 4 * fdBlockSize file descriptors per namespace per second
	fdBlockSize      = 100
	targetWalkTime   = 10 * time.Second // Aim at walking all files in 10 seconds
)

type backgroundReader struct {
	walker       process.Walker
	mtx          sync.Mutex
	running      bool
	pleaseStop   bool
	walkingBuf   *bytes.Buffer
	readyBuf     *bytes.Buffer
	readySockets map[uint64]*Proc
}

// starts a rate-limited background goroutine to read the expensive files from
// proc.
func newBackgroundReader(walker process.Walker) *backgroundReader {
	br := &backgroundReader{
		walker:     walker,
		walkingBuf: bytes.NewBuffer(make([]byte, 0, 5000)),
		readyBuf:   bytes.NewBuffer(make([]byte, 0, 5000)),
	}
	return br
}

func (br *backgroundReader) start() error {
	br.mtx.Lock()
	defer br.mtx.Unlock()
	if br.running {
		return fmt.Errorf("background reader already running")
	}
	br.running = true
	go br.loop()
	return nil
}

func (br *backgroundReader) stop() error {
	br.mtx.Lock()
	defer br.mtx.Unlock()
	if !br.running {
		return fmt.Errorf("background reader already not running")
	}
	br.pleaseStop = true
	return nil
}

func (br *backgroundReader) loop() {
	const (
		maxRateLimitF   = float64(maxRateLimit)
		targetWalkTimeF = float64(targetWalkTime)
	)

	rateLimit := initialRateLimit
	ticker := time.NewTicker(rateLimit)
	for {
		start := time.Now()
		sockets, err := walkProcPid(br.walkingBuf, br.walker, ticker.C, fdBlockSize)
		if err != nil {
			log.Printf("background reader: error walking /proc: %s\n", err)
			continue
		}

		br.mtx.Lock()

		// Should we stop?
		if br.pleaseStop {
			br.pleaseStop = false
			br.running = false
			ticker.Stop()
			br.mtx.Unlock()
			return
		}

		// Swap buffers
		br.readyBuf, br.walkingBuf = br.walkingBuf, br.readyBuf
		br.readySockets = sockets

		br.mtx.Unlock()

		walkTime := time.Now().Sub(start)
		walkTimeF := float64(walkTime)

		log.Printf("debug: background reader: full pass took %s\n", walkTime)
		if walkTimeF/targetWalkTimeF > 1.5 {
			log.Printf(
				"warn: background reader: full pass took %s: 50%% more than expected (%s)\n",
				walkTime,
				targetWalkTime,
			)
		}

		// Adjust rate limit to more-accurately meet the target walk time in next iteration
		scaledRateLimit := targetWalkTimeF / walkTimeF * float64(rateLimit)
		rateLimit = time.Duration(math.Min(scaledRateLimit, maxRateLimitF))
		log.Printf("debug: background reader: new rate limit %s\n", rateLimit)

		ticker.Stop()
		ticker = time.NewTicker(rateLimit)

		br.walkingBuf.Reset()

		// Sleep during spare time
		time.Sleep(targetWalkTime - walkTime)
	}
}

func (br *backgroundReader) getWalkedProcPid(buf *bytes.Buffer) map[uint64]*Proc {
	br.mtx.Lock()
	defer br.mtx.Unlock()

	reader := bytes.NewReader(br.readyBuf.Bytes())
	buf.ReadFrom(reader)

	return br.readySockets
}
