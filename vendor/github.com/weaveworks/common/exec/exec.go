package exec

import (
	"io"
	"os/exec"
)

// Cmd is a hook for mocking
type Cmd interface {
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
	Kill() error
	Output() ([]byte, error)
	Run() error
	SetEnv([]string)
}

// Command is a hook for mocking
var Command = func(name string, args ...string) Cmd {
	return &realCmd{exec.Command(name, args...)}
}

type realCmd struct {
	*exec.Cmd
}

func (c *realCmd) Kill() error {
	return c.Cmd.Process.Kill()
}

func (c *realCmd) SetEnv(env []string) {
	c.Cmd.Env = env
}
