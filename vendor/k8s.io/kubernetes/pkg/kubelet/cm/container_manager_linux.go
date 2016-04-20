// +build linux

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
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs"
	"github.com/opencontainers/runc/libcontainer/configs"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/kubelet/cadvisor"
	"k8s.io/kubernetes/pkg/util"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/util/oom"
	"k8s.io/kubernetes/pkg/util/sets"
	utilsysctl "k8s.io/kubernetes/pkg/util/sysctl"
	"k8s.io/kubernetes/pkg/util/wait"
)

const (
	// The percent of the machine memory capacity. The value is used to calculate
	// docker memory resource container's hardlimit to workaround docker memory
	// leakage issue. Please see kubernetes/issues/9881 for more detail.
	DockerMemoryLimitThresholdPercent = 70
	// The minimum memory limit allocated to docker container: 150Mi
	MinDockerMemoryLimit = 150 * 1024 * 1024
)

// A non-user container tracked by the Kubelet.
type systemContainer struct {
	// Absolute name of the container.
	name string

	// CPU limit in millicores.
	cpuMillicores int64

	// Function that ensures the state of the container.
	// m is the cgroup manager for the specified container.
	ensureStateFunc func(m *fs.Manager) error

	// Manager for the cgroups of the external container.
	manager *fs.Manager
}

func newSystemCgroups(containerName string) *systemContainer {
	return &systemContainer{
		name:    containerName,
		manager: createManager(containerName),
	}
}

type containerManagerImpl struct {
	sync.RWMutex
	cadvisorInterface cadvisor.Interface
	mountUtil         mount.Interface
	NodeConfig
	status Status
	// External containers being managed.
	systemContainers []*systemContainer
	periodicTasks    []func()
}

type features struct {
	cpuHardcapping bool
}

var _ ContainerManager = &containerManagerImpl{}

// checks if the required cgroups subsystems are mounted.
// As of now, only 'cpu' and 'memory' are required.
// cpu quota is a soft requirement.
func validateSystemRequirements(mountUtil mount.Interface) (features, error) {
	const (
		cgroupMountType = "cgroup"
		localErr        = "system validation failed"
	)
	var (
		cpuMountPoint string
		f             features
	)
	mountPoints, err := mountUtil.List()
	if err != nil {
		return f, fmt.Errorf("%s - %v", localErr, err)
	}

	expectedCgroups := sets.NewString("cpu", "cpuacct", "cpuset", "memory")
	for _, mountPoint := range mountPoints {
		if mountPoint.Type == cgroupMountType {
			for _, opt := range mountPoint.Opts {
				if expectedCgroups.Has(opt) {
					expectedCgroups.Delete(opt)
				}
				if opt == "cpu" {
					cpuMountPoint = mountPoint.Path
				}
			}
		}
	}

	if expectedCgroups.Len() > 0 {
		return f, fmt.Errorf("%s - Following Cgroup subsystem not mounted: %v", localErr, expectedCgroups.List())
	}

	// Check if cpu quota is available.
	// CPU cgroup is required and so it expected to be mounted at this point.
	periodExists, err := util.FileExists(path.Join(cpuMountPoint, "cpu.cfs_period_us"))
	if err != nil {
		glog.Errorf("failed to detect if CPU cgroup cpu.cfs_period_us is available - %v", err)
	}
	quotaExists, err := util.FileExists(path.Join(cpuMountPoint, "cpu.cfs_quota_us"))
	if err != nil {
		glog.Errorf("failed to detect if CPU cgroup cpu.cfs_quota_us is available - %v", err)
	}
	if quotaExists && periodExists {
		f.cpuHardcapping = true
	}
	return f, nil
}

// TODO(vmarmol): Add limits to the system containers.
// Takes the absolute name of the specified containers.
// Empty container name disables use of the specified container.
func NewContainerManager(mountUtil mount.Interface, cadvisorInterface cadvisor.Interface, nodeConfig NodeConfig) (ContainerManager, error) {
	return &containerManagerImpl{
		cadvisorInterface: cadvisorInterface,
		mountUtil:         mountUtil,
		NodeConfig:        nodeConfig,
	}, nil
}

