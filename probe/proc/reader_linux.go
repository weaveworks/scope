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
// Note: not intended to be used from multiple goroutines
type filesCache struct {
	handles gcache.Cache
}

func newFilesCache(proc Dir) *filesCache {
	loadFunc := func(fileName interface{}) (interface{}, error) {
		return proc.Open(fileName.(string))
	}
	evictionFunc := func(key, value interface{}) {
		value.(File).Close()
	}
	return &filesCache{
		handles: gcache.New(filesCacheLen).LoaderFunc(loadFunc).EvictedFunc(evictionFunc).Expiration(filesCacheExpiration).ARC().Build(),
	}
}

// Read a "/proc" file, identified as a subdir (eg "1134/comm"), into a buffer
// Note: this is not goroutine-safe: two goroutines getting and reading from
//       the same handle can obtain some unexpected contents...
func (fc *filesCache) ReadInto(filename string, buf *bytes.Buffer) error {
	// we could use a lock here, but this is only used from Processes()/Connections(),
	// and they are always invoked sequentially...
	h, err := fc.handles.Get(filename)
	if err != nil {
		return err
	}
	return h.(File).ReadInto(buf)
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
func (w *reader) Close() error {
	return w.handles.Close()
}

// Processes walks the /proc directory and marshalls the files into
// instances of Process, which it then passes one-by-one to the
// supplied function. Processes() is only made public so that is
// can be tested.
func (w *reader) Processes(f func(Process)) error {
	dirEntries, err := w.proc.ReadDirNames(w.proc.Root())
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(make([]byte, 0, 5000))
	readIntoBuffer := func(filename string) error {
		buf.Reset()
		res := w.handles.ReadInto(filename, buf)
		return res
	}

	for _, filename := range dirEntries {
		pid, err := strconv.Atoi(filename)
		if err != nil {
			continue
		}

		if readIntoBuffer(path.Join(filename, "stat")) != nil {
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
		if readIntoBuffer(path.Join(filename, "cmdline")) == nil {
			cmdlineBuf := bytes.Replace(buf.Bytes(), []byte{'\000'}, []byte{' '}, -1)
			cmdline = string(cmdlineBuf)
		}

		comm := "(unknown)"
		if readIntoBuffer(path.Join(filename, "comm")) == nil {
			comm = strings.TrimSpace(buf.String())
		}

		fdBase := path.Join(w.proc.Root(), strconv.Itoa(pid), "fd")
		fdNames, err := w.proc.ReadDirNames(fdBase)
		if err != nil {
			return err
		}

		inodes := []uint64{}
		for _, fdName := range fdNames {
			var fdStat syscall.Stat_t
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

var bufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 5000))
	},
}

// Connections walks through all the connections in the "/proc"
func (w *reader) Connections(withProcs bool, f func(Connection)) error {
	// create a map of inode->Process
	procs := make(map[uint64]Process)
	if withProcs {
		w.Processes(func(p Process) {
			for _, inode := range p.Inodes {
				procs[inode] = p
			}
		})
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	for _, tcpFile := range tcpFiles {
		err := w.handles.ReadInto(tcpFile, buf)
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
