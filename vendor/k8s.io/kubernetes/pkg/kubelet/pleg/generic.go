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

package pleg

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	"k8s.io/kubernetes/pkg/kubelet/metrics"
	"k8s.io/kubernetes/pkg/types"
	"k8s.io/kubernetes/pkg/util"
	"k8s.io/kubernetes/pkg/util/atomic"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/util/wait"
)

// GenericPLEG is an extremely simple generic PLEG that relies solely on
// periodic listing to discover container changes. It should be be used
// as temporary replacement for container runtimes do not support a proper
// event generator yet.
//
// Note that GenericPLEG assumes that a container would not be created,
// terminated, and garbage collected within one relist period. If such an
// incident happens, GenenricPLEG would miss all events regarding this
// container. In the case of relisting failure, the window may become longer.
// Note that this assumption is not unique -- many kubelet internal components
// rely on terminated containers as tombstones for bookkeeping purposes. The
// garbage collector is implemented to work with such situtations. However, to
// guarantee that kubelet can handle missing container events, it is
// recommended to set the relist period short and have an auxiliary, longer
// periodic sync in kubelet as the safety net.
type GenericPLEG struct {
	// The period for relisting.
	relistPeriod time.Duration
	// The container runtime.
	runtime kubecontainer.Runtime
	// The channel from which the subscriber listens events.
	eventChannel chan *PodLifecycleEvent
	// The internal cache for pod/container information.
	podRecords podRecords
	// Time of the last relisting.
	relistTime atomic.Value
	// Cache for storing the runtime states required for syncing pods.
	cache kubecontainer.Cache
	// For testability.
	clock util.Clock
}

// plegContainerState has a one-to-one mapping to the
// kubecontainer.ContainerState except for the non-existent state. This state
// is introduced here to complete the state transition scenarios.
type plegContainerState string

const (
	plegContainerRunning     plegContainerState = "running"
	plegContainerExited      plegContainerState = "exited"
	plegContainerUnknown     plegContainerState = "unknown"
	plegContainerNonExistent plegContainerState = "non-existent"
)

func convertState(state kubecontainer.ContainerState) plegContainerState {
	switch state {
	case kubecontainer.ContainerStateRunning:
		return plegContainerRunning
	case kubecontainer.ContainerStateExited:
		return plegContainerExited
	case kubecontainer.ContainerStateUnknown:
		return plegContainerUnknown
	default:
		panic(fmt.Sprintf("unrecognized container state: %v", state))
	}
}

type podRecord struct {
	old     *kubecontainer.Pod
	current *kubecontainer.Pod
}

type podRecords map[types.UID]*podRecord

func NewGenericPLEG(runtime kubecontainer.Runtime, channelCapacity int,
	relistPeriod time.Duration, cache kubecontainer.Cache, clock util.Clock) PodLifecycleEventGenerator {
	return &GenericPLEG{
		relistPeriod: relistPeriod,
		runtime:      runtime,
		eventChannel: make(chan *PodLifecycleEvent, channelCapacity),
		podRecords:   make(podRecords),
		cache:        cache,
		clock:        clock,
	}
}

// Returns a channel from which the subscriber can receive PodLifecycleEvent
// events.
// TODO: support multiple subscribers.
func (g *GenericPLEG) Watch() chan *PodLifecycleEvent {
	return g.eventChannel
}

// Start spawns a goroutine to relist periodically.
func (g *GenericPLEG) Start() {
	go wait.Until(g.relist, g.relistPeriod, wait.NeverStop)
}

func (g *GenericPLEG) Healthy() (bool, error) {
	relistTime := g.getRelistTime()
	// TODO: Evaluate if we can reduce this threshold.
	// The threshold needs to be greater than the relisting period + the
	// relisting time, which can vary significantly. Set a conservative
	// threshold so that we don't cause kubelet to be restarted unnecessarily.
	threshold := 2 * time.Minute
	if g.clock.Since(relistTime) > threshold {
		return false, fmt.Errorf("pleg was last seen active at %v", relistTime)
	}
	return true, nil
}