// Create a cgroup container manager.
func createManager(containerName string) *fs.Manager {
	return &fs.Manager{
		Cgroups: &configs.Cgroup{
			Parent: "/",
			Name:   containerName,
			Resources: &configs.Resources{
				AllowAllDevices: true,
			},
		},
	}
}

// TODO: plumb this up as a flag to Kubelet in a future PR
type KernelTunableBehavior string

const (
	KernelTunableWarn   KernelTunableBehavior = "warn"
	KernelTunableError  KernelTunableBehavior = "error"
	KernelTunableModify KernelTunableBehavior = "modify"
)

// setupKernelTunables validates kernel tunable flags are set as expected
// depending upon the specified option, it will either warn, error, or modify the kernel tunable flags
func setupKernelTunables(option KernelTunableBehavior) error {
	desiredState := map[string]int{
		utilsysctl.VmOvercommitMemory: utilsysctl.VmOvercommitMemoryAlways,
		utilsysctl.VmPanicOnOOM:       utilsysctl.VmPanicOnOOMInvokeOOMKiller,
		utilsysctl.KernelPanic:        utilsysctl.KernelPanicRebootTimeout,
		utilsysctl.KernelPanicOnOops:  utilsysctl.KernelPanicOnOopsAlways,
	}

	errList := []error{}
	for flag, expectedValue := range desiredState {
		val, err := utilsysctl.GetSysctl(flag)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		if val == expectedValue {
			continue
		}

		switch option {
		case KernelTunableError:
			errList = append(errList, fmt.Errorf("Invalid kernel flag: %v, expected value: %v, actual value: %v", flag, expectedValue, val))
		case KernelTunableWarn:
			glog.V(2).Infof("Invalid kernel flag: %v, expected value: %v, actual value: %v", flag, expectedValue, val)
		case KernelTunableModify:
			glog.V(2).Infof("Updating kernel flag: %v, expected value: %v, actual value: %v", flag, expectedValue, val)
			err = utilsysctl.SetSysctl(flag, expectedValue)
			if err != nil {
				errList = append(errList, err)
			}
		}
	}
	return utilerrors.NewAggregate(errList)
}

func (cm *containerManagerImpl) setupNode() error {
	f, err := validateSystemRequirements(cm.mountUtil)
	if err != nil {
		return err
	}
	if !f.cpuHardcapping {
		cm.status.SoftRequirements = fmt.Errorf("CPU hardcapping unsupported")
	}
	// TODO: plumb kernel tunable options into container manager, right now, we modify by default
	if err := setupKernelTunables(KernelTunableModify); err != nil {
		return err
	}

	systemContainers := []*systemContainer{}
	if cm.ContainerRuntime == "docker" {
		if cm.RuntimeCgroupsName != "" {
			cont := newSystemCgroups(cm.RuntimeCgroupsName)
			info, err := cm.cadvisorInterface.MachineInfo()
			var capacity = api.ResourceList{}
			if err != nil {
			} else {
				capacity = cadvisor.CapacityFromMachineInfo(info)
			}
			memoryLimit := (int64(capacity.Memory().Value() * DockerMemoryLimitThresholdPercent / 100))
			if memoryLimit < MinDockerMemoryLimit {
				glog.Warningf("Memory limit %d for container %s is too small, reset it to %d", memoryLimit, cm.RuntimeCgroupsName, MinDockerMemoryLimit)
				memoryLimit = MinDockerMemoryLimit
			}

			glog.V(2).Infof("Configure resource-only container %s with memory limit: %d", cm.RuntimeCgroupsName, memoryLimit)

			dockerContainer := &fs.Manager{
				Cgroups: &configs.Cgroup{
					Parent: "/",
					Name:   cm.RuntimeCgroupsName,
					Resources: &configs.Resources{
						Memory:          memoryLimit,
						MemorySwap:      -1,
						AllowAllDevices: true,
					},
				},
			}
			cont.ensureStateFunc = func(manager *fs.Manager) error {
				return ensureDockerInContainer(cm.cadvisorInterface, -900, dockerContainer)
			}
			systemContainers = append(systemContainers, cont)
		} else {
			cm.periodicTasks = append(cm.periodicTasks, func() {
				cont, err := getContainerNameForProcess("docker")
				if err != nil {
					glog.Error(err)
					return
				}
				cm.Lock()
				defer cm.Unlock()
				cm.RuntimeCgroupsName = cont
			})
		}
	}

	if cm.SystemCgroupsName != "" {
		if cm.SystemCgroupsName == "/" {
			return fmt.Errorf("system container cannot be root (\"/\")")
		}
		cont := newSystemCgroups(cm.SystemCgroupsName)
		rootContainer := &fs.Manager{
			Cgroups: &configs.Cgroup{
				Parent: "/",
				Name:   "/",
			},
		}
		cont.ensureStateFunc = func(manager *fs.Manager) error {
			return ensureSystemCgroups(rootContainer, manager)
		}
		systemContainers = append(systemContainers, cont)
	}

	if cm.KubeletCgroupsName != "" {
		cont := newSystemCgroups(cm.KubeletCgroupsName)
		manager := fs.Manager{
			Cgroups: &configs.Cgroup{
				Parent: "/",
				Name:   cm.KubeletCgroupsName,
				Resources: &configs.Resources{
					AllowAllDevices: true,
				},
			},
		}
		cont.ensureStateFunc = func(_ *fs.Manager) error {
			return manager.Apply(os.Getpid())
		}
		systemContainers = append(systemContainers, cont)
	} else {
		cm.periodicTasks = append(cm.periodicTasks, func() {
			cont, err := getContainer(os.Getpid())
			if err != nil {
				glog.Errorf("failed to find cgroups of kubelet - %v", err)
				return
			}
			cm.Lock()
			defer cm.Unlock()

			cm.KubeletCgroupsName = cont
		})
	}

	cm.systemContainers = systemContainers
	return nil
}

