package exec

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

// Cmd is a hook for mocking
type Cmd interface {
	StdoutPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
	Process() *os.Process
}

// Command is a hook for mocking
var Command = func(name string, args ...string) Cmd {
	return &realCmd{exec.Command(name, args...)}
}

type realCmd struct {
	*exec.Cmd
}

func (c *realCmd) Process() *os.Process {
	return c.Cmd.Process
}

type mockCmd struct {
	io.ReadCloser
}

// NewMockCmdString creates a new mock Cmd which has s on its stdout pipe
func NewMockCmdString(s string) Cmd {
	return &mockCmd{
		struct {
			io.Reader
			io.Closer
		}{
			bytes.NewBufferString(s),
			ioutil.NopCloser(nil),
		},
	}
}

// NewMockCmd creates a new mock Cmd with rc as its stdout pipe
func NewMockCmd(rc io.ReadCloser) Cmd {
	return &mockCmd{rc}
}

func (c *mockCmd) Start() error {
	return nil
}

func (c *mockCmd) Wait() error {
	return nil
}

func (c *mockCmd) StdoutPipe() (io.ReadCloser, error) {
	return c.ReadCloser, nil
}

func (c *mockCmd) Process() *os.Process {
	return nil
}
