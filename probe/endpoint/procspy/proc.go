package procspy

// /proc-based implementation.

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/weaveworks/scope/common/fs"
)

var (
	procRoot = "/proc"
)

// SetProcRoot sets the location of the proc filesystem.
func SetProcRoot(root string) {
	procRoot = root
}

// made variables for mocking
var (
	readDir = ioutil.ReadDir
	lstat   = syscall.Lstat
	stat    = syscall.Stat
	open    = fs.Open
)

// walkProcPid walks over all numerical (PID) /proc entries, and sees if their
// ./fd/* files are symlink to sockets. Returns a map from socket ID (inode)
// to PID. Will return an error if /proc isn't there.
func walkProcPid(buf *bytes.Buffer) (map[uint64]Proc, error) {
	dirNames, err := readDir(procRoot)
	if err != nil {
		return nil, err
	}

	var (
		res        = map[uint64]Proc{}
		namespaces = map[uint64]struct{}{}
		statT      syscall.Stat_t
	)
	for _, entry := range dirNames {
		dirName := entry.Name()
		pid, err := strconv.ParseUint(dirName, 10, 0)
		if err != nil {
			// Not a number, so not a PID subdir.
			continue
		}

		fdBase := filepath.Join(procRoot, dirName, "fd")
		fds, err := readDir(fdBase)
		if err != nil {
			// Process is be gone by now, or we don't have access.
			continue
		}

		// Read network namespace, and if we haven't seen it before,
		// read /proc/<pid>/net/tcp
		err = lstat(filepath.Join(procRoot, dirName, "/ns/net"), &statT)
		if err != nil {
			continue
		}

		if _, ok := namespaces[statT.Ino]; !ok {
			namespaces[statT.Ino] = struct{}{}
			readFile(filepath.Join(procRoot, dirName, "/net/tcp"), buf)
			readFile(filepath.Join(procRoot, dirName, "/net/tcp6"), buf)
		}

		var name string
		for _, fd := range fds {
			// Direct use of syscall.Stat() to save garbage.
			err = stat(filepath.Join(fdBase, fd.Name()), &statT)
			if err != nil {
				continue
			}

			// We want sockets only.
			if statT.Mode&syscall.S_IFMT != syscall.S_IFSOCK {
				continue
			}

			if name == "" {
				if name = procName(filepath.Join(procRoot, dirName)); name == "" {
					// Process might be gone by now
					break
				}
			}

			res[statT.Ino] = Proc{
				PID:  uint(pid),
				Name: name,
			}
		}
	}

	return res, nil
}

// procName does a pid->name lookup.
func procName(base string) string {
	fh, err := open(filepath.Join(base, "/comm"))
	if err != nil {
		return ""
	}

	name := make([]byte, 64)
	l, err := fh.Read(name)
	fh.Close()
	if err != nil {
		return ""
	}

	if l < 2 {
		return ""
	}

	// drop trailing "\n"
	return string(name[:l-1])
}

// readFile reads an arbitrary file into a buffer. It's a variable so it can
// be overwritten for benchmarks. That's bad practice and we should change it
// to be a dependency.
var readFile = func(filename string, buf *bytes.Buffer) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	_, err = buf.ReadFrom(f)
	f.Close()
	return err
}
