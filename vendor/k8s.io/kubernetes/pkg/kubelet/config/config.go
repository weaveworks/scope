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

package config

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/validation"
	"k8s.io/kubernetes/pkg/client/record"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	kubetypes "k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/kubernetes/pkg/kubelet/util/format"
	"k8s.io/kubernetes/pkg/util/config"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/util/validation/field"
)

// PodConfigNotificationMode describes how changes are sent to the update channel.
type PodConfigNotificationMode int

const (
	// PodConfigNotificationUnknown is the default value for
	// PodConfigNotificationMode when uninitialized.
	PodConfigNotificationUnknown = iota
	// PodConfigNotificationSnapshot delivers the full configuration as a SET whenever
	// any change occurs.
	PodConfigNotificationSnapshot
	// PodConfigNotificationSnapshotAndUpdates delivers an UPDATE message whenever pods are
	// changed, and a SET message if there are any additions or removals.
	PodConfigNotificationSnapshotAndUpdates
	// PodConfigNotificationIncremental delivers ADD, UPDATE, REMOVE, RECONCILE to the update channel.
	PodConfigNotificationIncremental
)

// PodConfig is a configuration mux that merges many sources of pod configuration into a single
// consistent structure, and then delivers incremental change notifications to listeners
// in order.
type PodConfig struct {
	pods *podStorage
	mux  *config.Mux

	// the channel of denormalized changes passed to listeners
	updates chan kubetypes.PodUpdate

	// contains the list of all configured sources
	sourcesLock sync.Mutex
	sources     sets.String
}

// NewPodConfig creates an object that can merge many configuration sources into a stream
// of normalized updates to a pod configuration.
func NewPodConfig(mode PodConfigNotificationMode, recorder record.EventRecorder) *PodConfig {
	updates := make(chan kubetypes.PodUpdate, 50)
	storage := newPodStorage(updates, mode, recorder)
	podConfig := &PodConfig{
		pods:    storage,
		mux:     config.NewMux(storage),
		updates: updates,
		sources: sets.String{},
	}
	return podConfig
}

// Channel creates or returns a config source channel.  The channel
// only accepts PodUpdates
func (c *PodConfig) Channel(source string) chan<- interface{} {
	c.sourcesLock.Lock()
	defer c.sourcesLock.Unlock()
	c.sources.Insert(source)
	return c.mux.Channel(source)
}

// SeenAllSources returns true if seenSources contains all sources in the
// config, and also this config has received a SET message from each source.
func (c *PodConfig) SeenAllSources(seenSources sets.String) bool {
	if c.pods == nil {
		return false
	}
	glog.V(6).Infof("Looking for %v, have seen %v", c.sources.List(), seenSources)
	return seenSources.HasAll(c.sources.List()...) && c.pods.seenSources(c.sources.List()...)
}

// Updates returns a channel of updates to the configuration, properly denormalized.
func (c *PodConfig) Updates() <-chan kubetypes.PodUpdate {
	return c.updates
}

// Sync requests the full configuration be delivered to the update channel.
func (c *PodConfig) Sync() {
	c.pods.Sync()
}

// podStorage manages the current pod state at any point in time and ensures updates
// to the channel are delivered in order.  Note that this object is an in-memory source of
// "truth" and on creation contains zero entries.  Once all previously read sources are
// available, then this object should be considered authoritative.
type podStorage struct {
	podLock sync.RWMutex
	// map of source name to pod name to pod reference
	pods map[string]map[string]*api.Pod
	mode PodConfigNotificationMode

	// ensures that updates are delivered in strict order
	// on the updates channel
	updateLock sync.Mutex
	updates    chan<- kubetypes.PodUpdate

	// contains the set of all sources that have sent at least one SET
	sourcesSeenLock sync.Mutex
	sourcesSeen     sets.String

	// the EventRecorder to use
	recorder record.EventRecorder
}

// TODO: PodConfigNotificationMode could be handled by a listener to the updates channel
// in the future, especially with multiple listeners.
// TODO: allow initialization of the current state of the store with snapshotted version.
func newPodStorage(updates chan<- kubetypes.PodUpdate, mode PodConfigNotificationMode, recorder record.EventRecorder) *podStorage {
	return &podStorage{
		pods:        make(map[string]map[string]*api.Pod),
		mode:        mode,
		updates:     updates,
		sourcesSeen: sets.String{},
		recorder:    recorder,
	}
}

