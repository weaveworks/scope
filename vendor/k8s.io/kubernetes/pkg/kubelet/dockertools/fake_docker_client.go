/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dockertools

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/util/sets"
)

// FakeDockerClient is a simple fake docker client, so that kubelet can be run for testing without requiring a real docker setup.
type FakeDockerClient struct {
	sync.Mutex
	ContainerList       []docker.APIContainers
	ExitedContainerList []docker.APIContainers
	ContainerMap        map[string]*docker.Container
	Image               *docker.Image
	Images              []docker.APIImages
	Errors              map[string]error
	called              []string
	pulled              []string
	// Created, Stopped and Removed all container docker ID
	Created       []string
	Stopped       []string
	Removed       []string
	RemovedImages sets.String
	VersionInfo   docker.Env
	Information   docker.Env
	ExecInspect   *docker.ExecInspect
	execCmd       []string
	EnableSleep   bool
}

func NewFakeDockerClient() *FakeDockerClient {
	return NewFakeDockerClientWithVersion("1.8.1", "1.20")
}

func NewFakeDockerClientWithVersion(version, apiVersion string) *FakeDockerClient {
	return &FakeDockerClient{
		VersionInfo:   docker.Env{fmt.Sprintf("Version=%s", version), fmt.Sprintf("ApiVersion=%s", apiVersion)},
		Errors:        make(map[string]error),
		RemovedImages: sets.String{},
		ContainerMap:  make(map[string]*docker.Container),
	}
}

func (f *FakeDockerClient) InjectError(fn string, err error) {
	f.Lock()
	defer f.Unlock()
	f.Errors[fn] = err
}

func (f *FakeDockerClient) InjectErrors(errs map[string]error) {
	f.Lock()
	defer f.Unlock()
	for fn, err := range errs {
		f.Errors[fn] = err
	}
}

func (f *FakeDockerClient) ClearErrors() {
	f.Lock()
	defer f.Unlock()
	f.Errors = map[string]error{}
}

func (f *FakeDockerClient) ClearCalls() {
	f.Lock()
	defer f.Unlock()
	f.called = []string{}
	f.Stopped = []string{}
	f.pulled = []string{}
	f.Created = []string{}
	f.Removed = []string{}
}

func (f *FakeDockerClient) SetFakeContainers(containers []*docker.Container) {
	f.Lock()
	defer f.Unlock()
	// Reset the lists and the map.
	f.ContainerMap = map[string]*docker.Container{}
	f.ContainerList = []docker.APIContainers{}
	f.ExitedContainerList = []docker.APIContainers{}

	for i := range containers {
		c := containers[i]
		if c.Config == nil {
			c.Config = &docker.Config{}
		}
		f.ContainerMap[c.ID] = c
		apiContainer := docker.APIContainers{
			Names: []string{c.Name},
			ID:    c.ID,
		}
		if c.State.Running {
			f.ContainerList = append(f.ContainerList, apiContainer)
		} else {
			f.ExitedContainerList = append(f.ExitedContainerList, apiContainer)
		}
	}
}

func (f *FakeDockerClient) SetFakeRunningContainers(containers []*docker.Container) {
	for _, c := range containers {
		c.State.Running = true
	}
	f.SetFakeContainers(containers)
}

func (f *FakeDockerClient) AssertCalls(calls []string) (err error) {
	f.Lock()
	defer f.Unlock()

	if !reflect.DeepEqual(calls, f.called) {
		err = fmt.Errorf("expected %#v, got %#v", calls, f.called)
	}

	return
}

func (f *FakeDockerClient) AssertCreated(created []string) error {
	f.Lock()
	defer f.Unlock()

	actualCreated := []string{}
	for _, c := range f.Created {
		dockerName, _, err := ParseDockerName(c)
		if err != nil {
			return fmt.Errorf("unexpected error: %v", err)
		}
		actualCreated = append(actualCreated, dockerName.ContainerName)
	}
	sort.StringSlice(created).Sort()
	sort.StringSlice(actualCreated).Sort()
	if !reflect.DeepEqual(created, actualCreated) {
		return fmt.Errorf("expected %#v, got %#v", created, actualCreated)
	}
	return nil
}

func (f *FakeDockerClient) AssertStopped(stopped []string) error {
	f.Lock()
	defer f.Unlock()
	sort.StringSlice(stopped).Sort()
	sort.StringSlice(f.Stopped).Sort()
	if !reflect.DeepEqual(stopped, f.Stopped) {
		return fmt.Errorf("expected %#v, got %#v", stopped, f.Stopped)
	}
	return nil
}

func (f *FakeDockerClient) AssertUnorderedCalls(calls []string) (err error) {
	f.Lock()
	defer f.Unlock()

	expected := make([]string, len(calls))
	actual := make([]string, len(f.called))
	copy(expected, calls)
	copy(actual, f.called)

	sort.StringSlice(expected).Sort()
	sort.StringSlice(actual).Sort()

	if !reflect.DeepEqual(actual, expected) {
		err = fmt.Errorf("expected(sorted) %#v, got(sorted) %#v", expected, actual)
	}
	return
}

