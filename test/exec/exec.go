package exec

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/weaveworks/scope/common/exec"
)

type mockCmd struct {
	io.ReadCloser
}

// NewMockCmdString creates a new mock Cmd which has s on its stdout pipe
func NewMockCmdString(s string) exec.Cmd {
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
func NewMockCmd(rc io.ReadCloser) exec.Cmd {
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