// Merge normalizes a set of incoming changes from different sources into a map of all Pods
// and ensures that redundant changes are filtered out, and then pushes zero or more minimal
// updates onto the update channel.  Ensures that updates are delivered in order.
func (s *podStorage) Merge(source string, change interface{}) error {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	seenBefore := s.sourcesSeen.Has(source)
	adds, updates, deletes, reconciles := s.merge(source, change)
	firstSet := !seenBefore && s.sourcesSeen.Has(source)

	// deliver update notifications
	switch s.mode {
	case PodConfigNotificationIncremental:
		if len(deletes.Pods) > 0 {
			s.updates <- *deletes
		}
		if len(adds.Pods) > 0 {
			s.updates <- *adds
		}
		if len(updates.Pods) > 0 {
			s.updates <- *updates
		}
		if firstSet && len(adds.Pods) == 0 && len(updates.Pods) == 0 {
			// Send an empty update when first seeing the source and there are
			// no ADD or UPDATE pods from the source. This signals kubelet that
			// the source is ready.
			s.updates <- *adds
		}
		// Only add reconcile support here, because kubelet doesn't support Snapshot update now.
		if len(reconciles.Pods) > 0 {
			s.updates <- *reconciles
		}

	case PodConfigNotificationSnapshotAndUpdates:
		if len(deletes.Pods) > 0 || len(adds.Pods) > 0 || firstSet {
			s.updates <- kubetypes.PodUpdate{Pods: s.MergedState().([]*api.Pod), Op: kubetypes.SET, Source: source}
		}
		if len(updates.Pods) > 0 {
			s.updates <- *updates
		}

	case PodConfigNotificationSnapshot:
		if len(updates.Pods) > 0 || len(deletes.Pods) > 0 || len(adds.Pods) > 0 || firstSet {
			s.updates <- kubetypes.PodUpdate{Pods: s.MergedState().([]*api.Pod), Op: kubetypes.SET, Source: source}
		}

	case PodConfigNotificationUnknown:
		fallthrough
	default:
		panic(fmt.Sprintf("unsupported PodConfigNotificationMode: %#v", s.mode))
	}

	return nil
}

func (s *podStorage) merge(source string, change interface{}) (adds, updates, deletes, reconciles *kubetypes.PodUpdate) {
	s.podLock.Lock()
	defer s.podLock.Unlock()

	addPods := []*api.Pod{}
	updatePods := []*api.Pod{}
	deletePods := []*api.Pod{}
	reconcilePods := []*api.Pod{}

	pods := s.pods[source]
	if pods == nil {
		pods = make(map[string]*api.Pod)
	}

	// updatePodFunc is the local function which updates the pod cache *oldPods* with new pods *newPods*.
	// After updated, new pod will be stored in the pod cache *pods*.
	// Notice that *pods* and *oldPods* could be the same cache.
	updatePodsFunc := func(newPods []*api.Pod, oldPods, pods map[string]*api.Pod) {
		filtered := filterInvalidPods(newPods, source, s.recorder)
		for _, ref := range filtered {
			name := kubecontainer.GetPodFullName(ref)
			// Annotate the pod with the source before any comparison.
			if ref.Annotations == nil {
				ref.Annotations = make(map[string]string)
			}
			ref.Annotations[kubetypes.ConfigSourceAnnotationKey] = source
			if existing, found := oldPods[name]; found {
				pods[name] = existing
				needUpdate, needReconcile := checkAndUpdatePod(existing, ref)
				if needUpdate {
					updatePods = append(updatePods, existing)
				} else if needReconcile {
					reconcilePods = append(reconcilePods, existing)
				}
				continue
			}
			recordFirstSeenTime(ref)
			pods[name] = ref
			addPods = append(addPods, ref)
		}
	}

	update := change.(kubetypes.PodUpdate)
	switch update.Op {
	case kubetypes.ADD, kubetypes.UPDATE:
		if update.Op == kubetypes.ADD {
			glog.V(4).Infof("Adding new pods from source %s : %v", source, update.Pods)
		} else {
			glog.V(4).Infof("Updating pods from source %s : %v", source, update.Pods)
		}
		updatePodsFunc(update.Pods, pods, pods)

	case kubetypes.REMOVE:
		glog.V(4).Infof("Removing a pod %v", update)
		for _, value := range update.Pods {
			name := kubecontainer.GetPodFullName(value)
			if existing, found := pods[name]; found {
				// this is a delete
				delete(pods, name)
				deletePods = append(deletePods, existing)
				continue
			}
			// this is a no-op
		}

	case kubetypes.SET:
		glog.V(4).Infof("Setting pods for source %s", source)
		s.markSourceSet(source)
		// Clear the old map entries by just creating a new map
		oldPods := pods
		pods = make(map[string]*api.Pod)
		updatePodsFunc(update.Pods, oldPods, pods)
		for name, existing := range oldPods {
			if _, found := pods[name]; !found {
				// this is a delete
				deletePods = append(deletePods, existing)
			}
		}

	default:
		glog.Warningf("Received invalid update type: %v", update)

	}

	s.pods[source] = pods

	adds = &kubetypes.PodUpdate{Op: kubetypes.ADD, Pods: copyPods(addPods), Source: source}
	updates = &kubetypes.PodUpdate{Op: kubetypes.UPDATE, Pods: copyPods(updatePods), Source: source}
	deletes = &kubetypes.PodUpdate{Op: kubetypes.REMOVE, Pods: copyPods(deletePods), Source: source}
	reconciles = &kubetypes.PodUpdate{Op: kubetypes.RECONCILE, Pods: copyPods(reconcilePods), Source: source}

	return adds, updates, deletes, reconciles
}