func (f *FakeDockerClient) popError(op string) error {
	if f.Errors == nil {
		return nil
	}
	err, ok := f.Errors[op]
	if ok {
		delete(f.Errors, op)
		return err
	} else {
		return nil
	}
}

// ListContainers is a test-spy implementation of DockerInterface.ListContainers.
// It adds an entry "list" to the internal method call record.
func (f *FakeDockerClient) ListContainers(options docker.ListContainersOptions) ([]docker.APIContainers, error) {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "list")
	err := f.popError("list")
	containerList := append([]docker.APIContainers{}, f.ContainerList...)
	if options.All {
		// Although the container is not sorted, but the container with the same name should be in order,
		// that is enough for us now.
		// TODO(random-liu): Is a fully sorted array needed?
		containerList = append(containerList, f.ExitedContainerList...)
	}
	return containerList, err
}

// InspectContainer is a test-spy implementation of DockerInterface.InspectContainer.
// It adds an entry "inspect" to the internal method call record.
func (f *FakeDockerClient) InspectContainer(id string) (*docker.Container, error) {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "inspect_container")
	err := f.popError("inspect_container")
	if container, ok := f.ContainerMap[id]; ok {
		return container, err
	}
	return nil, err
}

// InspectImage is a test-spy implementation of DockerInterface.InspectImage.
// It adds an entry "inspect" to the internal method call record.
func (f *FakeDockerClient) InspectImage(name string) (*docker.Image, error) {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "inspect_image")
	err := f.popError("inspect_image")
	return f.Image, err
}

// Sleeps random amount of time with the normal distribution with given mean and stddev
// (in milliseconds), we never sleep less than cutOffMillis
func (f *FakeDockerClient) normalSleep(mean, stdDev, cutOffMillis int) {
	if !f.EnableSleep {
		return
	}
	cutoff := (time.Duration)(cutOffMillis) * time.Millisecond
	delay := (time.Duration)(rand.NormFloat64()*float64(stdDev)+float64(mean)) * time.Millisecond
	if delay < cutoff {
		delay = cutoff
	}
	time.Sleep(delay)
}

// CreateContainer is a test-spy implementation of DockerInterface.CreateContainer.
// It adds an entry "create" to the internal method call record.
func (f *FakeDockerClient) CreateContainer(c docker.CreateContainerOptions) (*docker.Container, error) {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "create")
	if err := f.popError("create"); err != nil {
		return nil, err
	}
	// This is not a very good fake. We'll just add this container's name to the list.
	// Docker likes to add a '/', so copy that behavior.
	name := "/" + c.Name
	f.Created = append(f.Created, name)
	// The newest container should be in front, because we assume so in GetPodStatus()
	f.ContainerList = append([]docker.APIContainers{
		{ID: name, Names: []string{name}, Image: c.Config.Image, Labels: c.Config.Labels},
	}, f.ContainerList...)
	container := docker.Container{ID: name, Name: name, Config: c.Config, HostConfig: c.HostConfig}
	containerCopy := container
	f.ContainerMap[name] = &containerCopy
	f.normalSleep(100, 25, 25)
	return &container, nil
}

// StartContainer is a test-spy implementation of DockerInterface.StartContainer.
// It adds an entry "start" to the internal method call record.
// The HostConfig at StartContainer will be deprecated from docker 1.10. Now in
// docker manager the HostConfig is set when CreateContainer().
// TODO(random-liu): Remove the HostConfig here when it is completely removed in
// docker 1.12.
func (f *FakeDockerClient) StartContainer(id string, _ *docker.HostConfig) error {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "start")
	if err := f.popError("start"); err != nil {
		return err
	}
	container, ok := f.ContainerMap[id]
	if !ok {
		container = &docker.Container{ID: id, Name: id}
	}
	container.State = docker.State{
		Running:   true,
		Pid:       os.Getpid(),
		StartedAt: time.Now(),
	}
	container.NetworkSettings = &docker.NetworkSettings{IPAddress: "2.3.4.5"}
	f.ContainerMap[id] = container
	f.updateContainerStatus(id, statusRunningPrefix)
	f.normalSleep(200, 50, 50)
	return nil
}

