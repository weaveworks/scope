package procspy

// /proc-based implementation.

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/armon/go-metrics"

	"$GITHUB_URI/common/fs"
	"$GITHUB_URI/common/marshal"
	"$GITHUB_URI/probe/process"
)

var (
	procRoot               = "/proc"
	namespaceKey           = []string{"procspy", "namespaces"}
	netNamespacePathSuffix = ""
)

type pidWalker struct {
	walker      process.Walker
	tickc       <-chan time.Time // Rate-limit clock. Sets the pace when traversing namespaces and /proc/PID/fd/* files.
	stopc       chan struct{}    // Abort walk
	fdBlockSize uint64           // Maximum number of /proc/PID/fd/* files to stat() per tick
}

func newPidWalker(walker process.Walker, tickc <-chan time.Time, fdBlockSize uint64) pidWalker {
	w := pidWalker{
		walker:      walker,
		tickc:       tickc,
		fdBlockSize: fdBlockSize,
		stopc:       make(chan struct{}),
	}
	return w
}

// SetProcRoot sets the location of the proc filesystem.
func SetProcRoot(root string) {
	procRoot = root
}

func getKernelVersion() (major, minor int, err error) {
	var u syscall.Utsname
	if err = syscall.Uname(&u); err != nil {
		return
	}

	// Kernel versions are not always a semver, so we have to do minimal parsing.
	release := marshal.FromUtsname(u.Release)
	if n, err := fmt.Sscanf(release, "%d.%d", &major, &minor); err != nil || n != 2 {
		return 0, 0, fmt.Errorf("Malformed version: %s", release)
	}
	return
}

func getNetNamespacePathSuffix() string {
	// With Linux 3.8 or later the network namespace of a process can be
	// determined by the inode of /proc/PID/net/ns.  Before that, Any file
	// under /proc/PID/net/ could be used but it's not documented and may
	// break in newer kernels.
	const (
		post38Path = "ns/net"
		pre38Path  = "net/dev"
	)

	if netNamespacePathSuffix != "" {
		return netNamespacePathSuffix
	}

	major, minor, err := getKernelVersion()
	if err != nil {
		log.Errorf("getNamespacePathSuffix: cannot get kernel version: %s", err)
		netNamespacePathSuffix = post38Path
		return netNamespacePathSuffix
	}

	if major < 3 || (major == 3 && minor < 8) {
		netNamespacePathSuffix = pre38Path
	} else {
		netNamespacePathSuffix = post38Path
	}
	return netNamespacePathSuffix
}

// Read the connections for a group of processes living in the same namespace,
// which are found (identically) in /proc/PID/net/tcp{,6} for any of the
// processes.
func readProcessConnections(buf *bytes.Buffer, namespaceProcs []*process.Process) (bool, error) {
	var (
		errRead  error
		errRead6 error
		read     int64
		read6    int64
	)

	for _, p := range namespaceProcs {
		dirName := strconv.Itoa(p.PID)

		read, errRead = readFile(filepath.Join(procRoot, dirName, "/net/tcp"), buf)
		read6, errRead6 = readFile(filepath.Join(procRoot, dirName, "/net/tcp6"), buf)

		if errRead != nil || errRead6 != nil {
			// try next process
			continue
		}
		return read+read6 > 0, nil
	}

	// would be cool to have an or operation between errors
	if errRead != nil {
		return false, errRead
	}
	if errRead6 != nil {
		return false, errRead6
	}

	return false, nil

}

// walkNamespace does the work of walk for a single namespace
func (w pidWalker) walkNamespace(buf *bytes.Buffer, sockets map[uint64]*Proc, namespaceProcs []*process.Process) error {

	if found, err := readProcessConnections(buf, namespaceProcs); err != nil || !found {
		return err
	}

	var statT syscall.Stat_t
	var fdBlockCount uint64
	for i, p := range namespaceProcs {

		// Get the sockets for all the processes in the namespace
		dirName := strconv.Itoa(p.PID)
		fdBase := filepath.Join(procRoot, dirName, "fd")

		if fdBlockCount > w.fdBlockSize {
			// we surpassed the filedescriptor rate limit
			select {
			case <-w.tickc:
			case <-w.stopc:
				return nil // abort
			}

			fdBlockCount = 0
			// read the connections again to
			// avoid the race between between /net/tcp{,6} and /proc/PID/fd/*
			if found, err := readProcessConnections(buf, namespaceProcs[i:]); err != nil || !found {
				return err
			}
		}

		fds, err := fs.ReadDirNames(fdBase)
		if err != nil {
			// Process is gone by now, or we don't have access.
			continue
		}

		var proc *Proc
		for _, fd := range fds {
			fdBlockCount++

			// Direct use of syscall.Stat() to save garbage.
			err = fs.Stat(filepath.Join(fdBase, fd), &statT)
			if err != nil {
				continue
			}

			// We want sockets only.
			if statT.Mode&syscall.S_IFMT != syscall.S_IFSOCK {
				continue
			}

			// Initialize proc lazily to avoid creating unnecessary
			// garbage
			if proc == nil {
				proc = &Proc{
					PID:  uint(p.PID),
					Name: p.Name,
				}
			}

			sockets[statT.Ino] = proc
		}

	}

	return nil
}

// walk walks over all numerical (PID) /proc entries. It reads
// /proc/PID/net/tcp{,6} for each namespace and sees if the ./fd/* files of each
// process in that namespace are symlinks to sockets. Returns a map from socket
// ID (inode) to PID.
func (w pidWalker) walk(buf *bytes.Buffer) (map[uint64]*Proc, error) {
	var (
		sockets    = map[uint64]*Proc{}              // map socket inode -> process
		namespaces = map[uint64][]*process.Process{} // map network namespace id -> processes
		statT      syscall.Stat_t
	)

	// We do two process traversals: One to group processes by namespace and
	// another one to obtain their connections.
	//
	// The first traversal is needed to allow obtaining the connections on a
	// per-namespace basis. This is done to minimize the race condition
	// between reading /net/tcp{,6} of each namespace and /proc/PID/fd/* for
	// the processes living in that namespace.

	w.walker.Walk(func(p, _ process.Process) {
		dirName := strconv.Itoa(p.PID)

		netNamespacePath := filepath.Join(procRoot, dirName, getNetNamespacePathSuffix())
		if err := fs.Stat(netNamespacePath, &statT); err != nil {
			return
		}

		namespaceID := statT.Ino
		namespaces[namespaceID] = append(namespaces[namespaceID], &p)
	})

	for _, procs := range namespaces {
		select {
		case <-w.tickc:
			w.walkNamespace(buf, sockets, procs)
		case <-w.stopc:
			break // abort
		}
	}

	metrics.SetGauge(namespaceKey, float32(len(namespaces)))
	return sockets, nil
}

func (w pidWalker) stop() {
	close(w.stopc)
}

// readFile reads an arbitrary file into a buffer.
func readFile(filename string, buf *bytes.Buffer) (int64, error) {
	f, err := fs.Open(filename)
	if err != nil {
		return -1, err
	}
	defer f.Close()
	return buf.ReadFrom(f)
}
