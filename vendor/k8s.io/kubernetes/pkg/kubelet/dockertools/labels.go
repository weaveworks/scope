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

package dockertools

import (
	"encoding/json"
	"strconv"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	"k8s.io/kubernetes/pkg/kubelet/custommetrics"
	"k8s.io/kubernetes/pkg/kubelet/util/format"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/types"
)

// This file contains all docker label related constants and functions, including:
//  * label setters and getters
//  * label filters (maybe in the future)

const (
	kubernetesPodNameLabel                   = "io.kubernetes.pod.name"
	kubernetesPodNamespaceLabel              = "io.kubernetes.pod.namespace"
	kubernetesPodUIDLabel                    = "io.kubernetes.pod.uid"
	kubernetesPodDeletionGracePeriodLabel    = "io.kubernetes.pod.deletionGracePeriod"
	kubernetesPodTerminationGracePeriodLabel = "io.kubernetes.pod.terminationGracePeriod"

	kubernetesContainerNameLabel                   = "io.kubernetes.container.name"
	kubernetesContainerHashLabel                   = "io.kubernetes.container.hash"
	kubernetesContainerRestartCountLabel           = "io.kubernetes.container.restartCount"
	kubernetesContainerTerminationMessagePathLabel = "io.kubernetes.container.terminationMessagePath"
	kubernetesContainerPreStopHandlerLabel         = "io.kubernetes.container.preStopHandler"

	// TODO(random-liu): Keep this for old containers, remove this when we drop support for v1.1.
	kubernetesPodLabel = "io.kubernetes.pod.data"

	cadvisorPrometheusMetricsLabel = "io.cadvisor.metric.prometheus"
)

// Container information which has been labelled on each docker container
// TODO(random-liu): The type of Hash should be compliance with kubelet container status.
type labelledContainerInfo struct {
	PodName                   string
	PodNamespace              string
	PodUID                    types.UID
	PodDeletionGracePeriod    *int64
	PodTerminationGracePeriod *int64
	Name                      string
	Hash                      string
	RestartCount              int
	TerminationMessagePath    string
	PreStopHandler            *api.Handler
}

func GetContainerName(labels map[string]string) string {
	return labels[kubernetesContainerNameLabel]
}

func GetPodName(labels map[string]string) string {
	return labels[kubernetesPodNameLabel]
}

func GetPodUID(labels map[string]string) string {
	return labels[kubernetesPodUIDLabel]
}

func GetPodNamespace(labels map[string]string) string {
	return labels[kubernetesPodNamespaceLabel]
}

func newLabels(container *api.Container, pod *api.Pod, restartCount int, enableCustomMetrics bool) map[string]string {
	labels := map[string]string{}
	labels[kubernetesPodNameLabel] = pod.Name
	labels[kubernetesPodNamespaceLabel] = pod.Namespace
	labels[kubernetesPodUIDLabel] = string(pod.UID)
	if pod.DeletionGracePeriodSeconds != nil {
		labels[kubernetesPodDeletionGracePeriodLabel] = strconv.FormatInt(*pod.DeletionGracePeriodSeconds, 10)
	}
	if pod.Spec.TerminationGracePeriodSeconds != nil {
		labels[kubernetesPodTerminationGracePeriodLabel] = strconv.FormatInt(*pod.Spec.TerminationGracePeriodSeconds, 10)
	}

	labels[kubernetesContainerNameLabel] = container.Name
	labels[kubernetesContainerHashLabel] = strconv.FormatUint(kubecontainer.HashContainer(container), 16)
	labels[kubernetesContainerRestartCountLabel] = strconv.Itoa(restartCount)
	labels[kubernetesContainerTerminationMessagePathLabel] = container.TerminationMessagePath
	if container.Lifecycle != nil && container.Lifecycle.PreStop != nil {
		// Using json enconding so that the PreStop handler object is readable after writing as a label
		rawPreStop, err := json.Marshal(container.Lifecycle.PreStop)
		if err != nil {
			glog.Errorf("Unable to marshal lifecycle PreStop handler for container %q of pod %q: %v", container.Name, format.Pod(pod), err)
		} else {
			labels[kubernetesContainerPreStopHandlerLabel] = string(rawPreStop)
		}
	}

	if enableCustomMetrics {
		path, err := custommetrics.GetCAdvisorCustomMetricsDefinitionPath(container)
		if path != nil && err == nil {
			labels[cadvisorPrometheusMetricsLabel] = *path
		}
	}

	return labels
}

