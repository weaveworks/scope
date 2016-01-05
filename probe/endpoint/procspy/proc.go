package procspy

// /proc-based implementation.

import (
	"bytes"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/weaveworks/scope/common/fs"
	"github.com/weaveworks/scope/probe/process"
)

var (
	procRoot = "/proc"
)

// SetProcRoot sets the location of the proc filesystem.
func SetProcRoot(root string) {
	procRoot = root
}

// walkProcPid walks over all numerical (PID) /proc entries, and sees if their
// ./fd/* files are symlink to sockets. Returns a map from socket ID (inode)
// to PID. Will return an error if /proc isn't there.
func walkProcPid(buf *bytes.Buffer, walker process.Walker) (map[uint64]*Proc, error) {
	var (
		res        = map[uint64]*Proc{}
		namespaces = map[uint64]bool{} // map namespace id -> has connections
		statT      syscall.Stat_t
	)

	walker.Walk(func(p, _ process.Process) {
		dirName := strconv.Itoa(p.PID)
		fdBase := filepath.Join(procRoot, dirName, "fd")

		// Read network namespace, and if we haven't seen it before,
		// read /proc/<pid>/net/tcp
		if err := fs.Lstat(filepath.Join(procRoot, dirName, "/ns/net"), &statT); err != nil {
			return
		}
		hasConns, ok := namespaces[statT.Ino]
		if !ok {
			read, _ := readFile(filepath.Join(procRoot, dirName, "/net/tcp"), buf)
			read6, _ := readFile(filepath.Join(procRoot, dirName, "/net/tcp6"), buf)
			hasConns = read+read6 > 0
			namespaces[statT.Ino] = hasConns
		}
		if !hasConns {
			return
		}

		fds, err := fs.ReadDirNames(fdBase)
		if err != nil {
			// Process is be gone by now, or we don't have access.
			return
		}
		var proc *Proc
		for _, fd := range fds {
			// Direct use of syscall.Stat() to save garbage.
			err = fs.Stat(filepath.Join(fdBase, fd), &statT)
			if err != nil {
				continue
			}

			// We want sockets only.
			if statT.Mode&syscall.S_IFMT != syscall.S_IFSOCK {
				continue
			}
			if proc == nil {
				proc = &Proc{
					PID:  uint(p.PID),
					Name: p.Name,
				}
			}
			res[statT.Ino] = proc
		}
	})

	return res, nil
}

// readFile reads an arbitrary file into a buffer. It's a variable so it can
// be overwritten for benchmarks. That's bad practice and we should change it
// to be a dependency.
var readFile = func(filename string, buf *bytes.Buffer) (int64, error) {
	f, err := fs.Open(filename)
	if err != nil {
		return -1, err
	}
	defer f.Close()
	return buf.ReadFrom(f)
}
