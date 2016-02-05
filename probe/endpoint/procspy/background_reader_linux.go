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
	walkingBuf   *bytes.Buffer
	readyBuf     *bytes.Buffer
	readySockets map[uint64]*Proc
}

// HACK: Pretty ugly singleton interface (particularly the part of passing
// the walker to StartBackgroundReader() and ignoring it in in Connections() )
// experimenting with this for now.
var singleton *backgroundReader

func getBackgroundReader() (*backgroundReader, error) {
	var err error
	if singleton == nil {
		err = fmt.Errorf("background reader hasn't yet been started")
	}
	return singleton, err
}

// StartBackgroundReader starts a ratelimited background goroutine to
// read the expensive files from proc.
func StartBackgroundReader(walker process.Walker) {
	if singleton != nil {
		return
	}
	singleton = &backgroundReader{
		walker:     walker,
		walkingBuf: bytes.NewBuffer(make([]byte, 0, 5000)),
		readyBuf:   bytes.NewBuffer(make([]byte, 0, 5000)),
	}
	go singleton.loop()
}

func (br *backgroundReader) loop() {
	const (
		maxRateLimitF   = float64(maxRateLimit)
		targetWalkTimeF = float64(targetWalkTime)
	)

	rateLimit := initialRateLimit
	ticker := time.Tick(rateLimit)
	for {
		start := time.Now()
		sockets, err := walkProcPid(br.walkingBuf, br.walker, ticker, fdBlockSize)
		if err != nil {
			log.Printf("background reader: error walking /proc: %s\n", err)
			continue
		}
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

		ticker = time.Tick(rateLimit)

		// Swap buffers
		br.mtx.Lock()
		br.readyBuf, br.walkingBuf = br.walkingBuf, br.readyBuf
		br.readySockets = sockets
		br.mtx.Unlock()

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
