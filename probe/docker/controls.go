package docker

import (
	docker_client "github.com/fsouza/go-dockerclient"

	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/report"
)

// Control IDs used by the docker integration.
const (
	StopContainer    = "docker_stop_container"
	StartContainer   = "docker_start_container"
	RestartContainer = "docker_restart_container"
	PauseContainer   = "docker_pause_container"
	UnpauseContainer = "docker_unpause_container"
	RemoveContainer  = "docker_remove_container"
	AttachContainer  = "docker_attach_container"
	ExecContainer    = "docker_exec_container"
	ResizeExecTTY    = "docker_resize_exec_tty"

	waitTime = 10
)

func (r *registry) stopContainer(containerID string, _ xfer.Request) xfer.Response {
	log.Infof("Stopping container %s", containerID)
	return xfer.ResponseError(r.client.StopContainer(containerID, waitTime))
}

func (r *registry) startContainer(containerID string, _ xfer.Request) xfer.Response {
	log.Infof("Starting container %s", containerID)
	return xfer.ResponseError(r.client.StartContainer(containerID, nil))
}

func (r *registry) restartContainer(containerID string, _ xfer.Request) xfer.Response {
	log.Infof("Restarting container %s", containerID)
	return xfer.ResponseError(r.client.RestartContainer(containerID, waitTime))
}

func (r *registry) pauseContainer(containerID string, _ xfer.Request) xfer.Response {
	log.Infof("Pausing container %s", containerID)
	return xfer.ResponseError(r.client.PauseContainer(containerID))
}

func (r *registry) unpauseContainer(containerID string, _ xfer.Request) xfer.Response {
	log.Infof("Unpausing container %s", containerID)
	return xfer.ResponseError(r.client.UnpauseContainer(containerID))
}

func (r *registry) removeContainer(containerID string, req xfer.Request) xfer.Response {
	log.Infof("Removing container %s", containerID)
	if err := r.client.RemoveContainer(docker_client.RemoveContainerOptions{
		ID: containerID,
	}); err != nil {
		return xfer.ResponseError(err)
	}
	return xfer.Response{
		RemovedNode: req.NodeID,
	}
}

func (r *registry) attachContainer(containerID string, req xfer.Request) xfer.Response {
	c, ok := r.GetContainer(containerID)
	if !ok {
		return xfer.ResponseErrorf("Not found: %s", containerID)
	}

	hasTTY := c.HasTTY()
	id, pipe, err := controls.NewPipe(r.pipes, req.AppID)
	if err != nil {
		return xfer.ResponseError(err)
	}
	local, _ := pipe.Ends()
	cw, err := r.client.AttachToContainerNonBlocking(docker_client.AttachToContainerOptions{
		Container:    containerID,
		RawTerminal:  hasTTY,
		Stream:       true,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		InputStream:  local,
		OutputStream: local,
		ErrorStream:  local,
	})
	if err != nil {
		pipe.Close()
		return xfer.ResponseError(err)
	}
	pipe.OnClose(func() {
		if err := cw.Close(); err != nil {
			log.Errorf("Error closing attachment to container %s: %v", containerID, err)
			return
		}
	})
	go func() {
		if err := cw.Wait(); err != nil {
			log.Errorf("Error waiting on attachment to container %s: %v", containerID, err)
		}
		pipe.Close()
	}()
	return xfer.Response{
		Pipe:   id,
		RawTTY: hasTTY,
	}
}

func (r *registry) execContainer(containerID string, req xfer.Request) xfer.Response {
	exec, err := r.client.CreateExec(docker_client.CreateExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/sh", "-c", "TERM=xterm exec $( (type getent > /dev/null 2>&1  && getent passwd root | cut -d: -f7 2>/dev/null) || echo /bin/sh)"},
		Container:    containerID,
	})
	if err != nil {
		return xfer.ResponseError(err)
	}

	id, pipe, err := controls.NewPipe(r.pipes, req.AppID)
	if err != nil {
		return xfer.ResponseError(err)
	}

	local, _ := pipe.Ends()
	cw, err := r.client.StartExecNonBlocking(exec.ID, docker_client.StartExecOptions{
		Tty:          true,
		RawTerminal:  true,
		InputStream:  local,
		OutputStream: local,
		ErrorStream:  local,
	})
	if err != nil {
		pipe.Close()
		return xfer.ResponseError(err)
	}

	r.Lock()
	r.pipeIDToexecID[id] = exec.ID
	r.Unlock()

	pipe.OnClose(func() {
		if err := cw.Close(); err != nil {
			log.Errorf("Error closing exec in container %s: %v", containerID, err)
			return
		}
		r.Lock()
		delete(r.pipeIDToexecID, id)
		r.Unlock()
	})
	go func() {
		if err := cw.Wait(); err != nil {
			log.Errorf("Error waiting on exec in container %s: %v", containerID, err)
		}
		pipe.Close()
	}()
	return xfer.Response{
		Pipe:             id,
		RawTTY:           true,
		ResizeTTYControl: ResizeExecTTY,
	}
}

func (r *registry) resizeExecTTY(pipeID string, height, width uint) xfer.Response {
	r.Lock()
	execID, ok := r.pipeIDToexecID[pipeID]
	r.Unlock()

	if !ok {
		return xfer.ResponseErrorf("Unknown pipeID (%q)", pipeID)
	}

	if err := r.client.ResizeExecTTY(execID, int(height), int(width)); err != nil {
		return xfer.ResponseErrorf(
			"Error setting terminal size (%d, %d) of pipe %s: %v",
			height, width, pipeID, err)
	}

	return xfer.Response{}
}

func captureContainerID(f func(string, xfer.Request) xfer.Response) func(xfer.Request) xfer.Response {
	return func(req xfer.Request) xfer.Response {
		containerID, ok := report.ParseContainerNodeID(req.NodeID)
		if !ok {
			return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
		}
		return f(containerID, req)
	}
}

func (r *registry) registerControls() {
	controls := map[string]xfer.ControlHandlerFunc{
		StopContainer:    captureContainerID(r.stopContainer),
		StartContainer:   captureContainerID(r.startContainer),
		RestartContainer: captureContainerID(r.restartContainer),
		PauseContainer:   captureContainerID(r.pauseContainer),
		UnpauseContainer: captureContainerID(r.unpauseContainer),
		RemoveContainer:  captureContainerID(r.removeContainer),
		AttachContainer:  captureContainerID(r.attachContainer),
		ExecContainer:    captureContainerID(r.execContainer),
		ResizeExecTTY:    xfer.ResizeTTYControlWrapper(r.resizeExecTTY),
	}
	r.handlerRegistry.Batch(nil, controls)
}

func (r *registry) deregisterControls() {
	controls := []string{
		StopContainer,
		StartContainer,
		RestartContainer,
		PauseContainer,
		UnpauseContainer,
		RemoveContainer,
		AttachContainer,
		ExecContainer,
		ResizeExecTTY,
	}
	r.handlerRegistry.Batch(controls, nil)
}
