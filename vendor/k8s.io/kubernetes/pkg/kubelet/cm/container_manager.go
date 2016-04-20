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

package cm

import (
	"k8s.io/kubernetes/pkg/api"
)

// Manages the containers running on a machine.
type ContainerManager interface {
	// Runs the container manager's housekeeping.
	// - Ensures that the Docker daemon is in a container.
	// - Creates the system container where all non-containerized processes run.
	Start() error

	// Returns resources allocated to system cgroups in the machine.
	// These cgroups include the system and Kubernetes services.
	SystemCgroupsLimit() api.ResourceList

	// Returns a NodeConfig that is being used by the container manager.
	GetNodeConfig() NodeConfig

	// Returns internal Status.
	Status() Status
}

type NodeConfig struct {
	RuntimeCgroupsName string
	SystemCgroupsName  string
	KubeletCgroupsName string
	ContainerRuntime   string
}

type Status struct {
	// Any soft requirements that were unsatisfied.
	SoftRequirements error
}
