package procspy

// /proc-based implementation.

import (
	"bytes"
	"log"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

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
func walkProcPid(buf *bytes.Buffer, walker process.Walker, namespaceTicker <-chan time.Time) (map[uint64]*Proc, error) {
	var (
		res        = map[uint64]*Proc{}              // map socket inode -> process
		namespaces = map[uint64][]*process.Process{} // map network namespace id -> processes
		statT      syscall.Stat_t
	)

	// Two process passes: One to group processes by namespace and another
	// one to obtain their sockets.
	//
	// The first pass is done to allow inferring the connections on a
	// per-namespace basis in the second pass. This is done to minimize the
	// race condition between reading /net/tcp{,6} of each namespace and
	// /proc/PID/fd/* for the processes living in that namespace.

	walker.Walk(func(p, _ process.Process) {
		dirName := strconv.Itoa(p.PID)

		if err := fs.Lstat(filepath.Join(procRoot, dirName, "/ns/net"), &statT); err != nil {
			return
		}

		procs, ok := namespaces[statT.Ino]
		if ok {
			namespaces[statT.Ino] = append(procs, &p)
		} else {
			namespaces[statT.Ino] = []*process.Process{&p}
		}
	})

	log.Printf("debug: walkProcPid: found %d namespaces\n", len(namespaces))

	for _, procs := range namespaces {

		<-namespaceTicker

		// Read the namespace connections (i.e. read /proc/PID/net/tcp{,6} for
		// any of the processes in the namespace)

		// FIXME: try reading other processes in case this one's gone
		p := procs[0]
		dirName := strconv.Itoa(p.PID)

		read, _ := readFile(filepath.Join(procRoot, dirName, "/net/tcp"), buf)
		read6, _ := readFile(filepath.Join(procRoot, dirName, "/net/tcp6"), buf)

		if read+read6 == 0 {
			// no connections, don't bother reading /fd/*
			continue
		}

		// Get the sockets for all the processes in the namespace
		for _, p := range procs {
			dirName = strconv.Itoa(p.PID)
			fdBase := filepath.Join(procRoot, dirName, "fd")

			fds, err := fs.ReadDirNames(fdBase)
			if err != nil {
				// Process is be gone by now, or we don't have access.
				continue
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
		}
	}

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
