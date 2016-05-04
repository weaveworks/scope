package docker

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	docker_client "github.com/fsouza/go-dockerclient"

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
	AnalyzeTraffic   = "docker_analyze_traffic"

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

func (r *registry) analyzeTraffic(containerID string, req xfer.Request) xfer.Response {
	dockerContainer, err := r.client.InspectContainer(containerID)
	if err != nil {
		return xfer.ResponseError(err)
	}
	if !dockerContainer.State.Running {
		return xfer.ResponseError(fmt.Errorf("Container not running"))
	}
	pid := fmt.Sprintf("%d", dockerContainer.State.Pid)
	cmd := []string{"nsenter", "-t", pid, "-n"}
	if _, ok := req.ControlArgs["raw_pipe"]; ok {
		// Analyze externally with wireshark
		// TODO: It would be even better to expose rpcapd instead since we could apply filters remotely and reduce traffic.
		//       Also, I think we would be able to associate the rpcap:// uri with wireshark, avoiding instructions templates.
		//       How to proxy rpcap through websockets, though?
		cmd = append(cmd, "dumpcap", "-i", "any", "-w", "-")
		// TODO consider using a uri scheme, like weave://scope/wireshark/%pipeurl
		//      instead of asking the user to type a command. It would require
		//      registering a URL scheme handler, for instance see:
		//      http://superuser.com/questions/548119/how-do-i-configure-custom-url-handlers-on-os-x
		template := "wireshark -i <(scope grabpipe %pipe_url)"
		handler := controls.MakeRawCommandHandler("raw traffic analyzer", r.pipes, cmd, template)
		handler(req)
	}
	// TODO: better defaults for tshark?
	cmd = append(cmd, "tshark", "-i", "any")
	handler := controls.MakeTTYCommandHandler("traffic analyzer", r.pipes, cmd)
	return handler(req)
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
	controls.Register(AnalyzeTraffic, captureContainerID(r.analyzeTraffic))
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
	controls.Rm(AnalyzeTraffic)
}
