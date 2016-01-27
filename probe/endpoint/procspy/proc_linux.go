package procspy

// /proc-based implementation.

import (
	"bytes"
	"log"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/armon/go-metrics"
	"github.com/hashicorp/go-version"

	"github.com/weaveworks/scope/common/fs"
	"github.com/weaveworks/scope/common/marshal"
	"github.com/weaveworks/scope/probe/process"
)

var (
	procRoot               = "/proc"
	namespaceKey           = []string{"procspy", "namespaces"}
	netNamespacePathSuffix = ""
)

// SetProcRoot sets the location of the proc filesystem.
func SetProcRoot(root string) {
	procRoot = root
}

func getKernelVersion() (*version.Version, error) {
	var u syscall.Utsname
	if err := syscall.Uname(&u); err != nil {
		return nil, err
	}

	release := marshal.FromUtsname(u.Release)
	return version.NewVersion(release)
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

	v, err := getKernelVersion()
	if err != nil {
		log.Printf("error: getNeNameSpacePath: cannot get kernel version: %s\n", err)
		netNamespacePathSuffix = post38Path
		return netNamespacePathSuffix
	}

	v38, _ := version.NewVersion("3.8")
	if v.LessThan(v38) {
		netNamespacePathSuffix = pre38Path
	} else {
		netNamespacePathSuffix = post38Path
	}
	return netNamespacePathSuffix
}

// walkNamespacePid does the work of walkProcPid for a single namespace
func walkNamespacePid(buf *bytes.Buffer, sockets map[uint64]*Proc, namespaceProcs []*process.Process) {

	// Read the connections for the namespace, which are found (identically) in
	// /proc/PID/net/tcp{,6} for any of the processes in the namespace.
	var tcpSuccess bool
	for _, p := range namespaceProcs {
		dirName := strconv.Itoa(p.PID)

		read, errRead := readFile(filepath.Join(procRoot, dirName, "/net/tcp"), buf)
		read6, errRead6 := readFile(filepath.Join(procRoot, dirName, "/net/tcp6"), buf)

		if errRead != nil || errRead6 != nil {
			// try next process
			continue
		}

		if read+read6 == 0 {
			// No connections, don't bother reading /fd/*
			return
		}

		tcpSuccess = true
		break
	}

	if !tcpSuccess {
		// There's no point in reading /fd/*
		return
	}

	// Get the sockets for all the processes in the namespace
	var statT syscall.Stat_t
	for _, p := range namespaceProcs {
		dirName := strconv.Itoa(p.PID)
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
}

// walkProcPid walks over all numerical (PID) /proc entries. It reads
// /proc/PID/net/tcp{,6} for each namespace and sees if the ./fd/* files of each
// process in that namespace are symlinks to sockets. Returns a map from socket
// ID (inode) to PID.
func walkProcPid(buf *bytes.Buffer, walker process.Walker) (map[uint64]*Proc, error) {
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

	walker.Walk(func(p, _ process.Process) {
		dirName := strconv.Itoa(p.PID)

		netNamespacePath := filepath.Join(procRoot, dirName, getNetNamespacePathSuffix())
		if err := fs.Stat(netNamespacePath, &statT); err != nil {
			return
		}

		namespaceID := statT.Ino
		namespaces[namespaceID] = append(namespaces[namespaceID], &p)
	})

	for _, procs := range namespaces {
		walkNamespacePid(buf, sockets, procs)
	}

	metrics.SetGauge(namespaceKey, float32(len(namespaces)))
	return sockets, nil
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
