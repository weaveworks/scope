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

package kubectl

import (
	goerrors "errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/deployment"
	"k8s.io/kubernetes/pkg/util/integer"
	"k8s.io/kubernetes/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/util/wait"
)

const (
	sourceIdAnnotation         = kubectlAnnotationPrefix + "update-source-id"
	desiredReplicasAnnotation  = kubectlAnnotationPrefix + "desired-replicas"
	originalReplicasAnnotation = kubectlAnnotationPrefix + "original-replicas"
	nextControllerAnnotation   = kubectlAnnotationPrefix + "next-controller-id"
)

// RollingUpdaterConfig is the configuration for a rolling deployment process.
type RollingUpdaterConfig struct {
	// Out is a writer for progress output.
	Out io.Writer
	// OldRC is an existing controller to be replaced.
	OldRc *api.ReplicationController
	// NewRc is a controller that will take ownership of updated pods (will be
	// created if needed).
	NewRc *api.ReplicationController
	// UpdatePeriod is the time to wait between individual pod updates.
	UpdatePeriod time.Duration
	// Interval is the time to wait between polling controller status after
	// update.
	Interval time.Duration
	// Timeout is the time to wait for controller updates before giving up.
	Timeout time.Duration
	// CleanupPolicy defines the cleanup action to take after the deployment is
	// complete.
	CleanupPolicy RollingUpdaterCleanupPolicy
	// MaxUnavailable is the maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding up.
	// This can not be 0 if MaxSurge is 0.
	// By default, a fixed value of 1 is used.
	// Example: when this is set to 30%, the old RC can be scaled down to 70% of desired pods
	// immediately when the rolling update starts. Once new pods are ready, old RC
	// can be scaled down further, followed by scaling up the new RC, ensuring
	// that the total number of pods available at all times during the update is at
	// least 70% of desired pods.
	MaxUnavailable intstr.IntOrString
	// MaxSurge is the maximum number of pods that can be scheduled above the desired number of pods.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxUnavailable is 0.
	// Absolute number is calculated from percentage by rounding up.
	// By default, a value of 1 is used.
	// Example: when this is set to 30%, the new RC can be scaled up immediately
	// when the rolling update starts, such that the total number of old and new pods do not exceed
	// 130% of desired pods. Once old pods have been killed, new RC can be scaled up
	// further, ensuring that total number of pods running at any time during
	// the update is atmost 130% of desired pods.
	MaxSurge intstr.IntOrString
}

// RollingUpdaterCleanupPolicy is a cleanup action to take after the
// deployment is complete.
type RollingUpdaterCleanupPolicy string

const (
	// DeleteRollingUpdateCleanupPolicy means delete the old controller.
	DeleteRollingUpdateCleanupPolicy RollingUpdaterCleanupPolicy = "Delete"
	// PreserveRollingUpdateCleanupPolicy means keep the old controller.
	PreserveRollingUpdateCleanupPolicy RollingUpdaterCleanupPolicy = "Preserve"
	// RenameRollingUpdateCleanupPolicy means delete the old controller, and rename
	// the new controller to the name of the old controller.
	RenameRollingUpdateCleanupPolicy RollingUpdaterCleanupPolicy = "Rename"
)

// RollingUpdater provides methods for updating replicated pods in a predictable,
// fault-tolerant way.
type RollingUpdater struct {
	// Client interface for creating and updating controllers
	c client.Interface
	// Namespace for resources
	ns string
	// scaleAndWait scales a controller and returns its updated state.
	scaleAndWait func(rc *api.ReplicationController, retry *RetryParams, wait *RetryParams) (*api.ReplicationController, error)
	//getOrCreateTargetController gets and validates an existing controller or
	//makes a new one.
	getOrCreateTargetController func(controller *api.ReplicationController, sourceId string) (*api.ReplicationController, bool, error)
	// cleanup performs post deployment cleanup tasks for newRc and oldRc.
	cleanup func(oldRc, newRc *api.ReplicationController, config *RollingUpdaterConfig) error
	// getReadyPods returns the amount of old and new ready pods.
	getReadyPods func(oldRc, newRc *api.ReplicationController) (int, int, error)
}

