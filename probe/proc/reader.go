package proc

import (
	"bytes"
	"os"
	"path"
	"sync"
)

// File is a file in the "/proc" directory
type File interface {
	ReadInto(buf *bytes.Buffer) error
	Close() error
}

// OSFile is a native file
type OSFile struct{ *os.File }

// ReadInto reads the whole file into a buffer
func (of *OSFile) ReadInto(buf *bytes.Buffer) error {
	if _, err := of.File.Seek(0, 0); err != nil {
		return err
	}
	_, err := buf.ReadFrom(of.File)
	return err
}

// Dir is the '/proc' directory and the associated ops for
// reading subdirs or files.
type Dir interface {
	Root() string                          // the "/proc" directory
	Open(s string) (File, error)           // open a file in the "/proc" dir
	ReadDirNames(string) ([]string, error) // list a subdirectory in the "/proc"
}

// OSDir is a OS "/proc" firectory
type OSDir struct{ Dir string }

// Root returns the "/proc" top dir
func (dp OSDir) Root() string {
	return dp.Dir
}

// Open returns a ProcFile
func (dp OSDir) Open(s string) (File, error) {
	h, err := os.Open(path.Join(dp.Root(), s))
	if err != nil {
		return nil, err
	}
	return &OSFile{h}, nil
}

// ReadDirNames reads all the directory entries
func (dp OSDir) ReadDirNames(s string) ([]string, error) {
	f, err := os.Open(s)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	return list, nil
}

// DefaultProcDir is the default '/proc' directory
var DefaultProcDir = OSDir{Dir: "/proc"}

// Process represents a single process.
type Process struct {
	PID, PPID int
	Comm      string
	Cmdline   string
	Threads   int
	Inodes    []uint64
}

// Reader is something that reads the /proc directory and
// returns some info like processes and connections
type Reader interface {
	// Processes walks through the processes
	Processes(func(Process)) error
	// Connections walks through the connections
	Connections(bool, func(Connection)) error
	// Close closes the "/proc" reader
	Close() error
}

// CachingProcReader is a '/proc' reader than caches a copy of the output from another
// '/proc' reader, and then allows other concurrent readers to Walk that copy.
type CachingProcReader struct {
	procsCache   []Process
	connsCache   []Connection
	source       Reader
	includeProcs bool
	sync.RWMutex
}

// NewCachingProcReader returns a new CachingProcReader
func NewCachingProcReader(source Reader, includeProcs bool) *CachingProcReader {
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

// Close closes the "/proc" reader
func (c *CachingProcReader) Close() error {
	return c.source.Close()
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
