package process

import (
	"bytes"
	"errors"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const (
	filesCacheLen        = 2048
	filesCacheExpiration = 60 * time.Second
)

var (
	errInvalidCacheEntry = errors.New("invalid cache entry")
)

type filesCacheEntry struct {
	file     File
	contents *bytes.Buffer
	creation time.Time
}

func (fce *filesCacheEntry) Pinned() bool { return fce.contents != nil }
func (fce *filesCacheEntry) ReadInto(buf *bytes.Buffer) error {
	if fce.Pinned() { // use the saved contents
		buf.Write(fce.contents.Bytes())
		return nil
	}
	return fce.file.ReadInto(buf)
}

// A cache for /proc files
// The cache can hold either file handles or file contents ("pinned" content)
// This cache is not optimized for being used concurrently (it is goroutines-safe though)
type filesCache struct {
	sync.Mutex
	*lru.Cache
	proc Dir
	quit chan struct{}
}

func newFilesCache(proc Dir) *filesCache {
	evictionFunc := func(key, value interface{}) {
		entry, ok := value.(filesCacheEntry)
		if ok && entry.file != nil {
			entry.file.Close()
		}
	}

	cache, err := lru.NewWithEvict(filesCacheLen, evictionFunc)
	if err != nil {
		panic("could not create cache")
	}
	fc := filesCache{
		Cache: cache,
		proc:  proc,
		quit:  make(chan struct{}),
	}
	go fc.loop()

	return &fc
}

// Read a "/proc" file, identified as a sub-path (eg "1134/comm"), into a buffer
// When the "pin" argument is true, we will cache the file _content_.
func (fc *filesCache) ReadInto(filename string, buf *bytes.Buffer, pin bool) error {
	fc.Lock()
	defer fc.Unlock()

	e, ok := fc.Cache.Get(filename)
	if !ok {
		f, err := fc.proc.Open(filename)
		if err != nil {
			return err
		}
		entry := filesCacheEntry{
			creation: time.Now(),
		}
		b := bytes.Buffer{}
		if err := f.ReadInto(&b); err != nil {
			return err
		}
		if pin {
			entry.contents = &b // save the file contents
			f.Close()           // and close the file
		} else {
			entry.file = f
		}
		buf.Write(b.Bytes())
		fc.Cache.Add(filename, entry)
		return nil
	}

	entry := e.(filesCacheEntry)
	return entry.ReadInto(buf)
}

// Close closes all the handles in the cache
func (fc *filesCache) Close() error {
	fc.Lock()
	defer fc.Unlock()

	fc.Cache.Purge()
	close(fc.quit)
	return nil
}

// Go through all the entries in the cache and force the removal of
// the ones that were created too long ago
func (fc *filesCache) loop() {
	purge := time.Tick(filesCacheExpiration)

	for {
		select {
		case <-purge:
			fc.Lock()
			timeLimit := time.Now().Add(-filesCacheExpiration)
			for _, key := range fc.Cache.Keys() {
				e, _ := fc.Cache.Peek(key)
				if entry, ok := e.(filesCacheEntry); ok && entry.creation.Before(timeLimit) {
					fc.Cache.Remove(key)
					// TODO: for pinned entries, we could skip the removal if
					// the path still exists (and PIDs have not wrapped around)
				}
			}
			fc.Unlock()

		case <-fc.quit:
			return
		}
	}
}