// NewRollingUpdater creates a RollingUpdater from a client.
func NewRollingUpdater(namespace string, client client.Interface) *RollingUpdater {
	updater := &RollingUpdater{
		c:  client,
		ns: namespace,
	}
	// Inject real implementations.
	updater.scaleAndWait = updater.scaleAndWaitWithScaler
	updater.getOrCreateTargetController = updater.getOrCreateTargetControllerWithClient
	updater.getReadyPods = updater.readyPods
	updater.cleanup = updater.cleanupWithClients
	return updater
}

// Update all pods for a ReplicationController (oldRc) by creating a new
// controller (newRc) with 0 replicas, and synchronously scaling oldRc and
// newRc until oldRc has 0 replicas and newRc has the original # of desired
// replicas. Cleanup occurs based on a RollingUpdaterCleanupPolicy.
//
// Each interval, the updater will attempt to make progress however it can
// without violating any availability constraints defined by the config. This
// means the amount scaled up or down each interval will vary based on the
// timeliness of readiness and the updater will always try to make progress,
// even slowly.
//
// If an update from newRc to oldRc is already in progress, we attempt to
// drive it to completion. If an error occurs at any step of the update, the
// error will be returned.
//
// A scaling event (either up or down) is considered progress; if no progress
// is made within the config.Timeout, an error is returned.
//
// TODO: make this handle performing a rollback of a partially completed
// rollout.
func (r *RollingUpdater) Update(config *RollingUpdaterConfig) error {
	out := config.Out
	oldRc := config.OldRc
	scaleRetryParams := NewRetryParams(config.Interval, config.Timeout)

	// Find an existing controller (for continuing an interrupted update) or
	// create a new one if necessary.
	sourceId := fmt.Sprintf("%s:%s", oldRc.Name, oldRc.UID)
	newRc, existed, err := r.getOrCreateTargetController(config.NewRc, sourceId)
	if err != nil {
		return err
	}
	if existed {
		fmt.Fprintf(out, "Continuing update with existing controller %s.\n", newRc.Name)
	} else {
		fmt.Fprintf(out, "Created %s\n", newRc.Name)
	}
	// Extract the desired replica count from the controller.
	desired, err := strconv.Atoi(newRc.Annotations[desiredReplicasAnnotation])
	if err != nil {
		return fmt.Errorf("Unable to parse annotation for %s: %s=%s",
			newRc.Name, desiredReplicasAnnotation, newRc.Annotations[desiredReplicasAnnotation])
	}
	// Extract the original replica count from the old controller, adding the
	// annotation if it doesn't yet exist.
	_, hasOriginalAnnotation := oldRc.Annotations[originalReplicasAnnotation]
	if !hasOriginalAnnotation {
		existing, err := r.c.ReplicationControllers(oldRc.Namespace).Get(oldRc.Name)
		if err != nil {
			return err
		}
		if existing.Annotations == nil {
			existing.Annotations = map[string]string{}
		}
		existing.Annotations[originalReplicasAnnotation] = strconv.Itoa(existing.Spec.Replicas)
		updated, err := r.c.ReplicationControllers(existing.Namespace).Update(existing)
		if err != nil {
			return err
		}
		oldRc = updated
	}
	// maxSurge is the maximum scaling increment and maxUnavailable are the maximum pods
	// that can be unavailable during a rollout.
	maxSurge, maxUnavailable, err := deployment.ResolveFenceposts(&config.MaxSurge, &config.MaxUnavailable, desired)
	if err != nil {
		return err
	}
	// Validate maximums.
	if desired > 0 && maxUnavailable == 0 && maxSurge == 0 {
		return fmt.Errorf("one of maxSurge or maxUnavailable must be specified")
	}
	// The minumum pods which must remain available througout the update
	// calculated for internal convenience.
	minAvailable := integer.IntMax(0, desired-maxUnavailable)
	// If the desired new scale is 0, then the max unavailable is necessarily
	// the effective scale of the old RC regardless of the configuration
	// (equivalent to 100% maxUnavailable).
	if desired == 0 {
		maxUnavailable = oldRc.Spec.Replicas
		minAvailable = 0
	}

	fmt.Fprintf(out, "Scaling up %s from %d to %d, scaling down %s from %d to 0 (keep %d pods available, don't exceed %d pods)\n",
		newRc.Name, newRc.Spec.Replicas, desired, oldRc.Name, oldRc.Spec.Replicas, minAvailable, desired+maxSurge)

	// Scale newRc and oldRc until newRc has the desired number of replicas and
	// oldRc has 0 replicas.
	progressDeadline := time.Now().UnixNano() + config.Timeout.Nanoseconds()
	for newRc.Spec.Replicas != desired || oldRc.Spec.Replicas != 0 {
		// Store the existing replica counts for progress timeout tracking.
		newReplicas := newRc.Spec.Replicas
		oldReplicas := oldRc.Spec.Replicas

		// Scale up as much as possible.
		scaledRc, err := r.scaleUp(newRc, oldRc, desired, maxSurge, maxUnavailable, scaleRetryParams, config)
		if err != nil {
			return err
		}
		newRc = scaledRc

		// Wait between scaling operations for things to settle.
		time.Sleep(config.UpdatePeriod)

		// Scale down as much as possible.
		scaledRc, err = r.scaleDown(newRc, oldRc, desired, minAvailable, maxUnavailable, maxSurge, config)
		if err != nil {
			return err
		}
		oldRc = scaledRc

		// If we are making progress, continue to advance the progress deadline.
		// Otherwise, time out with an error.
		progressMade := (newRc.Spec.Replicas != newReplicas) || (oldRc.Spec.Replicas != oldReplicas)
		if progressMade {
			progressDeadline = time.Now().UnixNano() + config.Timeout.Nanoseconds()
		} else if time.Now().UnixNano() > progressDeadline {
			return fmt.Errorf("timed out waiting for any update progress to be made")
		}
	}

	// Housekeeping and cleanup policy execution.
	return r.cleanup(oldRc, newRc, config)
}

