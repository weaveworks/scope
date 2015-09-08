package proc

import (
	"bytes"
	"io/ioutil"
	"os"
	"sync"
)

// ProcDir is the '/proc' directory and the associated ops for
// reading subdirs or files.
type ProcDir interface {
	Root() string                             // proc directory
	ReadDir(string) ([]os.FileInfo, error)    // read a subdirectory in the "root"
	ReadFile(string) ([]byte, error)          // read a file in the "root"
	ReadFileInto(string, *bytes.Buffer) error // read a file in the "root" in a buffer
}

type OSProcDir struct{ Dir string }

func (dp OSProcDir) Root() string                            { return dp.Dir }
func (dp OSProcDir) ReadDir(s string) ([]os.FileInfo, error) { return ioutil.ReadDir(s) }
func (dp OSProcDir) ReadFile(s string) ([]byte, error)       { return ioutil.ReadFile(s) }
func (dp OSProcDir) ReadFileInto(filename string, buf *bytes.Buffer) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	_, err = buf.ReadFrom(f)
	f.Close()
	return err
}

// DefaultProcDir is the default '/proc' directory
var DefaultProcDir = OSProcDir{Dir: "/proc"}

// Process represents a single process.
type Process struct {
	PID, PPID int
	Comm      string
	Cmdline   string
	Threads   int
	Inodes    []uint64
}

// ProcReader is something that reads the /proc directory and
// returns some info like processes and connections
type ProcReader interface {
	// Processes walks through the processes
	Processes(func(Process)) error
	// Connections walks through the connections
	Connections(bool, func(Connection)) error
}

// CachingProcReader is a '/proc' reader than caches a copy of the output from another
// '/proc' reader, and then allows other concurrent readers to Walk that copy.
type CachingProcReader struct {
	procsCache   []Process
	connsCache   []Connection
	source       ProcReader
	includeProcs bool
	sync.RWMutex
}

// NewCachingProcReader returns a new CachingProcReader
func NewCachingProcReader(source ProcReader, includeProcs bool) *CachingProcReader {
	return &CachingProcReader{source: source, includeProcs: includeProcs}
}

// Processes walks a cached copy of process list
func (c *CachingProcReader) Processes(f func(Process)) error {
	c.RLock()
	defer c.RUnlock()

	for _, p := range c.procsCache {
		f(p)
	}
	return nil
}

// Connections walks a cached copy of the connections list
// Note: specifying 'includeProcs' has no effect here, as the cached copy
func (c *CachingProcReader) Connections(_ bool, f func(Connection)) error {
	c.RLock()
	defer c.RUnlock()

	for _, c := range c.connsCache {
		f(c)
	}
	return nil

}

// Update updates the cached copy of the processes and connections lists
func (c *CachingProcReader) Tick() error {
	newProcsCache := []Process{}
	newConnsCache := []Connection{}

	if err := c.source.Processes(func(p Process) {
		newProcsCache = append(newProcsCache, p)
	}); err != nil {
		return err
	}

	if err := c.source.Connections(c.includeProcs, func(conn Connection) {
		newConnsCache = append(newConnsCache, conn.Copy())
	}); err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()
	c.procsCache = newProcsCache
	c.connsCache = newConnsCache

	return nil
}
