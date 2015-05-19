package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

type dockerMapper struct {
	sync.RWMutex
	d        map[int]*docker.Container
	procRoot string
}

func newDockerMapper(procRoot string, interval time.Duration) *dockerMapper {
	m := dockerMapper{
		procRoot: procRoot,
		d:        map[int]*docker.Container{},
	}
	m.update()
	go m.loop(interval)
	return &m
}

func (m *dockerMapper) loop(d time.Duration) {
	for range time.Tick(d) {
		m.update()
	}
}

// for mocking
type dockerClient interface {
	ListContainers(docker.ListContainersOptions) ([]docker.APIContainers, error)
	InspectContainer(string) (*docker.Container, error)
}

func newRealDockerClient(endpoint string) (dockerClient, error) {
	return docker.NewClient(endpoint)
}

var (
	newDockerClient = newRealDockerClient
	newPIDTreeStub  = newPIDTree
)

func (m *dockerMapper) update() {
	pidTree, err := newPIDTreeStub(m.procRoot)
	if err != nil {
		log.Printf("docker mapper: %s", err)
		return
	}

	endpoint := "unix:///var/run/docker.sock"
	client, err := newDockerClient(endpoint)
	if err != nil {
		log.Printf("docker mapper: %s", err)
		return
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Printf("docker mapper: %s", err)
		return
	}

	pmap := map[int]*docker.Container{}
	for _, container := range containers {
		info, err := client.InspectContainer(container.ID)
		if err != nil {
			log.Printf("docker mapper: %s", err)
			continue
		}

		if !info.State.Running {
			continue
		}

		pids, err := pidTree.allChildren(info.State.Pid)
		if err != nil {
			log.Printf("docker mapper: %s", err)
			continue
		}
		for _, pid := range pids {
			pmap[pid] = info
		}
	}

	m.Lock()
	m.d = pmap
	m.Unlock()
}

type dockerIDMapper struct {
	*dockerMapper
}

func (m dockerIDMapper) Key() string { return "docker_id" }
func (m dockerIDMapper) Map(pid uint) (string, error) {
	m.RLock()
	container, ok := m.d[int(pid)]
	m.RUnlock()

	if !ok {
		return "", fmt.Errorf("no container found for PID %d", pid)
	}

	return container.ID, nil
}

type dockerNameMapper struct {
	*dockerMapper
}

func (m dockerNameMapper) Key() string { return "docker_name" }
func (m dockerNameMapper) Map(pid uint) (string, error) {
	m.RLock()
	container, ok := m.d[int(pid)]
	m.RUnlock()

	if !ok {
		return "", fmt.Errorf("no container found for PID %d", pid)
	}

	return container.Name, nil
}