// scaleUp scales up newRc to desired by whatever increment is possible given
// the configured surge threshold. scaleUp will safely no-op as necessary when
// it detects redundancy or other relevant conditions.
func (r *RollingUpdater) scaleUp(newRc, oldRc *api.ReplicationController, desired, maxSurge, maxUnavailable int, scaleRetryParams *RetryParams, config *RollingUpdaterConfig) (*api.ReplicationController, error) {
	// If we're already at the desired, do nothing.
	if newRc.Spec.Replicas == desired {
		return newRc, nil
	}

	// Scale up as far as we can based on the surge limit.
	increment := (desired + maxSurge) - (oldRc.Spec.Replicas + newRc.Spec.Replicas)
	// If the old is already scaled down, go ahead and scale all the way up.
	if oldRc.Spec.Replicas == 0 {
		increment = desired - newRc.Spec.Replicas
	}
	// We can't scale up without violating the surge limit, so do nothing.
	if increment <= 0 {
		return newRc, nil
	}
	// Increase the replica count, and deal with fenceposts.
	newRc.Spec.Replicas += increment
	if newRc.Spec.Replicas > desired {
		newRc.Spec.Replicas = desired
	}
	// Perform the scale-up.
	fmt.Fprintf(config.Out, "Scaling %s up to %d\n", newRc.Name, newRc.Spec.Replicas)
	scaledRc, err := r.scaleAndWait(newRc, scaleRetryParams, scaleRetryParams)
	if err != nil {
		return nil, err
	}
	return scaledRc, nil
}

