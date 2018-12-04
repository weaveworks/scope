package exec

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/weaveworks/common/exec"
)

type mockCmd struct {
	io.ReadCloser
}

// NewMockCmdString creates a new mock Cmd which has s on its stdout pipe
func NewMockCmdString(s string) exec.Cmd {
	return &mockCmd{
		ReadCloser: struct {
			io.Reader
			io.Closer
		}{
			bytes.NewBufferString(s),
			ioutil.NopCloser(nil),
		},
	}
}

// NewMockCmd creates a new mock Cmd with rc as its stdout pipe
func NewMockCmd(rc io.ReadCloser) exec.Cmd {
	return &mockCmd{
		ReadCloser: rc,
	}
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

func (c *mockCmd) StderrPipe() (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader(nil)), nil
}

func (c *mockCmd) Kill() error {
	return nil
}

func (c *mockCmd) Output() ([]byte, error) {
	return ioutil.ReadAll(c.ReadCloser)
}

func (c *mockCmd) Run() error {
	return nil
}

func (c *mockCmd) SetEnv([]string) {}
