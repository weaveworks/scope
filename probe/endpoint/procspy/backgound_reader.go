package procspy

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/weaveworks/scope/probe/process"
)

const (
	ratelimit = 20 * time.Millisecond // read 50 namespaces per second
)

type backgroundReader struct {
	walker       process.Walker
	mtx          sync.Mutex
	walkingBuf   *bytes.Buffer
	readyBuf     *bytes.Buffer
	readySockets map[uint64]*Proc
}

// HACK: Pretty ugly singleton interface (particularly the part part of passing
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
	namespaceTicker := time.Tick(ratelimit)
	for {
		sockets, err := walkProcPid(br.walkingBuf, br.walker, namespaceTicker)
		if err != nil {
			fmt.Printf("background reader: error reading walking /proc: %s\n", err)
			continue
		}

		// Swap buffers
		br.mtx.Lock()
		br.readyBuf, br.walkingBuf = br.walkingBuf, br.readyBuf
		br.readySockets = sockets
		br.mtx.Unlock()

		br.walkingBuf.Reset()
	}
}

func (br *backgroundReader) getWalkedProcPid(buf *bytes.Buffer) map[uint64]*Proc {
	br.mtx.Lock()
	defer br.mtx.Unlock()

	reader := bytes.NewReader(br.readyBuf.Bytes())
	buf.ReadFrom(reader)

	return br.readySockets
}
