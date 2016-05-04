package controls

import (
	"io"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/kr/pty"

	"github.com/weaveworks/scope/common/xfer"
)

// MakeTTYCommandHandler creates a control handler which spawns a command in a tty
func MakeTTYCommandHandler(description string, client PipeClient, command []string) xfer.ControlHandlerFunc {
	return func(req xfer.Request) xfer.Response {
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Env = []string{"TERM=xterm"}
		ptyPipe, err := pty.Start(cmd)
		if err != nil {
			return xfer.ResponseError(err)
		}

		id, pipe, err := NewPipeFromEnds(nil, ptyPipe, client, req.AppID)
		if err != nil {
			return xfer.ResponseError(err)
		}
		deferCleanUp(description, cmd, ptyPipe, pipe)
		return xfer.Response{
			Pipe:   id,
			RawTTY: true,
		}
	}
}

type rawCommandPipe struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (rcp rawCommandPipe) Read(p []byte) (int, error) {
	return rcp.stdout.Read(p)
}

func (rcp rawCommandPipe) Write(p []byte) (int, error) {
	return rcp.stdin.Write(p)
}

func (rcp *rawCommandPipe) Close() error {
	err1 := rcp.stdout.Close()
	err2 := rcp.stdin.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

// MakeRawCommandHandler creates a control handler which spawns a command on a raw pipe
func MakeRawCommandHandler(description string, client PipeClient, command []string, template string) xfer.ControlHandlerFunc {
	return func(req xfer.Request) xfer.Response {
		cmd := exec.Command(command[0], command[1:]...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return xfer.ResponseError(err)
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return xfer.ResponseError(err)
		}
		// TODO: handle stderr. how?
		commandPipe := rawCommandPipe{stdin, stdout}
		if err := cmd.Start(); err != nil {
			return xfer.ResponseError(err)
		}
		id, pipe, err := NewPipeFromEnds(nil, commandPipe, client, req.AppID)
		if err != nil {
			return xfer.ResponseError(err)
		}
		deferCleanUp(description, cmd, stdout, pipe)
		return xfer.Response{
			Pipe:            id,
			RawPipeTemplate: template,
		}
	}
}

func deferCleanUp(description string, cmd *exec.Cmd, cmdOutput io.ReadCloser, pipe xfer.Pipe) {
	pipe.OnClose(func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Errorf("Error stopping %s: %v", description, err)
		}
		if err := cmdOutput.Close(); err != nil {
			log.Errorf("Error closing %s output: %v", description, err)
		}
		log.Infof("%s closed.", description)
	})
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Errorf("Error waiting on %s: %v", description, err)
		}
		pipe.Close()
	}()
}
