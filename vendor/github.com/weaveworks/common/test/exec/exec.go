package exec

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/weaveworks/common/exec"
)

type mockCmd struct {
	io.ReadCloser
	quit chan struct{}
}

type blockingReader struct {
	quit chan struct{}
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
		quit: make(chan struct{}),
	}
}

// NewMockCmd creates a new mock Cmd with rc as its stdout pipe
func NewMockCmd(rc io.ReadCloser) exec.Cmd {
	return &mockCmd{
		ReadCloser: rc,
		quit:       make(chan struct{}),
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
	return &blockingReader{c.quit}, nil
}

func (c *mockCmd) Kill() error {
	close(c.quit)
	return nil
}

func (c *mockCmd) Output() ([]byte, error) {
	return ioutil.ReadAll(c.ReadCloser)
}

func (c *mockCmd) Run() error {
	return nil
}

func (c *mockCmd) SetEnv([]string) {}

func (b *blockingReader) Read(p []byte) (n int, err error) {
	<-b.quit
	return 0, nil
}

func (b *blockingReader) Close() error {
	<-b.quit
	return nil
}
