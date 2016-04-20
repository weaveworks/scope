/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package container

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/types"
	"k8s.io/kubernetes/pkg/util/flowcontrol"
	"k8s.io/kubernetes/pkg/volume"
)

type Version interface {
	// Compare compares two versions of the runtime. On success it returns -1
	// if the version is less than the other, 1 if it is greater than the other,
	// or 0 if they are equal.
	Compare(other string) (int, error)
	// String returns a string that represents the version.
	String() string
}

// ImageSpec is an internal representation of an image.  Currently, it wraps the
// value of a Container's Image field, but in the future it will include more detailed
// information about the different image types.
type ImageSpec struct {
	Image string
}

// Runtime interface defines the interfaces that should be implemented
// by a container runtime.
// Thread safety is required from implementations of this interface.
type Runtime interface {
	// Type returns the type of the container runtime.
	Type() string

	// Version returns the version information of the container runtime.
	Version() (Version, error)
	// APIVersion returns the API version information of the container
	// runtime. This may be different from the runtime engine's version.
	// TODO(random-liu): We should fold this into Version()
	APIVersion() (Version, error)
	// Status returns error if the runtime is unhealthy; nil otherwise.
	Status() error
	// GetPods returns a list containers group by pods. The boolean parameter
	// specifies whether the runtime returns all containers including those already
	// exited and dead containers (used for garbage collection).
	GetPods(all bool) ([]*Pod, error)
	// GarbageCollect removes dead containers using the specified container gc policy
	GarbageCollect(gcPolicy ContainerGCPolicy) error
	// Syncs the running pod into the desired pod.
	SyncPod(pod *api.Pod, apiPodStatus api.PodStatus, podStatus *PodStatus, pullSecrets []api.Secret, backOff *flowcontrol.Backoff) PodSyncResult
	// KillPod kills all the containers of a pod. Pod may be nil, running pod must not be.
	// TODO(random-liu): Return PodSyncResult in KillPod.
	KillPod(pod *api.Pod, runningPod Pod) error
	// GetPodStatus retrieves the status of the pod, including the
	// information of all containers in the pod that are visble in Runtime.
	GetPodStatus(uid types.UID, name, namespace string) (*PodStatus, error)
	// PullImage pulls an image from the network to local storage using the supplied
	// secrets if necessary.
	PullImage(image ImageSpec, pullSecrets []api.Secret) error
	// IsImagePresent checks whether the container image is already in the local storage.
	IsImagePresent(image ImageSpec) (bool, error)
	// Gets all images currently on the machine.
	ListImages() ([]Image, error)
	// Removes the specified image.
	RemoveImage(image ImageSpec) error
	// TODO(vmarmol): Unify pod and containerID args.
	// GetContainerLogs returns logs of a specific container. By
	// default, it returns a snapshot of the container log. Set 'follow' to true to
	// stream the log. Set 'follow' to false and specify the number of lines (e.g.
	// "100" or "all") to tail the log.
	GetContainerLogs(pod *api.Pod, containerID ContainerID, logOptions *api.PodLogOptions, stdout, stderr io.Writer) (err error)
	// ContainerCommandRunner encapsulates the command runner interfaces for testability.
	ContainerCommandRunner
	// ContainerAttach encapsulates the attaching to containers for testability
	ContainerAttacher
}

type ContainerAttacher interface {
	AttachContainer(id ContainerID, stdin io.Reader, stdout, stderr io.WriteCloser, tty bool) (err error)
}

// CommandRunner encapsulates the command runner interfaces for testability.
type ContainerCommandRunner interface {
	// TODO(vmarmol): Merge RunInContainer and ExecInContainer.
	// Runs the command in the container of the specified pod.
	RunInContainer(containerID ContainerID, cmd []string) ([]byte, error)
	// Runs the command in the container of the specified pod using nsenter.
	// Attaches the processes stdin, stdout, and stderr. Optionally uses a
	// tty.
	ExecInContainer(containerID ContainerID, cmd []string, stdin io.Reader, stdout, stderr io.WriteCloser, tty bool) error
	// Forward the specified port from the specified pod to the stream.
	PortForward(pod *Pod, port uint16, stream io.ReadWriteCloser) error
}

// ImagePuller wraps Runtime.PullImage() to pull a container image.
// It will check the presence of the image, and report the 'image pulling',
// 'image pulled' events correspondingly.
type ImagePuller interface {
	PullImage(pod *api.Pod, container *api.Container, pullSecrets []api.Secret) (error, string)
}

// Pod is a group of containers.
type Pod struct {
	// The ID of the pod, which can be used to retrieve a particular pod
	// from the pod list returned by GetPods().
	ID types.UID
	// The name and namespace of the pod, which is readable by human.
	Name      string
	Namespace string
	// List of containers that belongs to this pod. It may contain only
	// running containers, or mixed with dead ones (when GetPods(true)).
	Containers []*Container
}

