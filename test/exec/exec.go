package exec

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/weaveworks/scope/common/exec"
)

type mockCmd struct {
	stdout io.ReadCloser
	stdin  io.WriteCloser
	quit   chan struct{}
}

type blockingReader struct {
	quit chan struct{}
}

// NewMockCmdString creates a new mock Cmd which has s on its stdout pipe
func NewMockCmdString(s string) exec.Cmd {
	return &mockCmd{
		stdout: struct {
			io.Reader
			io.Closer
		}{
			bytes.NewBufferString(s),
			ioutil.NopCloser(nil),
		},
		quit: make(chan struct{}),
	}
}

// NewMockCmd creates a new mock Cmd with rc as its stdout pipe
func NewMockCmd(stdout io.ReadCloser, stdin io.WriteCloser) exec.Cmd {
	return &mockCmd{
		stdout: stdout,
		stdin:  stdin,
		quit:   make(chan struct{}),
	}
}

func (c *mockCmd) Start() error {
	return nil
}

func (c *mockCmd) Wait() error {
	return nil
}

func (c *mockCmd) StdoutPipe() (io.ReadCloser, error) {
	return c.stdout, nil
}

func (c *mockCmd) StderrPipe() (io.ReadCloser, error) {
	return &blockingReader{c.quit}, nil
}

func (c *mockCmd) StdinPipe() (io.WriteCloser, error) {
	return c.stdin, nil
}

func (c *mockCmd) Kill() error {
	close(c.quit)
	return nil
}

func (b *blockingReader) Read(p []byte) (n int, err error) {
	<-b.quit
	return 0, nil
}

func (b *blockingReader) Close() error {
	<-b.quit
	return nil
}
