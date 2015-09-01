package process

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"
)

// File is a file in the "/proc" directory
type File interface {
	ReadInto(buf *bytes.Buffer) error
	Close() error
}

type seekableReader interface {
	io.Reader
	io.Closer
	Seek(int64, int) (int64, error)
}

// OSFile is a native file
type OSFile struct{ reader seekableReader }

// ReadInto reads the whole file into a buffer
func (of *OSFile) ReadInto(buf *bytes.Buffer) error {
	if _, err := of.reader.Seek(0, 0); err != nil {
		return err
	}
	_, err := buf.ReadFrom(of.reader)
	return err
}

// Close closes the reader
func (of *OSFile) Close() error {
	return of.reader.Close()
}

// Dir is the '/proc' directory and the associated ops for
// reading subdirs or files.
type Dir interface {
	Root() string                             // the "/proc" directory
	Open(s string) (File, error)              // open a file in the "/proc" dir
	ReadDirNames(string) ([]string, error)    // list a subdirectory in the "/proc"
	Stat(string, bool, *syscall.Stat_t) error // stats/lstat
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
	f, err := os.Open(path.Join(dp.Root(), s))
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

// Stat returns the stat info for a file in the /proc directory
func (dp OSDir) Stat(s string, follow bool, stat *syscall.Stat_t) error {
	f := syscall.Stat
	if follow {
		f = syscall.Lstat
	}
	return f(path.Join(dp.Root(), s), stat)
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

// Copy returns a copy of a process
func (p Process) Copy() Process {
	dup := make([]uint64, len(p.Inodes))
	copy(dup, p.Inodes)
	p.Inodes = dup
	return p
}

// String returns the string repr
func (p Process) String() string {
	return fmt.Sprintf("%s[%d]", p.Comm, p.PID)
}

// ProcsInodes is a map from inodes to processes
type ProcsInodes map[uint64]*Process

// Reader is something that reads the /proc directory and
// returns some info like processes and connections
type Reader interface {
	// Read reads the processes and connections
	Read() error
	// Processes walks through the processes
	Processes(func(Process)) error
	// Connections walks through the connections
	Connections(func(Connection)) error
	// Close closes the "/proc" reader
	Close() error
}
