package procspy

import (
	"bytes"
	"sync"
	"time"

	//"github.com/armon/go-metrics"
	"github.com/coocood/freecache"

	"github.com/weaveworks/scope/common/fs"
)

const (
	timeout   = 60                     // keep contents for 60 seconds
	size      = 10 * 1024 * 1024       // keep upto 10MB worth
	ratelimit = 100 * time.Millisecond // read 10 files per second
)

type backgroundReader struct {
	mtx  sync.Mutex
	cond *sync.Cond

	queue      []string            // sorted list of files to fetch
	queueIndex map[string]struct{} // entry here indicates file is already queued for fetching
	cache      *freecache.Cache
}

// StartBackgroundReader starts a ratelimited background goroutine to
// read the expensive files from proc.
func StartBackgroundReader() {
	br := &backgroundReader{
		queueIndex: map[string]struct{}{},
		cache:      freecache.NewCache(size),
	}
	br.cond = sync.NewCond(&br.mtx)
	go br.loop()
	readFile = br.readFile
}

func (br *backgroundReader) next() string {
	br.mtx.Lock()
	defer br.mtx.Unlock()
	for len(br.queue) == 0 {
		br.cond.Wait()
	}
	filename := br.queue[0]
	br.queue = br.queue[1:]
	delete(br.queueIndex, filename)
	return filename
}

func (br *backgroundReader) enqueue(filename string) {
	br.mtx.Lock()
	defer br.mtx.Unlock()
	if _, ok := br.queueIndex[filename]; !ok {
		br.queue = append(br.queue, filename)
		br.queueIndex[filename] = struct{}{}
		br.cond.Broadcast()
	}
}

func (br *backgroundReader) readFileIntoCache(filename string) error {
	contents, err := fs.ReadFile(filename)
	if err != nil {
		return err
	}
	br.cache.Set([]byte(filename), contents, timeout)
	return nil
}

func (br *backgroundReader) loop() {
	ticker := time.Tick(ratelimit)
	for {
		err := br.readFileIntoCache(br.next())
		// Only rate limit if we succesfully read a file
		if err == nil {
			<-ticker
		}
	}
}

func (br *backgroundReader) readFile(filename string, buf *bytes.Buffer) (int64, error) {
	// We always schedule the filename for reading, as on quiet systems
	// we want to have a fresh as possible resuls
	br.enqueue(filename)

	v, err := br.cache.Get([]byte(filename))
	if err != nil {
		return 0, nil
	}
	n, err := buf.Write(v)
	return int64(n), err
}
