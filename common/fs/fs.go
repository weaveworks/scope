package fs

import (
	"io"
	"io/ioutil"
	"os"
	"syscall"
)

// T is the filesystem interface type.
type T interface {
	ReadDir(string) ([]os.FileInfo, error)
	ReadFile(string) ([]byte, error)
	Lstat(string, *syscall.Stat_t) error
	Stat(string, *syscall.Stat_t) error
	Open(string) (io.ReadWriteCloser, error)
}

type realFS struct{}

// FS is the way you should access the filesystem.
var FS T = realFS{}

// Mock is used to switch out the filesystem for a mock.
func Mock(fs T) {
	FS = fs
}

// Restore puts back the real filesystem.
func Restore() {
	FS = realFS{}
}

func (realFS) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
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
