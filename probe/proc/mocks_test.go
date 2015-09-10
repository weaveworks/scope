package proc_test

import (
	"bytes"

	"github.com/weaveworks/scope/probe/proc"
)

// mockedFile is a mocked file in the "/proc" directory
type mockedFile struct {
	Path         string
	ReadIntoFunc func(buf *bytes.Buffer) error
}

func (mpf mockedFile) ReadInto(buf *bytes.Buffer) error { return mpf.ReadIntoFunc(buf) }
func (mpf mockedFile) Close() error                     { return nil }

// mockedDir is a mocked "/proc" directory
type mockedDir struct {
	Dir              string
	OpenFunc         func(string) (proc.File, error)
	ReadDirNamesFunc func(string) ([]string, error)
}

func (p mockedDir) Root() string                            { return p.Dir }
func (p mockedDir) Open(s string) (proc.File, error)        { return p.OpenFunc(s) }
func (p mockedDir) ReadDirNames(s string) ([]string, error) { return p.ReadDirNamesFunc(s) }