// PodPair contains both runtime#Pod and api#Pod
type PodPair struct {
	// APIPod is the api.Pod
	APIPod *api.Pod
	// RunningPod is the pod defined defined in pkg/kubelet/container/runtime#Pod
	RunningPod *Pod
}

// ContainerID is a type that identifies a container.
type ContainerID struct {
	// The type of the container runtime. e.g. 'docker', 'rkt'.
	Type string
	// The identification of the container, this is comsumable by
	// the underlying container runtime. (Note that the container
	// runtime interface still takes the whole struct as input).
	ID string
}

func BuildContainerID(typ, ID string) ContainerID {
	return ContainerID{Type: typ, ID: ID}
}

// Convenience method for creating a ContainerID from an ID string.
func ParseContainerID(containerID string) ContainerID {
	var id ContainerID
	if err := id.ParseString(containerID); err != nil {
		glog.Error(err)
	}
	return id
}

func (c *ContainerID) ParseString(data string) error {
	// Trim the quotes and split the type and ID.
	parts := strings.Split(strings.Trim(data, "\""), "://")
	if len(parts) != 2 {
		return fmt.Errorf("invalid container ID: %q", data)
	}
	c.Type, c.ID = parts[0], parts[1]
	return nil
}

func (c *ContainerID) String() string {
	return fmt.Sprintf("%s://%s", c.Type, c.ID)
}

func (c *ContainerID) IsEmpty() bool {
	return *c == ContainerID{}
}

func (c *ContainerID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", c.String())), nil
}

func (c *ContainerID) UnmarshalJSON(data []byte) error {
	return c.ParseString(string(data))
}

// DockerID is an ID of docker container. It is a type to make it clear when we're working with docker container Ids
type DockerID string

func (id DockerID) ContainerID() ContainerID {
	return ContainerID{
		Type: "docker",
		ID:   string(id),
	}
}

type ContainerState string

const (
	ContainerStateRunning ContainerState = "running"
	ContainerStateExited  ContainerState = "exited"
	// This unknown encompasses all the states that we currently don't care.
	ContainerStateUnknown ContainerState = "unknown"
)

// Container provides the runtime information for a container, such as ID, hash,
// state of the container.
type Container struct {
	// The ID of the container, used by the container runtime to identify
	// a container.
	ID ContainerID
	// The name of the container, which should be the same as specified by
	// api.Container.
	Name string
	// The image name of the container, this also includes the tag of the image,
	// the expected form is "NAME:TAG".
	Image string
	// Hash of the container, used for comparison. Optional for containers
	// not managed by kubelet.
	Hash uint64
	// The timestamp of the creation time of the container.
	// TODO(yifan): Consider to move it to api.ContainerStatus.
	Created int64
	// State is the state of the container.
	State ContainerState
}

// PodStatus represents the status of the pod and its containers.
// api.PodStatus can be derived from examining PodStatus and api.Pod.
type PodStatus struct {
	// ID of the pod.
	ID types.UID
	// Name of the pod.
	Name string
	// Namspace of the pod.
	Namespace string
	// IP of the pod.
	IP string
	// Status of containers in the pod.
	ContainerStatuses []*ContainerStatus
}

// ContainerStatus represents the status of a container.
type ContainerStatus struct {
	// ID of the container.
	ID ContainerID
	// Name of the container.
	Name string
	// Status of the container.
	State ContainerState
	// Creation time of the container.
	CreatedAt time.Time
	// Start time of the container.
	StartedAt time.Time
	// Finish time of the container.
	FinishedAt time.Time
	// Exit code of the container.
	ExitCode int
	// Name of the image, this also includes the tag of the image,
	// the expected form is "NAME:TAG".
	Image string
	// ID of the image.
	ImageID string
	// Hash of the container, used for comparison.
	Hash uint64
	// Number of times that the container has been restarted.
	RestartCount int
	// A string explains why container is in such a status.
	Reason string
	// Message written by the container before exiting (stored in
	// TerminationMessagePath).
	Message string
}

// FindContainerStatusByName returns container status in the pod status with the given name.
// When there are multiple containers' statuses with the same name, the first match will be returned.
func (podStatus *PodStatus) FindContainerStatusByName(containerName string) *ContainerStatus {
	for _, containerStatus := range podStatus.ContainerStatuses {
		if containerStatus.Name == containerName {
			return containerStatus
		}
	}
	return nil
}

// Get container status of all the running containers in a pod
func (podStatus *PodStatus) GetRunningContainerStatuses() []*ContainerStatus {
	runnningContainerStatues := []*ContainerStatus{}
	for _, containerStatus := range podStatus.ContainerStatuses {
		if containerStatus.State == ContainerStateRunning {
			runnningContainerStatues = append(runnningContainerStatues, containerStatus)
		}
	}
	return runnningContainerStatues
}

// Basic information about a container image.
type Image struct {
	// ID of the image.
	ID string
	// Other names by which this image is known.
	RepoTags []string
	// The size of the image in bytes.
	Size int64
}