func generateEvent(podID types.UID, cid string, oldState, newState plegContainerState) *PodLifecycleEvent {
	if newState == oldState {
		return nil
	}
	glog.V(4).Infof("GenericPLEG: %v/%v: %v -> %v", podID, cid, oldState, newState)
	switch newState {
	case plegContainerRunning:
		return &PodLifecycleEvent{ID: podID, Type: ContainerStarted, Data: cid}
	case plegContainerExited:
		return &PodLifecycleEvent{ID: podID, Type: ContainerDied, Data: cid}
	case plegContainerUnknown:
		return &PodLifecycleEvent{ID: podID, Type: ContainerChanged, Data: cid}
	case plegContainerNonExistent:
		// We report "ContainerDied" when container was stopped OR removed. We
		// may want to distinguish the two cases in the future.
		switch oldState {
		case plegContainerExited:
			// We already reported that the container died before.
			return &PodLifecycleEvent{ID: podID, Type: ContainerRemoved, Data: cid}
		default:
			// TODO: We may want to generate a ContainerRemoved event as well.
			// It's ok now because no one relies on the ContainerRemoved event.
			return &PodLifecycleEvent{ID: podID, Type: ContainerDied, Data: cid}
		}
	default:
		panic(fmt.Sprintf("unrecognized container state: %v", newState))
	}
}

func (g *GenericPLEG) getRelistTime() time.Time {
	val := g.relistTime.Load()
	if val == nil {
		return time.Time{}
	}
	return val.(time.Time)
}

func (g *GenericPLEG) updateRelisTime(timestamp time.Time) {
	g.relistTime.Store(timestamp)
}

// relist queries the container runtime for list of pods/containers, compare
// with the internal pods/containers, and generats events accordingly.
func (g *GenericPLEG) relist() {
	glog.V(5).Infof("GenericPLEG: Relisting")

	if lastRelistTime := g.getRelistTime(); !lastRelistTime.IsZero() {
		metrics.PLEGRelistInterval.Observe(metrics.SinceInMicroseconds(lastRelistTime))
	}

	timestamp := g.clock.Now()
	// Update the relist time.
	g.updateRelisTime(timestamp)
	defer func() {
		metrics.PLEGRelistLatency.Observe(metrics.SinceInMicroseconds(timestamp))
	}()

	// Get all the pods.
	podList, err := g.runtime.GetPods(true)
	if err != nil {
		glog.Errorf("GenericPLEG: Unable to retrieve pods: %v", err)
		return
	}
	pods := kubecontainer.Pods(podList)
	g.podRecords.setCurrent(pods)

	// Compare the old and the current pods, and generate events.
	eventsByPodID := map[types.UID][]*PodLifecycleEvent{}
	for pid := range g.podRecords {
		oldPod := g.podRecords.getOld(pid)
		pod := g.podRecords.getCurrent(pid)
		// Get all containers in the old and the new pod.
		allContainers := getContainersFromPods(oldPod, pod)
		for _, container := range allContainers {
			e := computeEvent(oldPod, pod, &container.ID)
			updateEvents(eventsByPodID, e)
		}
	}

	// If there are events associated with a pod, we should update the
	// podCache.
	for pid, events := range eventsByPodID {
		pod := g.podRecords.getCurrent(pid)
		if g.cacheEnabled() {
			// updateCache() will inspect the pod and update the cache. If an
			// error occurs during the inspection, we want PLEG to retry again
			// in the next relist. To achieve this, we do not update the
			// associated podRecord of the pod, so that the change will be
			// detect again in the next relist.
			// TODO: If many pods changed during the same relist period,
			// inspecting the pod and getting the PodStatus to update the cache
			// serially may take a while. We should be aware of this and
			// parallelize if needed.
			if err := g.updateCache(pod, pid); err != nil {
				glog.Errorf("PLEG: Ignoring events for pod %s/%s: %v", pod.Name, pod.Namespace, err)
				continue
			}
		}
		// Update the internal storage and send out the events.
		g.podRecords.update(pid)
		for i := range events {
			// Filter out events that are not reliable and no other components use yet.
			if events[i].Type == ContainerChanged || events[i].Type == ContainerRemoved {
				continue
			}
			g.eventChannel <- events[i]
		}
	}

	if g.cacheEnabled() {
		// Update the cache timestamp.  This needs to happen *after*
		// all pods have been properly updated in the cache.
		g.cache.UpdateTime(timestamp)
	}
}

