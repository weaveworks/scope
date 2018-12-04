package process

import "sync"

// Process represents a single process.
type Process struct {
	PID, PPID         int
	Name              string
	Cmdline           string
	Threads           int
	Jiffies           uint64
	RSSBytes          uint64
	RSSBytesLimit     uint64
	OpenFilesCount    int
	OpenFilesLimit    uint64
	IsWaitingInAccept bool
}

// Walker is something that walks the /proc directory
type Walker interface {
	Walk(func(Process, Process)) error
}

// CachingWalker is a walker than caches a copy of the output from another
// Walker, and then allows other concurrent readers to Walk that copy.
type CachingWalker struct {
	cache         map[int]Process
	previousByPID map[int]Process
	cacheLock     sync.RWMutex
	source        Walker
}

// NewCachingWalker returns a new CachingWalker
func NewCachingWalker(source Walker) *CachingWalker {
	return &CachingWalker{source: source}
}

// Name of this ticker, for metrics gathering
func (*CachingWalker) Name() string { return "Process" }

// Walk walks a cached copy of process list
func (c *CachingWalker) Walk(f func(Process, Process)) error {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()

	for _, p := range c.cache {
		f(p, c.previousByPID[p.PID])
	}
	return nil
}

// Tick updates cached copy of process list
func (c *CachingWalker) Tick() error {
	newCache := map[int]Process{}
	err := c.source.Walk(func(p, _ Process) {
		newCache[p.PID] = p
	})
	if err != nil {
		return err
	}

	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()
	c.previousByPID = c.cache
	c.cache = newCache
	return nil
}
