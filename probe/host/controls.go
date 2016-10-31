package host

import (
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/term"
	"github.com/kr/pty"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
)

// Control IDs used by the host integration.
const (
	ExecHost      = "host_exec"
	ResizeExecTTY = "host_resize_exec_tty"
)

func (r *Reporter) registerControls() {
	r.handlerRegistry.Register(ExecHost, r.execHost)
	r.handlerRegistry.Register(ResizeExecTTY, xfer.ResizeTTYControlWrapper(r.resizeExecTTY))
}

func (r *Reporter) deregisterControls() {
	r.handlerRegistry.Rm(ExecHost)
	r.handlerRegistry.Rm(ResizeExecTTY)
}

func (r *Reporter) execHost(req xfer.Request) xfer.Response {
	cmd := exec.Command(r.hostShellCmd[0], r.hostShellCmd[1:]...)
	cmd.Env = []string{"TERM=xterm"}
	ptyPipe, err := pty.Start(cmd)
	if err != nil {
		return xfer.ResponseError(err)
	}

	id, pipe, err := controls.NewPipeFromEnds(nil, ptyPipe, r.pipes, req.AppID)
	if err != nil {
		return xfer.ResponseError(err)
	}

	r.Lock()
	r.pipeIDToTTY[id] = ptyPipe.Fd()
	r.Unlock()

	pipe.OnClose(func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Errorf("Error stopping host shell: %v", err)
		}
		if err := ptyPipe.Close(); err != nil {
			log.Errorf("Error closing host shell's pty: %v", err)
		}
		r.Lock()
		delete(r.pipeIDToTTY, id)
		r.Unlock()
		log.Info("Host shell closed.")
	})
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Errorf("Error waiting on host shell: %v", err)
		}
		pipe.Close()
	}()

	return xfer.Response{
		Pipe:             id,
		RawTTY:           true,
		ResizeTTYControl: ResizeExecTTY,
	}
}

func (r *Reporter) resizeExecTTY(pipeID string, height, width uint) xfer.Response {
	r.Lock()
	fd, ok := r.pipeIDToTTY[pipeID]
	r.Unlock()

	if !ok {
		return xfer.ResponseErrorf("Unknown pipeID (%q)", pipeID)
	}

	size := term.Winsize{
		Height: uint16(height),
		Width:  uint16(width),
	}

	if err := term.SetWinsize(fd, &size); err != nil {
		return xfer.ResponseErrorf(
			"Error setting terminal size (%d, %d) of pipe %s: %v",
			height, width, pipeID, err)
	}

	return xfer.Response{}

}