func (s *podStorage) markSourceSet(source string) {
	s.sourcesSeenLock.Lock()
	defer s.sourcesSeenLock.Unlock()
	s.sourcesSeen.Insert(source)
}

func (s *podStorage) seenSources(sources ...string) bool {
	s.sourcesSeenLock.Lock()
	defer s.sourcesSeenLock.Unlock()
	return s.sourcesSeen.HasAll(sources...)
}

func filterInvalidPods(pods []*api.Pod, source string, recorder record.EventRecorder) (filtered []*api.Pod) {
	names := sets.String{}
	for i, pod := range pods {
		var errlist field.ErrorList
		if errs := validation.ValidatePod(pod); len(errs) != 0 {
			errlist = append(errlist, errs...)
			// If validation fails, don't trust it any further -
			// even Name could be bad.
		} else {
			name := kubecontainer.GetPodFullName(pod)
			if names.Has(name) {
				// TODO: when validation becomes versioned, this gets a bit
				// more complicated.
				errlist = append(errlist, field.Duplicate(field.NewPath("metadata", "name"), pod.Name))
			} else {
				names.Insert(name)
			}
		}
		if len(errlist) > 0 {
			name := bestPodIdentString(pod)
			err := errlist.ToAggregate()
			glog.Warningf("Pod[%d] (%s) from %s failed validation, ignoring: %v", i+1, name, source, err)
			recorder.Eventf(pod, api.EventTypeWarning, kubecontainer.FailedValidation, "Error validating pod %s from %s, ignoring: %v", name, source, err)
			continue
		}
		filtered = append(filtered, pod)
	}
	return
}

// Annotations that the kubelet adds to the pod.
var localAnnotations = []string{
	kubetypes.ConfigSourceAnnotationKey,
	kubetypes.ConfigMirrorAnnotationKey,
	kubetypes.ConfigFirstSeenAnnotationKey,
}

func isLocalAnnotationKey(key string) bool {
	for _, localKey := range localAnnotations {
		if key == localKey {
			return true
		}
	}
	return false
}

// isAnnotationMapEqual returns true if the existing annotation Map is equal to candidate except
// for local annotations.
func isAnnotationMapEqual(existingMap, candidateMap map[string]string) bool {
	if candidateMap == nil {
		candidateMap = make(map[string]string)
	}
	for k, v := range candidateMap {
		if isLocalAnnotationKey(k) {
			continue
		}
		if existingValue, ok := existingMap[k]; ok && existingValue == v {
			continue
		}
		return false
	}
	for k := range existingMap {
		if isLocalAnnotationKey(k) {
			continue
		}
		// stale entry in existing map.
		if _, exists := candidateMap[k]; !exists {
			return false
		}
	}
	return true
}

