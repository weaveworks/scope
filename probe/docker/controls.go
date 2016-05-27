package docker

import (
	docker_client "github.com/fsouza/go-dockerclient"

	log "github.com/Sirupsen/logrus"

	"$GITHUB_URI/common/xfer"
	"$GITHUB_URI/probe/controls"
	"$GITHUB_URI/report"
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
		return xfer.ResponseError(err)
	}
	pipe.OnClose(func() {
		if err := cw.Close(); err != nil {
			log.Errorf("Error closing attachment: %v", err)
			return
		}
		log.Infof("Attachment to container %s closed.", containerID)
	})
	go func() {
		if err := cw.Wait(); err != nil {
			log.Errorf("Error waiting on exec: %v", err)
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
		return xfer.ResponseError(err)
	}
	pipe.OnClose(func() {
		if err := cw.Close(); err != nil {
			log.Errorf("Error closing exec: %v", err)
			return
		}
		log.Infof("Exec on container %s closed.", containerID)
	})
	go func() {
		if err := cw.Wait(); err != nil {
			log.Errorf("Error waiting on exec: %v", err)
		}
		pipe.Close()
	}()
	return xfer.Response{
		Pipe:   id,
		RawTTY: true,
	}
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
	controls.Register(StopContainer, captureContainerID(r.stopContainer))
	controls.Register(StartContainer, captureContainerID(r.startContainer))
	controls.Register(RestartContainer, captureContainerID(r.restartContainer))
	controls.Register(PauseContainer, captureContainerID(r.pauseContainer))
	controls.Register(UnpauseContainer, captureContainerID(r.unpauseContainer))
	controls.Register(RemoveContainer, captureContainerID(r.removeContainer))
	controls.Register(AttachContainer, captureContainerID(r.attachContainer))
	controls.Register(ExecContainer, captureContainerID(r.execContainer))
}

func (r *registry) deregisterControls() {
	controls.Rm(StopContainer)
	controls.Rm(StartContainer)
	controls.Rm(RestartContainer)
	controls.Rm(PauseContainer)
	controls.Rm(UnpauseContainer)
	controls.Rm(RemoveContainer)
	controls.Rm(AttachContainer)
	controls.Rm(ExecContainer)
}
