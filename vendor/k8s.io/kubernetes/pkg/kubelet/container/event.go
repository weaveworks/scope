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

package container

const (
	// Container event reason list
	CreatedContainer        = "Created"
	StartedContainer        = "Started"
	FailedToCreateContainer = "Failed"
	FailedToStartContainer  = "Failed"
	KillingContainer        = "Killing"
	BackOffStartContainer   = "BackOff"

	// Image event reason list
	PullingImage            = "Pulling"
	PulledImage             = "Pulled"
	FailedToPullImage       = "Failed"
	FailedToInspectImage    = "InspectFailed"
	ErrImageNeverPullPolicy = "ErrImageNeverPull"
	BackOffPullImage        = "BackOff"

	// kubelet event reason list
	NodeReady               = "NodeReady"
	NodeNotReady            = "NodeNotReady"
	NodeSchedulable         = "NodeSchedulable"
	NodeNotSchedulable      = "NodeNotSchedulable"
	StartingKubelet         = "Starting"
	KubeletSetupFailed      = "KubeletSetupFailed"
	FailedMountVolume       = "FailedMount"
	HostPortConflict        = "HostPortConflict"
	NodeSelectorMismatching = "NodeSelectorMismatching"
	InsufficientFreeCPU     = "InsufficientFreeCPU"
	InsufficientFreeMemory  = "InsufficientFreeMemory"
	OutOfDisk               = "OutOfDisk"
	HostNetworkNotSupported = "HostNetworkNotSupported"
	UndefinedShaper         = "NilShaper"
	NodeRebooted            = "Rebooted"

	// Image manager event reason list
	InvalidDiskCapacity = "InvalidDiskCapacity"
	FreeDiskSpaceFailed = "FreeDiskSpaceFailed"

	// Probe event reason list
	ContainerUnhealthy = "Unhealthy"

	// Pod worker event reason list
	FailedSync = "FailedSync"

	// Config event reason list
	FailedValidation = "FailedValidation"
)
