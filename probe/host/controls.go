package host

import (
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/kr/pty"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
)

// Control IDs used by the host integration.
const (
	ExecHost = "host_exec"
)

func (r *Reporter) registerControls() {
	controls.Register(ExecHost, r.execHost)
}

func (*Reporter) deregisterControls() {
	controls.Rm(ExecHost)
}

func (r *Reporter) execHost(req xfer.Request) xfer.Response {
	cmd := exec.Command(hostShellCmd[0], hostShellCmd[1:]...)
	cmd.Env = []string{"TERM=xterm"}
	ptyPipe, err := pty.Start(cmd)
	if err != nil {
		return xfer.ResponseError(err)
	}

	id, pipe, err := controls.NewPipeFromEnds(nil, ptyPipe, r.pipes, req.AppID)
	if err != nil {
		return xfer.ResponseError(err)
	}
	pipe.OnClose(func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Errorf("Error closing host shell: %v", err)
			return
		}
		log.Info("Host shell closed.")
	})
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Errorf("Error waiting on host shell: %v", err)
		}
		ptyPipe.Close()
		pipe.Close()
	}()

	return xfer.Response{
		Pipe:   id,
		RawTTY: true,
	}
}
