package docker

import (
	"log"

	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

// Control IDs used by the docker intergation.
const (
	StopContainer    = "docker_stop_container"
	StartContainer   = "docker_start_container"
	RestartContainer = "docker_restart_container"
	PauseContainer   = "docker_pause_container"
	UnpauseContainer = "docker_unpause_container"

	waitTime = 10
)

func (r *registry) stopContainer(req xfer.Request) xfer.Response {
	log.Printf("Stopping container %s", req.NodeID)

	_, containerID, ok := report.ParseContainerNodeID(req.NodeID)
	if !ok {
		return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
	}

	return xfer.ResponseError(r.client.StopContainer(containerID, waitTime))
}

func (r *registry) startContainer(req xfer.Request) xfer.Response {
	log.Printf("Starting container %s", req.NodeID)

	_, containerID, ok := report.ParseContainerNodeID(req.NodeID)
	if !ok {
		return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
	}

	return xfer.ResponseError(r.client.StartContainer(containerID, nil))
}

func (r *registry) restartContainer(req xfer.Request) xfer.Response {
	log.Printf("Restarting container %s", req.NodeID)

	_, containerID, ok := report.ParseContainerNodeID(req.NodeID)
	if !ok {
		return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
	}

	return xfer.ResponseError(r.client.RestartContainer(containerID, waitTime))
}

func (r *registry) pauseContainer(req xfer.Request) xfer.Response {
	log.Printf("Pausing container %s", req.NodeID)

	_, containerID, ok := report.ParseContainerNodeID(req.NodeID)
	if !ok {
		return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
	}

	return xfer.ResponseError(r.client.PauseContainer(containerID))
}

func (r *registry) unpauseContainer(req xfer.Request) xfer.Response {
	log.Printf("Unpausing container %s", req.NodeID)

	_, containerID, ok := report.ParseContainerNodeID(req.NodeID)
	if !ok {
		return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
	}

	return xfer.ResponseError(r.client.UnpauseContainer(containerID))
}

func (r *registry) registerControls() {
	controls.Register(StopContainer, r.stopContainer)
	controls.Register(StartContainer, r.startContainer)
	controls.Register(RestartContainer, r.restartContainer)
	controls.Register(PauseContainer, r.pauseContainer)
	controls.Register(UnpauseContainer, r.unpauseContainer)
}