func getContainerNameForProcess(name string) (string, error) {
	pids, err := getPidsForProcess(name)
	if err != nil {
		return "", fmt.Errorf("failed to detect process id for %q - %v", name, err)
	}
	if len(pids) == 0 {
		return "", nil
	}
	cont, err := getContainer(pids[0])
	if err != nil {
		return "", err
	}
	return cont, nil
}

func (cm *containerManagerImpl) GetNodeConfig() NodeConfig {
	cm.RLock()
	defer cm.RUnlock()
	return cm.NodeConfig
}

func (cm *containerManagerImpl) Status() Status {
	cm.RLock()
	defer cm.RUnlock()
	return cm.status
}

func (cm *containerManagerImpl) Start() error {
	// Setup the node
	if err := cm.setupNode(); err != nil {
		return err
	}
	// Don't run a background thread if there are no ensureStateFuncs.
	numEnsureStateFuncs := 0
	for _, cont := range cm.systemContainers {
		if cont.ensureStateFunc != nil {
			numEnsureStateFuncs++
		}
	}
	if numEnsureStateFuncs >= 0 {
		go wait.Until(func() {
			for _, cont := range cm.systemContainers {
				if cont.ensureStateFunc != nil {
					if err := cont.ensureStateFunc(cont.manager); err != nil {
						glog.Warningf("[ContainerManager] Failed to ensure state of %q: %v", cont.name, err)
					}
				}
			}
		}, time.Minute, wait.NeverStop)

	}

	// Run ensure state functions every minute.
	if len(cm.periodicTasks) > 0 {
		go wait.Until(func() {
			for _, task := range cm.periodicTasks {
				if task != nil {
					task()
				}
			}
		}, 5*time.Minute, wait.NeverStop)
	}

	return nil
}

func (cm *containerManagerImpl) SystemCgroupsLimit() api.ResourceList {
	cpuLimit := int64(0)

	// Sum up resources of all external containers.
	for _, cont := range cm.systemContainers {
		cpuLimit += cont.cpuMillicores
	}

	return api.ResourceList{
		api.ResourceCPU: *resource.NewMilliQuantity(
			cpuLimit,
			resource.DecimalSI),
	}
}

