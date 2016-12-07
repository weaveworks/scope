package fs

import (
	"io"
	"io/ioutil"
	"os"
	"syscall"
)

// Interface is the filesystem interface type.
type Interface interface {
	ReadDir(string) ([]os.FileInfo, error)
	ReadDirNames(string) ([]string, error)
	ReadFile(string) ([]byte, error)
	Lstat(string, *syscall.Stat_t) error
	Stat(string, *syscall.Stat_t) error
	Open(string) (io.ReadWriteCloser, error)
}

type realFS struct{}

// FS is the way you should access the filesystem.
var fs Interface = realFS{}

func (realFS) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}

func (realFS) ReadDirNames(path string) ([]string, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	return fh.Readdirnames(-1)
}

func (realFS) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (realFS) Lstat(path string, stat *syscall.Stat_t) error {
	return syscall.Lstat(path, stat)
}

func (realFS) Stat(path string, stat *syscall.Stat_t) error {
	return syscall.Stat(path, stat)
}

func (realFS) Open(path string) (io.ReadWriteCloser, error) {
	return os.Open(path)
}

// trampolines here to allow users to do fs.ReadDir etc

// ReadDir see ioutil.ReadDir
func ReadDir(path string) ([]os.FileInfo, error) {
	return fs.ReadDir(path)
}

// ReadDirNames see os.File.ReadDirNames
func ReadDirNames(path string) ([]string, error) {
	return fs.ReadDirNames(path)
}

// ReadFile see ioutil.ReadFile
func ReadFile(path string) ([]byte, error) {
	return fs.ReadFile(path)
}

// Lstat see syscall.Lstat
func Lstat(path string, stat *syscall.Stat_t) error {
	return fs.Lstat(path, stat)
}

// Stat see syscall.Stat
func Stat(path string, stat *syscall.Stat_t) error {
	return fs.Stat(path, stat)
}

// Open see os.Open
func Open(path string) (io.ReadWriteCloser, error) {
	return fs.Open(path)
}

// Mock is used to switch out the filesystem for a mock.
func Mock(mock Interface) {
	fs = mock
}

// Restore puts back the real filesystem.
func Restore() {
	fs = realFS{}
}