// scaleDown scales down oldRc to 0 at whatever decrement possible given the
// thresholds defined on the config. scaleDown will safely no-op as necessary
// when it detects redundancy or other relevant conditions.
func (r *RollingUpdater) scaleDown(newRc, oldRc *api.ReplicationController, desired, minAvailable, maxUnavailable, maxSurge int, config *RollingUpdaterConfig) (*api.ReplicationController, error) {
	// Already scaled down; do nothing.
	if oldRc.Spec.Replicas == 0 {
		return oldRc, nil
	}
	// Get ready pods. We shouldn't block, otherwise in case both old and new
	// pods are unavailable then the rolling update process blocks.
	// Timeout-wise we are already covered by the progress check.
	_, newAvailable, err := r.getReadyPods(oldRc, newRc)
	if err != nil {
		return nil, err
	}
	// The old controller is considered as part of the total because we want to
	// maintain minimum availability even with a volatile old controller.
	// Scale down as much as possible while maintaining minimum availability
	allPods := oldRc.Spec.Replicas + newRc.Spec.Replicas
	newUnavailable := newRc.Spec.Replicas - newAvailable
	decrement := allPods - minAvailable - newUnavailable
	// The decrement normally shouldn't drop below 0 because the available count
	// always starts below the old replica count, but the old replica count can
	// decrement due to externalities like pods death in the replica set. This
	// will be considered a transient condition; do nothing and try again later
	// with new readiness values.
	//
	// If the most we can scale is 0, it means we can't scale down without
	// violating the minimum. Do nothing and try again later when conditions may
	// have changed.
	if decrement <= 0 {
		return oldRc, nil
	}
	// Reduce the replica count, and deal with fenceposts.
	oldRc.Spec.Replicas -= decrement
	if oldRc.Spec.Replicas < 0 {
		oldRc.Spec.Replicas = 0
	}
	// If the new is already fully scaled and available up to the desired size, go
	// ahead and scale old all the way down.
	if newRc.Spec.Replicas == desired && newAvailable == desired {
		oldRc.Spec.Replicas = 0
	}
	// Perform the scale-down.
	fmt.Fprintf(config.Out, "Scaling %s down to %d\n", oldRc.Name, oldRc.Spec.Replicas)
	retryWait := &RetryParams{config.Interval, config.Timeout}
	scaledRc, err := r.scaleAndWait(oldRc, retryWait, retryWait)
	if err != nil {
		return nil, err
	}
	return scaledRc, nil
}

// scalerScaleAndWait scales a controller using a Scaler and a real client.
func (r *RollingUpdater) scaleAndWaitWithScaler(rc *api.ReplicationController, retry *RetryParams, wait *RetryParams) (*api.ReplicationController, error) {
	scaler, err := ScalerFor(api.Kind("ReplicationController"), r.c)
	if err != nil {
		return nil, fmt.Errorf("Couldn't make scaler: %s", err)
	}
	if err := scaler.Scale(rc.Namespace, rc.Name, uint(rc.Spec.Replicas), &ScalePrecondition{-1, ""}, retry, wait); err != nil {
		return nil, err
	}
	return r.c.ReplicationControllers(rc.Namespace).Get(rc.Name)
}

// readyPods returns the old and new ready counts for their pods.
// If a pod is observed as being ready, it's considered ready even
// if it later becomes notReady.
func (r *RollingUpdater) readyPods(oldRc, newRc *api.ReplicationController) (int, int, error) {
	controllers := []*api.ReplicationController{oldRc, newRc}
	oldReady := 0
	newReady := 0

	for i := range controllers {
		controller := controllers[i]
		selector := labels.Set(controller.Spec.Selector).AsSelector()
		options := api.ListOptions{LabelSelector: selector}
		pods, err := r.c.Pods(controller.Namespace).List(options)
		if err != nil {
			return 0, 0, err
		}
		for _, pod := range pods.Items {
			if api.IsPodReady(&pod) {
				switch controller.Name {
				case oldRc.Name:
					oldReady++
				case newRc.Name:
					newReady++
				}
			}
		}
	}
	return oldReady, newReady, nil
}

// getOrCreateTargetControllerWithClient looks for an existing controller with
// sourceId. If found, the existing controller is returned with true
// indicating that the controller already exists. If the controller isn't
// found, a new one is created and returned along with false indicating the
// controller was created.
//
// Existing controllers are validated to ensure their sourceIdAnnotation
// matches sourceId; if there's a mismatch, an error is returned.
func (r *RollingUpdater) getOrCreateTargetControllerWithClient(controller *api.ReplicationController, sourceId string) (*api.ReplicationController, bool, error) {
	existingRc, err := r.existingController(controller)
	if err != nil {
		if !errors.IsNotFound(err) {
			// There was an error trying to find the controller; don't assume we
			// should create it.
			return nil, false, err
		}
		if controller.Spec.Replicas <= 0 {
			return nil, false, fmt.Errorf("Invalid controller spec for %s; required: > 0 replicas, actual: %d\n", controller.Name, controller.Spec.Replicas)
		}
		// The controller wasn't found, so create it.
		if controller.Annotations == nil {
			controller.Annotations = map[string]string{}
		}
		controller.Annotations[desiredReplicasAnnotation] = fmt.Sprintf("%d", controller.Spec.Replicas)
		controller.Annotations[sourceIdAnnotation] = sourceId
		controller.Spec.Replicas = 0
		newRc, err := r.c.ReplicationControllers(r.ns).Create(controller)
		return newRc, false, err
	}
	// Validate and use the existing controller.
	annotations := existingRc.Annotations
	source := annotations[sourceIdAnnotation]
	_, ok := annotations[desiredReplicasAnnotation]
	if source != sourceId || !ok {
		return nil, false, fmt.Errorf("Missing/unexpected annotations for controller %s, expected %s : %s", controller.Name, sourceId, annotations)
	}
	return existingRc, true, nil
}

