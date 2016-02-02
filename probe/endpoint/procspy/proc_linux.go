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

// walkProcPid walks over all numerical (PID) /proc entries, and sees if their
// ./fd/* files are symlink to sockets. Returns a map from socket ID (inode)
// to PID. Will return an error if /proc isn't there.
func walkProcPid(buf *bytes.Buffer, walker process.Walker) (map[uint64]*Proc, error) {
	var (
		res        = map[uint64]*Proc{} // map socket inode -> process
		namespaces = map[uint64]bool{}  // map namespace id -> has connections
		statT      syscall.Stat_t
	)

	walker.Walk(func(p, _ process.Process) {
		dirName := strconv.Itoa(p.PID)
		fdBase := filepath.Join(procRoot, dirName, "fd")

		// Read network namespace, and if we haven't seen it before,
		// read /proc/<pid>/net/tcp
		netNamespacePath := filepath.Join(procRoot, dirName, getNetNamespacePathSuffix())
		if err := fs.Stat(netNamespacePath, &statT); err != nil {
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

	metrics.SetGauge(namespaceKey, float32(len(namespaces)))
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