func getContainersFromPods(pods ...*kubecontainer.Pod) []*kubecontainer.Container {
	cidSet := sets.NewString()
	var containers []*kubecontainer.Container
	for _, p := range pods {
		if p == nil {
			continue
		}
		for _, c := range p.Containers {
			cid := string(c.ID.ID)
			if cidSet.Has(cid) {
				continue
			}
			cidSet.Insert(cid)
			containers = append(containers, c)
		}
	}
	return containers
}

func computeEvent(oldPod, newPod *kubecontainer.Pod, cid *kubecontainer.ContainerID) *PodLifecycleEvent {
	var pid types.UID
	if oldPod != nil {
		pid = oldPod.ID
	} else if newPod != nil {
		pid = newPod.ID
	}
	oldState := getContainerState(oldPod, cid)
	newState := getContainerState(newPod, cid)
	return generateEvent(pid, cid.ID, oldState, newState)
}

func (g *GenericPLEG) cacheEnabled() bool {
	return g.cache != nil
}

func (g *GenericPLEG) updateCache(pod *kubecontainer.Pod, pid types.UID) error {
	if pod == nil {
		// The pod is missing in the current relist. This means that
		// the pod has no visible (active or inactive) containers.
		glog.V(4).Infof("PLEG: Delete status for pod %q", string(pid))
		g.cache.Delete(pid)
		return nil
	}
	timestamp := g.clock.Now()
	// TODO: Consider adding a new runtime method
	// GetPodStatus(pod *kubecontainer.Pod) so that Docker can avoid listing
	// all containers again.
	status, err := g.runtime.GetPodStatus(pod.ID, pod.Name, pod.Namespace)
	glog.V(4).Infof("PLEG: Write status for %s/%s: %+v (err: %v)", pod.Name, pod.Namespace, status, err)
	g.cache.Set(pod.ID, status, err, timestamp)
	return err
}

func updateEvents(eventsByPodID map[types.UID][]*PodLifecycleEvent, e *PodLifecycleEvent) {
	if e == nil {
		return
	}
	eventsByPodID[e.ID] = append(eventsByPodID[e.ID], e)
}

func getContainerState(pod *kubecontainer.Pod, cid *kubecontainer.ContainerID) plegContainerState {
	// Default to the non-existent state.
	state := plegContainerNonExistent
	if pod == nil {
		return state
	}
	container := pod.FindContainerByID(*cid)
	if container == nil {
		return state
	}
	return convertState(container.State)
}

func (pr podRecords) getOld(id types.UID) *kubecontainer.Pod {
	r, ok := pr[id]
	if !ok {
		return nil
	}
	return r.old
}

func (pr podRecords) getCurrent(id types.UID) *kubecontainer.Pod {
	r, ok := pr[id]
	if !ok {
		return nil
	}
	return r.current
}

func (pr podRecords) setCurrent(pods []*kubecontainer.Pod) {
	for i := range pr {
		pr[i].current = nil
	}
	for _, pod := range pods {
		if r, ok := pr[pod.ID]; ok {
			r.current = pod
		} else {
			pr[pod.ID] = &podRecord{current: pod}
		}
	}
}

func (pr podRecords) update(id types.UID) {
	r, ok := pr[id]
	if !ok {
		return
	}
	pr.updateInternal(id, r)
}

func (pr podRecords) updateInternal(id types.UID, r *podRecord) {
	if r.current == nil {
		// Pod no longer exists; delete the entry.
		delete(pr, id)
		return
	}
	r.old = r.current
	r.current = nil
}
