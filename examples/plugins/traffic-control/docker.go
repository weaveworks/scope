package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	docker "github.com/fsouza/go-dockerclient"
)

type DockerClient struct {
	store  *Store
	client *docker.Client
}

func NewDockerClient(store *Store) (*DockerClient, error) {
	dc, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker daemon: %v", err)
	}
	return &DockerClient{
		store:  store,
		client: dc,
	}, nil
}

func (c *DockerClient) Start() {
	for {
		c.loopIteration()
		time.Sleep(time.Second)
	}
}

func (c *DockerClient) loopIteration() {
	events := make(chan *docker.APIEvents)
	if err := c.client.AddEventListener(events); err != nil {
		log.Error(err)
		return
	}
	defer func() {
		if err := c.client.RemoveEventListener(events); err != nil {
			log.Error(err)
		}
	}()
	if err := c.getContainers(); err != nil {
		log.Error(err)
		return
	}
	for {
		event, ok := <-events
		if !ok {
			log.Error("event listener unexpectedly disconnected")
			return
		}
		c.handleEvent(event)
	}
}

func (c *DockerClient) getContainers() error {
	apiContainers, err := c.client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return err
	}

	for _, apiContainer := range apiContainers {
		containerState, err := c.getContainerState(apiContainer.ID)
		if err != nil {
			log.Error(err)
			continue
		}
		state := Destroyed
		switch {
		case containerState.Dead || containerState.Paused || containerState.Restarting || containerState.OOMKilled:
			state = Stopped
		case containerState.Running:
			state = Running
		}
		c.updateContainer(apiContainer.ID, state, containerState.Pid)
	}

	return nil
}

func (c *DockerClient) handleEvent(event *docker.APIEvents) {
	var state State
	switch event.Status {
	case "create":
		state = Created
	case "destroy":
		state = Destroyed
	case "start", "unpause":
		state = Running
	case "die", "pause":
		state = Stopped
	default:
		return
	}
	pid, err := c.getContainerPID(event.ID)
	if err != nil {
		log.Error(err)
		return
	}
	c.updateContainer(event.ID, state, pid)
}

func (c *DockerClient) getContainerPID(containerID string) (int, error) {
	containerState, err := c.getContainerState(containerID)
	if containerState == nil {
		return 0, err
	}
	return containerState.Pid, nil
}

func (c *DockerClient) getContainerState(containerID string) (*docker.State, error) {
	dockerContainer, err := c.getContainer(containerID)
	if dockerContainer == nil {
		return nil, err
	}
	return &dockerContainer.State, nil
}

func (c *DockerClient) getContainer(containerID string) (*docker.Container, error) {
	dockerContainer, err := c.client.InspectContainer(containerID)
	if err != nil {
		if _, ok := err.(*docker.NoSuchContainer); !ok {
			return nil, err
		}
		return nil, nil
	}
	return dockerContainer, nil
}

func (c *DockerClient) updateContainer(containerID string, state State, pid int) {
	if state == Destroyed {
		c.store.DeleteContainer(containerID)
		return
	}
	cont := Container{
		State: state,
		PID:   pid,
	}
	c.store.SetContainer(containerID, cont)
}