type EnvVar struct {
	Name  string
	Value string
}

type Mount struct {
	// Name of the volume mount.
	Name string
	// Path of the mount within the container.
	ContainerPath string
	// Path of the mount on the host.
	HostPath string
	// Whether the mount is read-only.
	ReadOnly bool
	// Whether the mount needs SELinux relabeling
	SELinuxRelabel bool
}

type PortMapping struct {
	// Name of the port mapping
	Name string
	// Protocol of the port mapping.
	Protocol api.Protocol
	// The port number within the container.
	ContainerPort int
	// The port number on the host.
	HostPort int
	// The host IP.
	HostIP string
}

// RunContainerOptions specify the options which are necessary for running containers
type RunContainerOptions struct {
	// The environment variables list.
	Envs []EnvVar
	// The mounts for the containers.
	Mounts []Mount
	// The port mappings for the containers.
	PortMappings []PortMapping
	// If the container has specified the TerminationMessagePath, then
	// this directory will be used to create and mount the log file to
	// container.TerminationMessagePath
	PodContainerDir string
	// The list of DNS servers for the container to use.
	DNS []string
	// The list of DNS search domains.
	DNSSearch []string
	// The parent cgroup to pass to Docker
	CgroupParent string
	// The type of container rootfs
	ReadOnly bool
	// hostname for pod containers
	Hostname string
}

// VolumeInfo contains information about the volume.
type VolumeInfo struct {
	// Mounter is the volume's mounter
	Mounter volume.Mounter
	// SELinuxLabeled indicates whether this volume has had the
	// pod's SELinux label applied to it or not
	SELinuxLabeled bool
}

type VolumeMap map[string]VolumeInfo

type Pods []*Pod

// FindPodByID finds and returns a pod in the pod list by UID. It will return an empty pod
// if not found.
func (p Pods) FindPodByID(podUID types.UID) Pod {
	for i := range p {
		if p[i].ID == podUID {
			return *p[i]
		}
	}
	return Pod{}
}

// FindPodByFullName finds and returns a pod in the pod list by the full name.
// It will return an empty pod if not found.
func (p Pods) FindPodByFullName(podFullName string) Pod {
	for i := range p {
		if BuildPodFullName(p[i].Name, p[i].Namespace) == podFullName {
			return *p[i]
		}
	}
	return Pod{}
}

// FindPod combines FindPodByID and FindPodByFullName, it finds and returns a pod in the
// pod list either by the full name or the pod ID. It will return an empty pod
// if not found.
func (p Pods) FindPod(podFullName string, podUID types.UID) Pod {
	if len(podFullName) > 0 {
		return p.FindPodByFullName(podFullName)
	}
	return p.FindPodByID(podUID)
}

// FindContainerByName returns a container in the pod with the given name.
// When there are multiple containers with the same name, the first match will
// be returned.
func (p *Pod) FindContainerByName(containerName string) *Container {
	for _, c := range p.Containers {
		if c.Name == containerName {
			return c
		}
	}
	return nil
}

func (p *Pod) FindContainerByID(id ContainerID) *Container {
	for _, c := range p.Containers {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// ToAPIPod converts Pod to api.Pod. Note that if a field in api.Pod has no
// corresponding field in Pod, the field would not be populated.
func (p *Pod) ToAPIPod() *api.Pod {
	var pod api.Pod
	pod.UID = p.ID
	pod.Name = p.Name
	pod.Namespace = p.Namespace

	for _, c := range p.Containers {
		var container api.Container
		container.Name = c.Name
		container.Image = c.Image
		pod.Spec.Containers = append(pod.Spec.Containers, container)
	}
	return &pod
}

// IsEmpty returns true if the pod is empty.
func (p *Pod) IsEmpty() bool {
	return reflect.DeepEqual(p, &Pod{})
}

// GetPodFullName returns a name that uniquely identifies a pod.
func GetPodFullName(pod *api.Pod) string {
	// Use underscore as the delimiter because it is not allowed in pod name
	// (DNS subdomain format), while allowed in the container name format.
	return pod.Name + "_" + pod.Namespace
}

// Build the pod full name from pod name and namespace.
func BuildPodFullName(name, namespace string) string {
	return name + "_" + namespace
}

// Parse the pod full name.
func ParsePodFullName(podFullName string) (string, string, error) {
	parts := strings.Split(podFullName, "_")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("failed to parse the pod full name %q", podFullName)
	}
	return parts[0], parts[1], nil
}

// Option is a functional option type for Runtime, useful for
// completely optional settings.
type Option func(Runtime)

// Sort the container statuses by creation time.
type SortContainerStatusesByCreationTime []*ContainerStatus

func (s SortContainerStatusesByCreationTime) Len() int      { return len(s) }
func (s SortContainerStatusesByCreationTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortContainerStatusesByCreationTime) Less(i, j int) bool {
	return s[i].CreatedAt.Before(s[j].CreatedAt)
}
