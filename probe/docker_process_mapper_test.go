package main

import (
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

type mockDockerClient struct {
	containers    []docker.APIContainers
	containerInfo map[string]*docker.Container
}

func (m mockDockerClient) ListContainers(options docker.ListContainersOptions) ([]docker.APIContainers, error) {
	return m.containers, nil
}

func (m mockDockerClient) InspectContainer(id string) (*docker.Container, error) {
	return m.containerInfo[id], nil
}

func TestDockerProcessMapper(t *testing.T) {
	oldPIDTreeStub, oldDockerClientStub := newPIDTreeStub, newDockerClient
	defer func() {
		newPIDTreeStub = oldPIDTreeStub
		newDockerClient = oldDockerClientStub
	}()

	newPIDTreeStub = func(procRoot string) (*pidTree, error) {
		pid1 := &process{pid: 1}
		pid2 := &process{pid: 2, ppid: 1, parent: pid1}
		pid1.children = []*process{pid2}

		return &pidTree{
			processes: map[int]*process{
				1: pid1, 2: pid2,
			},
		}, nil
	}

	newDockerClient = func(endpoint string) (dockerClient, error) {
		return mockDockerClient{
			containers: []docker.APIContainers{{ID: "foo"}},
			containerInfo: map[string]*docker.Container{
				"foo": {
					ID:    "foo",
					Name:  "bar",
					State: docker.State{Pid: 1, Running: true},
				},
			},
		}, nil
	}

	dockerMapper := newDockerMapper("/proc", 10*time.Second)
	dockerIDMapper := dockerIDMapper{dockerMapper}
	dockerNameMapper := dockerNameMapper{dockerMapper}

	for pid, want := range map[uint]struct{ id, name string }{
		1: {"foo", "bar"},
		2: {"foo", "bar"},
	} {
		haveID, err := dockerIDMapper.Map(pid)
		if err != nil || want.id != haveID {
			t.Errorf("%d: want %q, have %q (%v)", pid, want.id, haveID, err)
		}
		haveName, err := dockerNameMapper.Map(pid)
		if err != nil || want.name != haveName {
			t.Errorf("%d: want %q, have %q (%v)", pid, want.name, haveName, err)
		}
	}
}