// existingController verifies if the controller already exists
func (r *RollingUpdater) existingController(controller *api.ReplicationController) (*api.ReplicationController, error) {
	// without rc name but generate name, there's no existing rc
	if len(controller.Name) == 0 && len(controller.GenerateName) > 0 {
		return nil, errors.NewNotFound(api.Resource("replicationcontrollers"), controller.Name)
	}
	// controller name is required to get rc back
	return r.c.ReplicationControllers(controller.Namespace).Get(controller.Name)
}

// cleanupWithClients performs cleanup tasks after the rolling update. Update
// process related annotations are removed from oldRc and newRc. The
// CleanupPolicy on config is executed.
func (r *RollingUpdater) cleanupWithClients(oldRc, newRc *api.ReplicationController, config *RollingUpdaterConfig) error {
	// Clean up annotations
	var err error
	newRc, err = r.c.ReplicationControllers(r.ns).Get(newRc.Name)
	if err != nil {
		return err
	}
	delete(newRc.Annotations, sourceIdAnnotation)
	delete(newRc.Annotations, desiredReplicasAnnotation)

	newRc, err = r.c.ReplicationControllers(r.ns).Update(newRc)
	if err != nil {
		return err
	}
	if err = wait.Poll(config.Interval, config.Timeout, client.ControllerHasDesiredReplicas(r.c, newRc)); err != nil {
		return err
	}
	newRc, err = r.c.ReplicationControllers(r.ns).Get(newRc.Name)
	if err != nil {
		return err
	}

	switch config.CleanupPolicy {
	case DeleteRollingUpdateCleanupPolicy:
		// delete old rc
		fmt.Fprintf(config.Out, "Update succeeded. Deleting %s\n", oldRc.Name)
		return r.c.ReplicationControllers(r.ns).Delete(oldRc.Name)
	case RenameRollingUpdateCleanupPolicy:
		// delete old rc
		fmt.Fprintf(config.Out, "Update succeeded. Deleting old controller: %s\n", oldRc.Name)
		if err := r.c.ReplicationControllers(r.ns).Delete(oldRc.Name); err != nil {
			return err
		}
		fmt.Fprintf(config.Out, "Renaming %s to %s\n", newRc.Name, oldRc.Name)
		return Rename(r.c, newRc, oldRc.Name)
	case PreserveRollingUpdateCleanupPolicy:
		return nil
	default:
		return nil
	}
}