// recordFirstSeenTime records the first seen time of this pod.
func recordFirstSeenTime(pod *api.Pod) {
	glog.V(4).Infof("Receiving a new pod %q", format.Pod(pod))
	pod.Annotations[kubetypes.ConfigFirstSeenAnnotationKey] = kubetypes.NewTimestamp().GetString()
}

// updateAnnotations returns an Annotation map containing the api annotation map plus
// locally managed annotations
func updateAnnotations(existing, ref *api.Pod) {
	annotations := make(map[string]string, len(ref.Annotations)+len(localAnnotations))
	for k, v := range ref.Annotations {
		annotations[k] = v
	}
	for _, k := range localAnnotations {
		if v, ok := existing.Annotations[k]; ok {
			annotations[k] = v
		}
	}
	existing.Annotations = annotations
}

func podsDifferSemantically(existing, ref *api.Pod) bool {
	if reflect.DeepEqual(existing.Spec, ref.Spec) &&
		reflect.DeepEqual(existing.Labels, ref.Labels) &&
		reflect.DeepEqual(existing.DeletionTimestamp, ref.DeletionTimestamp) &&
		reflect.DeepEqual(existing.DeletionGracePeriodSeconds, ref.DeletionGracePeriodSeconds) &&
		isAnnotationMapEqual(existing.Annotations, ref.Annotations) {
		return false
	}
	return true
}

// checkAndUpdatePod updates existing, and:
//   * if ref makes a meaningful change, returns needUpdate=true
//   * if ref makes no meaningful change, but changes the pod status, returns needReconcile=true
//   * else return both false
//   Now, needUpdate and needReconcile should never be both true
func checkAndUpdatePod(existing, ref *api.Pod) (needUpdate, needReconcile bool) {
	// TODO: it would be better to update the whole object and only preserve certain things
	//       like the source annotation or the UID (to ensure safety)
	if !podsDifferSemantically(existing, ref) {
		// this is not an update
		// Only check reconcile when it is not an update, because if the pod is going to
		// be updated, an extra reconcile is unnecessary
		if !reflect.DeepEqual(existing.Status, ref.Status) {
			// Pod with changed pod status needs reconcile, because kubelet should
			// be the source of truth of pod status.
			existing.Status = ref.Status
			needReconcile = true
		}
		return
	}
	// this is an update

	// Overwrite the first-seen time with the existing one. This is our own
	// internal annotation, there is no need to update.
	ref.Annotations[kubetypes.ConfigFirstSeenAnnotationKey] = existing.Annotations[kubetypes.ConfigFirstSeenAnnotationKey]

	existing.Spec = ref.Spec
	existing.Labels = ref.Labels
	existing.DeletionTimestamp = ref.DeletionTimestamp
	existing.DeletionGracePeriodSeconds = ref.DeletionGracePeriodSeconds
	existing.Status = ref.Status
	updateAnnotations(existing, ref)
	needUpdate = true
	return
}

// Sync sends a copy of the current state through the update channel.
func (s *podStorage) Sync() {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()
	s.updates <- kubetypes.PodUpdate{Pods: s.MergedState().([]*api.Pod), Op: kubetypes.SET, Source: kubetypes.AllSource}
}

// Object implements config.Accessor
func (s *podStorage) MergedState() interface{} {
	s.podLock.RLock()
	defer s.podLock.RUnlock()
	pods := make([]*api.Pod, 0)
	for _, sourcePods := range s.pods {
		for _, podRef := range sourcePods {
			pod, err := api.Scheme.Copy(podRef)
			if err != nil {
				glog.Errorf("unable to copy pod: %v", err)
			}
			pods = append(pods, pod.(*api.Pod))
		}
	}
	return pods
}

func bestPodIdentString(pod *api.Pod) string {
	namespace := pod.Namespace
	if namespace == "" {
		namespace = "<empty-namespace>"
	}
	name := pod.Name
	if name == "" {
		name = "<empty-name>"
	}
	return fmt.Sprintf("%s.%s", name, namespace)
}

func copyPods(sourcePods []*api.Pod) []*api.Pod {
	pods := []*api.Pod{}
	for _, source := range sourcePods {
		// Use a deep copy here just in case
		pod, err := api.Scheme.Copy(source)
		if err != nil {
			glog.Errorf("unable to copy pod: %v", err)
		}
		pods = append(pods, pod.(*api.Pod))
	}
	return pods
}
