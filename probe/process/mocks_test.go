package process_test

import (
	"bytes"
	"syscall"

	"github.com/weaveworks/scope/probe/process"
)

// mockedFileWithBytes is a mocked file in the "/proc" directory
type mockedFileWithBytes struct {
	content []byte
}

func (mpf mockedFileWithBytes) ReadInto(buf *bytes.Buffer) error {
	_, err := buf.Write(mpf.content)
	return err
}

func (mpf mockedFileWithBytes) Close() error {
	return nil
}

// mockedDir is a mocked "/proc" directory
type mockedDir struct {
	Dir              string
	OpenFunc         func(string) (process.File, error)
	ReadDirNamesFunc func(string) ([]string, error)
}

func (p mockedDir) Root() string                                           { return p.Dir }
func (p mockedDir) Open(s string) (process.File, error)                    { return p.OpenFunc(s) }
func (p mockedDir) ReadDirNames(s string) ([]string, error)                { return p.ReadDirNamesFunc(s) }
func (p mockedDir) Stat(s string, follow bool, stat *syscall.Stat_t) error { return nil }