func isProcessRunningInHost(pid int) (bool, error) {
	// Get init mount namespace. Mount namespace is unique for all containers.
	initMntNs, err := os.Readlink("/proc/1/ns/mnt")
	if err != nil {
		return false, fmt.Errorf("failed to find mount namespace of init process")
	}
	processMntNs, err := os.Readlink(fmt.Sprintf("/proc/%d/ns/mnt", pid))
	if err != nil {
		return false, fmt.Errorf("failed to find mount namespace of process %q", pid)
	}
	return initMntNs == processMntNs, nil
}

func getPidsForProcess(name string) ([]int, error) {
	out, err := exec.Command("pidof", name).Output()
	if err != nil {
		return []int{}, fmt.Errorf("failed to find pid of %q: %v", name, err)
	}

	// The output of pidof is a list of pids.
	pids := []int{}
	for _, pidStr := range strings.Split(strings.TrimSpace(string(out)), " ") {
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

// Ensures that the Docker daemon is in the desired container.
func ensureDockerInContainer(cadvisor cadvisor.Interface, oomScoreAdj int, manager *fs.Manager) error {
	pids, err := getPidsForProcess("docker")
	if err != nil {
		return err
	}
	// Move if the pid is not already in the desired container.
	errs := []error{}
	for _, pid := range pids {
		if runningInHost, err := isProcessRunningInHost(pid); err != nil {
			errs = append(errs, err)
			// Err on the side of caution. Avoid moving the docker daemon unless we are able to identify its context.
			continue
		} else if !runningInHost {
			// Docker daemon is running inside a container. Don't touch that.
			continue
		}

		cont, err := getContainer(pid)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to find container of PID %d: %v", pid, err))
		}

		if cont != manager.Cgroups.Name {
			err = manager.Apply(pid)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to move PID %d (in %q) to %q", pid, cont, manager.Cgroups.Name))
			}
		}

		// Also apply oom-score-adj to processes
		oomAdjuster := oom.NewOOMAdjuster()
		if err := oomAdjuster.ApplyOOMScoreAdj(pid, oomScoreAdj); err != nil {
			errs = append(errs, fmt.Errorf("failed to apply oom score %d to PID %d", oomScoreAdj, pid))
		}
	}

	return utilerrors.NewAggregate(errs)
}

// Gets the (CPU) container the specified pid is in.
func getContainer(pid int) (string, error) {
	cgs, err := cgroups.ParseCgroupFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return "", err
	}

	cg, ok := cgs["cpu"]
	if ok {
		return cg, nil
	}

	return "", cgroups.NewNotFoundError("cpu")
}

// Ensures the system container is created and all non-kernel threads and process 1
// without a container are moved to it.
//
// The reason of leaving kernel threads at root cgroup is that we don't want to tie the
// execution of these threads with to-be defined /system quota and create priority inversions.
//
func ensureSystemCgroups(rootContainer *fs.Manager, manager *fs.Manager) error {
	// Move non-kernel PIDs to the system container.
	attemptsRemaining := 10
	var errs []error
	for attemptsRemaining >= 0 {
		// Only keep errors on latest attempt.
		errs = []error{}
		attemptsRemaining--

		allPids, err := rootContainer.GetPids()
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to list PIDs for root: %v", err))
			continue
		}

		// Remove kernel pids and other protected PIDs (pid 1, PIDs already in system & kubelet containers)
		pids := make([]int, 0, len(allPids))
		for _, pid := range allPids {
			if pid == 1 || isKernelPid(pid) {
				continue
			}

			pids = append(pids, pid)
		}
		glog.Infof("Found %d PIDs in root, %d of them are not to be moved", len(allPids), len(allPids)-len(pids))

		// Check if we have moved all the non-kernel PIDs.
		if len(pids) == 0 {
			break
		}

		glog.Infof("Moving non-kernel processes: %v", pids)
		for _, pid := range pids {
			err := manager.Apply(pid)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to move PID %d into the system container %q: %v", pid, manager.Cgroups.Name, err))
			}
		}

	}
	if attemptsRemaining < 0 {
		errs = append(errs, fmt.Errorf("ran out of attempts to create system containers %q", manager.Cgroups.Name))
	}

	return utilerrors.NewAggregate(errs)
}

// Determines whether the specified PID is a kernel PID.
func isKernelPid(pid int) bool {
	// Kernel threads have no associated executable.
	_, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	return err != nil
}
