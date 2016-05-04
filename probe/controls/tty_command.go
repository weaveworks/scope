package controls

import (
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
		pipe.OnClose(func() {
			if err := cmd.Process.Kill(); err != nil {
				log.Errorf("Error stopping %s: %v", description, err)
			}
			if err := ptyPipe.Close(); err != nil {
				log.Errorf("Error closing %s pty: %v", description, err)
			}
			log.Infof("%s closed.", description)
		})
		go func() {
			if err := cmd.Wait(); err != nil {
				log.Errorf("Error waiting on %s: %v", description, err)
			}
			pipe.Close()
		}()

		return xfer.Response{
			Pipe:   id,
			RawTTY: true,
		}
	}
}
