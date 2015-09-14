package proc

import (
	"bytes"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bluele/gcache"
)

const (
	filesCacheLen        = 512
	filesCacheExpiration = 60 * time.Second
)

var tcpFiles = []string{
	"net/tcp",
	"net/tcp6",
}

// A cache for files handles
type filesCache struct {
	handles gcache.Cache
}

type filesCacheEntry struct {
	sync.RWMutex
	File
}

func newFilesCache(proc Dir) *filesCache {
	loadFunc := func(fileName interface{}) (interface{}, error) {
		f, err := proc.Open(fileName.(string))
		if err != nil {
			return nil, err
		}
		return filesCacheEntry{File: f}, nil
	}
	evictionFunc := func(key, value interface{}) {
		value.(filesCacheEntry).Close()
	}
	return &filesCache{
		handles: gcache.New(filesCacheLen).LoaderFunc(loadFunc).EvictedFunc(evictionFunc).Expiration(filesCacheExpiration).ARC().Build(),
	}
}

// Read a "/proc" file, identified as a file in a subdir (eg "1134/comm"), into a buffer
func (fc *filesCache) ReadInto(filename string, buf *bytes.Buffer) error {
	// we could use a lock here, but this is only used from Processes()/Connections(),
	// and they are always invoked sequentially...
	h, err := fc.handles.Get(filename)
	if err != nil {
		return err
	}
	handle := h.(filesCacheEntry)
	handle.Lock()
	defer handle.Unlock()
	return handle.ReadInto(buf)
}

// Close closes all the handles in the cache
func (fc *filesCache) Close() error {
	for _, key := range fc.handles.Keys() {
		fc.handles.Remove(key)
	}
	return nil
}

// the Linux "/proc" reader
type reader struct {
	proc    Dir
	handles *filesCache
}

// NewReader creates a new /proc reader.
func NewReader(proc Dir) Reader {
	return &reader{
		proc:    proc,
		handles: newFilesCache(proc),
	}
}

// Close closes the Linux "/proc" reader
func (r *reader) Close() error {
	return r.handles.Close()
}

// Processes walks the /proc directory and marshalls the files into
// instances of Process, which it then passes one-by-one to the
// supplied function. Processes() is only made public so that is
// can be tested.
func (r *reader) Processes(f func(Process)) error {
	dirEntries, err := r.proc.ReadDirNames(r.proc.Root())
	if err != nil {
		return err
	}

	var fdStat syscall.Stat_t
	buf := bytes.Buffer{}

	for _, subdir := range dirEntries {
		readIntoBuffer := func(filename string) error {
			buf.Reset()
			return r.handles.ReadInto(path.Join(subdir, filename), &buf)
		}

		pid, err := strconv.Atoi(subdir)
		if err != nil {
			continue
		}

		if readIntoBuffer("stat") != nil {
			continue
		}
		splits := strings.Fields(buf.String())
		ppid, err := strconv.Atoi(splits[3])
		if err != nil {
			return err
		}

		threads, err := strconv.Atoi(splits[19])
		if err != nil {
			return err
		}

		cmdline := ""
		if readIntoBuffer("cmdline") == nil {
			cmdlineBuf := bytes.Replace(buf.Bytes(), []byte{'\000'}, []byte{' '}, -1)
			cmdline = string(cmdlineBuf)
		}

		comm := "(unknown)"
		if readIntoBuffer("comm") == nil {
			comm = strings.TrimSpace(buf.String())
		}

		fdBase := path.Join(r.proc.Root(), strconv.Itoa(pid), "fd")
		fdNames, err := r.proc.ReadDirNames(fdBase)
		if err != nil {
			return err
		}

		inodes := []uint64{}
		for _, fdName := range fdNames {
			// Direct use of syscall.Stat() to save garbage.
			fdPath := path.Join(fdBase, fdName)
			err = syscall.Stat(fdPath, &fdStat)
			if err == nil && (fdStat.Mode&syscall.S_IFMT == syscall.S_IFSOCK) { // We want sockets only.
				inodes = append(inodes, fdStat.Ino)
			}
		}

		f(Process{
			PID:     pid,
			PPID:    ppid,
			Comm:    comm,
			Cmdline: cmdline,
			Threads: threads,
			Inodes:  inodes,
		})
	}

	return nil
}

// Connections walks through all the connections in the "/proc"
func (r *reader) Connections(withProcs bool, f func(Connection)) error {
	// create a map of inode->Process
	procs := make(map[uint64]Process)
	if withProcs {
		r.Processes(func(p Process) {
			for _, inode := range p.Inodes {
				procs[inode] = p
			}
		})
	}

	buf := bytes.Buffer{}
	for _, tcpFile := range tcpFiles {
		err := r.handles.ReadInto(tcpFile, &buf)
		if err != nil {
			return err
		}
	}

	pn := newNetReader(buf.Bytes(), tcpEstablished)
	for {
		conn := pn.Next()
		if conn == nil {
			break // Done!
		}
		if proc, ok := procs[conn.inode]; ok {
			conn.Process = proc
		}
		f(*conn)
	}
	return nil
}
