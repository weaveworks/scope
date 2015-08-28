package exec

import (
	"io"
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
