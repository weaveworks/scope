package proc

import (
	"bytes"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type procReader struct {
	proc ProcDir
}

// NewProcReader creates a new /proc reader.
func NewProcReader(proc ProcDir) *procReader {
	return &procReader{proc}
}

// Processes walks the /proc directory and marshalls the files into
// instances of Process, which it then passes one-by-one to the
// supplied function. Processes() is only made public so that is
// can be tested.
func (w *procReader) Processes(f func(Process)) error {
	dirEntries, err := w.proc.ReadDir(w.proc.Root())
	if err != nil {
		return err
	}

	for _, dirEntry := range dirEntries {
		filename := dirEntry.Name()
		pid, err := strconv.Atoi(filename)
		if err != nil {
			continue
		}

		stat, err := w.proc.ReadFile(path.Join(w.proc.Root(), filename, "stat"))
		if err != nil {
			continue
		}
		splits := strings.Fields(string(stat))
		ppid, err := strconv.Atoi(splits[3])
		if err != nil {
			return err
		}

		threads, err := strconv.Atoi(splits[19])
		if err != nil {
			return err
		}

		cmdline := ""
		if cmdlineBuf, err := w.proc.ReadFile(path.Join(w.proc.Root(), filename, "cmdline")); err == nil {
			cmdlineBuf = bytes.Replace(cmdlineBuf, []byte{'\000'}, []byte{' '}, -1)
			cmdline = string(cmdlineBuf)
		}

		comm := "(unknown)"
		if commBuf, err := w.proc.ReadFile(path.Join(w.proc.Root(), filename, "comm")); err == nil {
			comm = strings.TrimSpace(string(commBuf))
		}

		fdBase := path.Join(w.proc.Root(), strconv.Itoa(pid), "fd")
		fdNames, err := w.proc.ReadDir(fdBase)
		if err != nil {
			return err
		}

		inodes := []uint64{}
		for _, fdName := range fdNames {
			var fdStat syscall.Stat_t
			// Direct use of syscall.Stat() to save garbage.
			fdPath := path.Join(fdBase, fdName.Name())
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

func (w *procReader) Connections(withProcs bool, f func(Connection)) error {
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

	w.proc.ReadFileInto(path.Join(w.proc.Root(), "net", "tcp"), buf)
	w.proc.ReadFileInto(path.Join(w.proc.Root(), "net", "tcp6"), buf)

	pn := NewProcNet(buf.Bytes(), tcpEstablished)
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
