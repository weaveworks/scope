package exec

import (
	"io"
	"os/exec"
)

// Cmd is a hook for mocking
type Cmd interface {
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
	Kill() error
}

// Command is a hook for mocking
var Command = realCommand

func realCommand(name string, args ...string) Cmd {
	return &realCmd{exec.Command(name, args...)}
}

type realCmd struct {
	*exec.Cmd
}

func (c *realCmd) Kill() error {
	return c.Cmd.Process.Kill()
}

// Mock out Command with the supplied function
func Mock(f func(string, ...string) Cmd) {
	Command = f
}

// Restore the original Command
func Restore() {
	Command = realCommand
}