// StopContainer is a test-spy implementation of DockerInterface.StopContainer.
// It adds an entry "stop" to the internal method call record.
func (f *FakeDockerClient) StopContainer(id string, timeout uint) error {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "stop")
	if err := f.popError("stop"); err != nil {
		return err
	}
	f.Stopped = append(f.Stopped, id)
	// Container status should be Updated before container moved to ExitedContainerList
	f.updateContainerStatus(id, statusExitedPrefix)
	var newList []docker.APIContainers
	for _, container := range f.ContainerList {
		if container.ID == id {
			// The newest exited container should be in front. Because we assume so in GetPodStatus()
			f.ExitedContainerList = append([]docker.APIContainers{container}, f.ExitedContainerList...)
			continue
		}
		newList = append(newList, container)
	}
	f.ContainerList = newList
	container, ok := f.ContainerMap[id]
	if !ok {
		container = &docker.Container{
			ID:   id,
			Name: id,
			State: docker.State{
				Running:    false,
				StartedAt:  time.Now().Add(-time.Second),
				FinishedAt: time.Now(),
			},
		}
	} else {
		container.State.FinishedAt = time.Now()
		container.State.Running = false
	}
	f.ContainerMap[id] = container
	f.normalSleep(200, 50, 50)
	return nil
}

func (f *FakeDockerClient) RemoveContainer(opts docker.RemoveContainerOptions) error {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "remove")
	err := f.popError("remove")
	if err != nil {
		return err
	}
	for i := range f.ExitedContainerList {
		if f.ExitedContainerList[i].ID == opts.ID {
			delete(f.ContainerMap, opts.ID)
			f.ExitedContainerList = append(f.ExitedContainerList[:i], f.ExitedContainerList[i+1:]...)
			f.Removed = append(f.Removed, opts.ID)
			return nil
		}

	}
	// To be a good fake, report error if container is not stopped.
	return fmt.Errorf("container not stopped")
}

// Logs is a test-spy implementation of DockerInterface.Logs.
// It adds an entry "logs" to the internal method call record.
func (f *FakeDockerClient) Logs(opts docker.LogsOptions) error {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "logs")
	return f.popError("logs")
}

// PullImage is a test-spy implementation of DockerInterface.StopContainer.
// It adds an entry "pull" to the internal method call record.
func (f *FakeDockerClient) PullImage(opts docker.PullImageOptions, auth docker.AuthConfiguration) error {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "pull")
	err := f.popError("pull")
	if err == nil {
		registry := opts.Registry
		if len(registry) != 0 {
			registry = registry + "/"
		}
		authJson, _ := json.Marshal(auth)
		f.pulled = append(f.pulled, fmt.Sprintf("%s%s:%s using %s", registry, opts.Repository, opts.Tag, string(authJson)))
	}
	return err
}

func (f *FakeDockerClient) Version() (*docker.Env, error) {
	return &f.VersionInfo, f.popError("version")
}

func (f *FakeDockerClient) Info() (*docker.Env, error) {
	return &f.Information, nil
}

func (f *FakeDockerClient) CreateExec(opts docker.CreateExecOptions) (*docker.Exec, error) {
	f.Lock()
	defer f.Unlock()
	f.execCmd = opts.Cmd
	f.called = append(f.called, "create_exec")
	return &docker.Exec{ID: "12345678"}, nil
}

func (f *FakeDockerClient) StartExec(_ string, _ docker.StartExecOptions) error {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "start_exec")
	return nil
}

func (f *FakeDockerClient) AttachToContainer(opts docker.AttachToContainerOptions) error {
	f.Lock()
	defer f.Unlock()
	f.called = append(f.called, "attach")
	return nil
}

func (f *FakeDockerClient) InspectExec(id string) (*docker.ExecInspect, error) {
	return f.ExecInspect, f.popError("inspect_exec")
}

func (f *FakeDockerClient) ListImages(opts docker.ListImagesOptions) ([]docker.APIImages, error) {
	err := f.popError("list_images")
	return f.Images, err
}

func (f *FakeDockerClient) RemoveImage(image string) error {
	err := f.popError("remove_image")
	if err == nil {
		f.RemovedImages.Insert(image)
	}
	return err
}

func (f *FakeDockerClient) updateContainerStatus(id, status string) {
	for i := range f.ContainerList {
		if f.ContainerList[i].ID == id {
			f.ContainerList[i].Status = status
		}
	}
}

// FakeDockerPuller is a stub implementation of DockerPuller.
type FakeDockerPuller struct {
	sync.Mutex

	HasImages    []string
	ImagesPulled []string

	// Every pull will return the first error here, and then reslice
	// to remove it. Will give nil errors if this slice is empty.
	ErrorsToInject []error
}

// Pull records the image pull attempt, and optionally injects an error.
func (f *FakeDockerPuller) Pull(image string, secrets []api.Secret) (err error) {
	f.Lock()
	defer f.Unlock()
	f.ImagesPulled = append(f.ImagesPulled, image)

	if len(f.ErrorsToInject) > 0 {
		err = f.ErrorsToInject[0]
		f.ErrorsToInject = f.ErrorsToInject[1:]
	}
	return err
}

func (f *FakeDockerPuller) IsImagePresent(name string) (bool, error) {
	f.Lock()
	defer f.Unlock()
	if f.HasImages == nil {
		return true, nil
	}
	for _, s := range f.HasImages {
		if s == name {
			return true, nil
		}
	}
	return false, nil
}
