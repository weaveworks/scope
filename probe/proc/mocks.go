package proc

import (
	"bytes"
	"os"
	"time"
)

// a mocked process
type MockedProcess struct {
	Id, Comm, Cmdline string
}

func (p MockedProcess) Name() string       { return p.Id }
func (p MockedProcess) Size() int64        { return 0 }
func (p MockedProcess) Mode() os.FileMode  { return 0 }
func (p MockedProcess) ModTime() time.Time { return time.Now() }
func (p MockedProcess) IsDir() bool        { return true }
func (p MockedProcess) Sys() interface{}   { return nil }

// a mocked "/proc" directory
type MockedProcDir struct {
	Dir              string
	ReadDirFunc      func(string) ([]os.FileInfo, error)
	ReadFileFunc     func(string) ([]byte, error)
	ReadFileIntoFunc func(string, *bytes.Buffer) error
}

func (p MockedProcDir) Root() string                                 { return p.Dir }
func (p MockedProcDir) ReadDir(s string) ([]os.FileInfo, error)      { return p.ReadDirFunc(s) }
func (p MockedProcDir) ReadFile(s string) ([]byte, error)            { return p.ReadFileFunc(s) }
func (p MockedProcDir) ReadFileInto(s string, b *bytes.Buffer) error { return p.ReadFileIntoFunc(s, b) }

var EmptyProcDir = MockedProcDir{
	Dir:              "",
	ReadDirFunc:      func(string) ([]os.FileInfo, error) { return []os.FileInfo{}, nil },
	ReadFileFunc:     func(string) ([]byte, error) { return []byte{}, nil },
	ReadFileIntoFunc: func(string, *bytes.Buffer) error { return nil },
}

// a mocked /proc reader
type MockedProcReader struct {
	Procs []Process
	Conns []Connection
}

func (mw MockedProcReader) Processes(f func(Process)) error {
	for _, p := range mw.Procs {
		f(p)
	}
	return nil
}

func (mw *MockedProcReader) Connections(_ bool, f func(Connection)) error {
	for _, c := range mw.Conns {
		f(c)
	}
	return nil
}