func Rename(c client.ReplicationControllersNamespacer, rc *api.ReplicationController, newName string) error {
	oldName := rc.Name
	rc.Name = newName
	rc.ResourceVersion = ""

	_, err := c.ReplicationControllers(rc.Namespace).Create(rc)
	if err != nil {
		return err
	}
	err = c.ReplicationControllers(rc.Namespace).Delete(oldName)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func LoadExistingNextReplicationController(c client.ReplicationControllersNamespacer, namespace, newName string) (*api.ReplicationController, error) {
	if len(newName) == 0 {
		return nil, nil
	}
	newRc, err := c.ReplicationControllers(namespace).Get(newName)
	if err != nil && errors.IsNotFound(err) {
		return nil, nil
	}
	return newRc, err
}

func CreateNewControllerFromCurrentController(c client.Interface, codec runtime.Codec, namespace, oldName, newName, image, container, deploymentKey string) (*api.ReplicationController, error) {
	containerIndex := 0
	// load the old RC into the "new" RC
	newRc, err := c.ReplicationControllers(namespace).Get(oldName)
	if err != nil {
		return nil, err
	}

	if len(container) != 0 {
		containerFound := false

		for i, c := range newRc.Spec.Template.Spec.Containers {
			if c.Name == container {
				containerIndex = i
				containerFound = true
				break
			}
		}

		if !containerFound {
			return nil, fmt.Errorf("container %s not found in pod", container)
		}
	}

	if len(newRc.Spec.Template.Spec.Containers) > 1 && len(container) == 0 {
		return nil, goerrors.New("Must specify container to update when updating a multi-container pod")
	}

	if len(newRc.Spec.Template.Spec.Containers) == 0 {
		return nil, goerrors.New(fmt.Sprintf("Pod has no containers! (%v)", newRc))
	}
	newRc.Spec.Template.Spec.Containers[containerIndex].Image = image

	newHash, err := api.HashObject(newRc, codec)
	if err != nil {
		return nil, err
	}

	if len(newName) == 0 {
		newName = fmt.Sprintf("%s-%s", newRc.Name, newHash)
	}
	newRc.Name = newName

	newRc.Spec.Selector[deploymentKey] = newHash
	newRc.Spec.Template.Labels[deploymentKey] = newHash
	// Clear resource version after hashing so that identical updates get different hashes.
	newRc.ResourceVersion = ""
	return newRc, nil
}

func AbortRollingUpdate(c *RollingUpdaterConfig) error {
	// Swap the controllers
	tmp := c.OldRc
	c.OldRc = c.NewRc
	c.NewRc = tmp

	if c.NewRc.Annotations == nil {
		c.NewRc.Annotations = map[string]string{}
	}
	c.NewRc.Annotations[sourceIdAnnotation] = fmt.Sprintf("%s:%s", c.OldRc.Name, c.OldRc.UID)

	// Use the original value since the replica count change from old to new
	// could be asymmetric. If we don't know the original count, we can't safely
	// roll back to a known good size.
	originalSize, foundOriginal := tmp.Annotations[originalReplicasAnnotation]
	if !foundOriginal {
		return fmt.Errorf("couldn't find original replica count of %q", tmp.Name)
	}
	fmt.Fprintf(c.Out, "Setting %q replicas to %s\n", c.NewRc.Name, originalSize)
	c.NewRc.Annotations[desiredReplicasAnnotation] = originalSize
	c.CleanupPolicy = DeleteRollingUpdateCleanupPolicy
	return nil
}

func GetNextControllerAnnotation(rc *api.ReplicationController) (string, bool) {
	res, found := rc.Annotations[nextControllerAnnotation]
	return res, found
}

func SetNextControllerAnnotation(rc *api.ReplicationController, name string) {
	if rc.Annotations == nil {
		rc.Annotations = map[string]string{}
	}
	rc.Annotations[nextControllerAnnotation] = name
}

func UpdateExistingReplicationController(c client.Interface, oldRc *api.ReplicationController, namespace, newName, deploymentKey, deploymentValue string, out io.Writer) (*api.ReplicationController, error) {
	SetNextControllerAnnotation(oldRc, newName)
	if _, found := oldRc.Spec.Selector[deploymentKey]; !found {
		return AddDeploymentKeyToReplicationController(oldRc, c, deploymentKey, deploymentValue, namespace, out)
	} else {
		// If we didn't need to update the controller for the deployment key, we still need to write
		// the "next" controller.
		return c.ReplicationControllers(namespace).Update(oldRc)
	}
}

const MaxRetries = 3

func AddDeploymentKeyToReplicationController(oldRc *api.ReplicationController, client client.Interface, deploymentKey, deploymentValue, namespace string, out io.Writer) (*api.ReplicationController, error) {
	var err error
	// First, update the template label.  This ensures that any newly created pods will have the new label
	if oldRc, err = updateWithRetries(client.ReplicationControllers(namespace), oldRc, func(rc *api.ReplicationController) {
		if rc.Spec.Template.Labels == nil {
			rc.Spec.Template.Labels = map[string]string{}
		}
		rc.Spec.Template.Labels[deploymentKey] = deploymentValue
	}); err != nil {
		return nil, err
	}

	// Update all pods managed by the rc to have the new hash label, so they are correctly adopted
	// TODO: extract the code from the label command and re-use it here.
	selector := labels.SelectorFromSet(oldRc.Spec.Selector)
	options := api.ListOptions{LabelSelector: selector}
	podList, err := client.Pods(namespace).List(options)
	if err != nil {
		return nil, err
	}
	for ix := range podList.Items {
		pod := &podList.Items[ix]
		if pod.Labels == nil {
			pod.Labels = map[string]string{
				deploymentKey: deploymentValue,
			}
		} else {
			pod.Labels[deploymentKey] = deploymentValue
		}
		err = nil
		delay := 3
		for i := 0; i < MaxRetries; i++ {
			_, err = client.Pods(namespace).Update(pod)
			if err != nil {
				fmt.Fprintf(out, "Error updating pod (%v), retrying after %d seconds", err, delay)
				time.Sleep(time.Second * time.Duration(delay))
				delay *= delay
			} else {
				break
			}
		}
		if err != nil {
			return nil, err
		}
	}

	if oldRc.Spec.Selector == nil {
		oldRc.Spec.Selector = map[string]string{}
	}
	// Copy the old selector, so that we can scrub out any orphaned pods
	selectorCopy := map[string]string{}
	for k, v := range oldRc.Spec.Selector {
		selectorCopy[k] = v
	}
	oldRc.Spec.Selector[deploymentKey] = deploymentValue

	// Update the selector of the rc so it manages all the pods we updated above
	if oldRc, err = updateWithRetries(client.ReplicationControllers(namespace), oldRc, func(rc *api.ReplicationController) {
		rc.Spec.Selector[deploymentKey] = deploymentValue
	}); err != nil {
		return nil, err
	}

	// Clean up any orphaned pods that don't have the new label, this can happen if the rc manager
	// doesn't see the update to its pod template and creates a new pod with the old labels after
	// we've finished re-adopting existing pods to the rc.
	selector = labels.SelectorFromSet(selectorCopy)
	options = api.ListOptions{LabelSelector: selector}
	podList, err = client.Pods(namespace).List(options)
	for ix := range podList.Items {
		pod := &podList.Items[ix]
		if value, found := pod.Labels[deploymentKey]; !found || value != deploymentValue {
			if err := client.Pods(namespace).Delete(pod.Name, nil); err != nil {
				return nil, err
			}
		}
	}

	return oldRc, nil
}

type updateFunc func(controller *api.ReplicationController)

// updateWithRetries updates applies the given rc as an update.
func updateWithRetries(rcClient client.ReplicationControllerInterface, rc *api.ReplicationController, applyUpdate updateFunc) (*api.ReplicationController, error) {
	var err error
	oldRc := rc
	err = wait.Poll(10*time.Millisecond, 1*time.Minute, func() (bool, error) {
		// Apply the update, then attempt to push it to the apiserver.
		applyUpdate(rc)
		if rc, err = rcClient.Update(rc); err == nil {
			// rc contains the latest controller post update
			return true, nil
		}
		// Update the controller with the latest resource version, if the update failed we
		// can't trust rc so use oldRc.Name.
		if rc, err = rcClient.Get(oldRc.Name); err != nil {
			// The Get failed: Value in rc cannot be trusted.
			rc = oldRc
		}
		// The Get passed: rc contains the latest controller, expect a poll for the update.
		return false, nil
	})
	// If the error is non-nil the returned controller cannot be trusted, if it is nil, the returned
	// controller contains the applied update.
	return rc, err
}

func FindSourceController(r client.ReplicationControllersNamespacer, namespace, name string) (*api.ReplicationController, error) {
	list, err := r.ReplicationControllers(namespace).List(api.ListOptions{})
	if err != nil {
		return nil, err
	}
	for ix := range list.Items {
		rc := &list.Items[ix]
		if rc.Annotations != nil && strings.HasPrefix(rc.Annotations[sourceIdAnnotation], name) {
			return rc, nil
		}
	}
	return nil, fmt.Errorf("couldn't find a replication controller with source id == %s/%s", namespace, name)
}