func getContainerInfoFromLabel(labels map[string]string) *labelledContainerInfo {
	var err error
	containerInfo := &labelledContainerInfo{
		PodName:      getStringValueFromLabel(labels, kubernetesPodNameLabel),
		PodNamespace: getStringValueFromLabel(labels, kubernetesPodNamespaceLabel),
		PodUID:       types.UID(getStringValueFromLabel(labels, kubernetesPodUIDLabel)),
		Name:         getStringValueFromLabel(labels, kubernetesContainerNameLabel),
		Hash:         getStringValueFromLabel(labels, kubernetesContainerHashLabel),
		TerminationMessagePath: getStringValueFromLabel(labels, kubernetesContainerTerminationMessagePathLabel),
	}
	if containerInfo.RestartCount, err = getIntValueFromLabel(labels, kubernetesContainerRestartCountLabel); err != nil {
		logError(containerInfo, kubernetesContainerRestartCountLabel, err)
	}
	if containerInfo.PodDeletionGracePeriod, err = getInt64PointerFromLabel(labels, kubernetesPodDeletionGracePeriodLabel); err != nil {
		logError(containerInfo, kubernetesPodDeletionGracePeriodLabel, err)
	}
	if containerInfo.PodTerminationGracePeriod, err = getInt64PointerFromLabel(labels, kubernetesPodTerminationGracePeriodLabel); err != nil {
		logError(containerInfo, kubernetesPodTerminationGracePeriodLabel, err)
	}
	preStopHandler := &api.Handler{}
	if found, err := getJsonObjectFromLabel(labels, kubernetesContainerPreStopHandlerLabel, preStopHandler); err != nil {
		logError(containerInfo, kubernetesContainerPreStopHandlerLabel, err)
	} else if found {
		containerInfo.PreStopHandler = preStopHandler
	}
	supplyContainerInfoWithOldLabel(labels, containerInfo)
	return containerInfo
}

func getStringValueFromLabel(labels map[string]string, label string) string {
	if value, found := labels[label]; found {
		return value
	}
	// Do not report error, because there should be many old containers without label now.
	glog.V(3).Infof("Container doesn't have label %s, it may be an old or invalid container", label)
	// Return empty string "" for these containers, the caller will get value by other ways.
	return ""
}

func getIntValueFromLabel(labels map[string]string, label string) (int, error) {
	if strValue, found := labels[label]; found {
		intValue, err := strconv.Atoi(strValue)
		if err != nil {
			// This really should not happen. Just set value to 0 to handle this abnormal case
			return 0, err
		}
		return intValue, nil
	}
	// Do not report error, because there should be many old containers without label now.
	glog.V(3).Infof("Container doesn't have label %s, it may be an old or invalid container", label)
	// Just set the value to 0
	return 0, nil
}

func getInt64PointerFromLabel(labels map[string]string, label string) (*int64, error) {
	if strValue, found := labels[label]; found {
		int64Value, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return nil, err
		}
		return &int64Value, nil
	}
	// Because it's normal that a container has no PodDeletionGracePeriod and PodTerminationGracePeriod label,
	// don't report any error here.
	return nil, nil
}

// getJsonObjectFromLabel returns a bool value indicating whether an object is found
func getJsonObjectFromLabel(labels map[string]string, label string, value interface{}) (bool, error) {
	if strValue, found := labels[label]; found {
		err := json.Unmarshal([]byte(strValue), value)
		return found, err
	}
	// Because it's normal that a container has no PreStopHandler label, don't report any error here.
	return false, nil
}

// The label kubernetesPodLabel is added a long time ago (#7421), it serialized the whole api.Pod to a docker label.
// We want to remove this label because it serialized too much useless information. However kubelet may still work
// with old containers which only have this label for a long time until we completely deprecate the old label.
// Before that to ensure correctness we have to supply information with the old labels when newly added labels
// are not available.
// TODO(random-liu): Remove this function when we can completely remove label kubernetesPodLabel, probably after
// dropping support for v1.1.
func supplyContainerInfoWithOldLabel(labels map[string]string, containerInfo *labelledContainerInfo) {
	// Get api.Pod from old label
	var pod *api.Pod
	data, found := labels[kubernetesPodLabel]
	if !found {
		// Don't report any error here, because it's normal that a container has no pod label, especially
		// when we gradually deprecate the old label
		return
	}
	pod = &api.Pod{}
	if err := runtime.DecodeInto(api.Codecs.UniversalDecoder(), []byte(data), pod); err != nil {
		// If the pod label can't be parsed, we should report an error
		logError(containerInfo, kubernetesPodLabel, err)
		return
	}
	if containerInfo.PodDeletionGracePeriod == nil {
		containerInfo.PodDeletionGracePeriod = pod.DeletionGracePeriodSeconds
	}
	if containerInfo.PodTerminationGracePeriod == nil {
		containerInfo.PodTerminationGracePeriod = pod.Spec.TerminationGracePeriodSeconds
	}

	// Get api.Container from api.Pod
	var container *api.Container
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name == containerInfo.Name {
			container = &pod.Spec.Containers[i]
			break
		}
	}
	if container == nil {
		glog.Errorf("Unable to find container %q in pod %q", containerInfo.Name, format.Pod(pod))
		return
	}
	if containerInfo.PreStopHandler == nil && container.Lifecycle != nil {
		containerInfo.PreStopHandler = container.Lifecycle.PreStop
	}
}

func logError(containerInfo *labelledContainerInfo, label string, err error) {
	glog.Errorf("Unable to get %q for container %q of pod %q: %v", label, containerInfo.Name,
		kubecontainer.BuildPodFullName(containerInfo.PodName, containerInfo.PodNamespace), err)
}
