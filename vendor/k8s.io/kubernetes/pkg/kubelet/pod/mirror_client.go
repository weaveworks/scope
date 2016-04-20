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

package pod

import (
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	kubetypes "k8s.io/kubernetes/pkg/kubelet/types"
)

// Mirror client is used to create/delete a mirror pod.
type MirrorClient interface {
	CreateMirrorPod(*api.Pod) error
	DeleteMirrorPod(string) error
}

type basicMirrorClient struct {
	// mirror pods are stored in the kubelet directly because they need to be
	// in sync with the internal pods.
	apiserverClient clientset.Interface
}

func NewBasicMirrorClient(apiserverClient clientset.Interface) MirrorClient {
	return &basicMirrorClient{apiserverClient: apiserverClient}
}

// Creates a mirror pod.
func (mc *basicMirrorClient) CreateMirrorPod(pod *api.Pod) error {
	if mc.apiserverClient == nil {
		return nil
	}
	// Make a copy of the pod.
	copyPod := *pod
	copyPod.Annotations = make(map[string]string)

	for k, v := range pod.Annotations {
		copyPod.Annotations[k] = v
	}
	hash := getPodHash(pod)
	copyPod.Annotations[kubetypes.ConfigMirrorAnnotationKey] = hash
	apiPod, err := mc.apiserverClient.Core().Pods(copyPod.Namespace).Create(&copyPod)
	if err != nil && errors.IsAlreadyExists(err) {
		// Check if the existing pod is the same as the pod we want to create.
		if h, ok := apiPod.Annotations[kubetypes.ConfigMirrorAnnotationKey]; ok && h == hash {
			return nil
		}
	}
	return err
}

// Deletes a mirror pod.
func (mc *basicMirrorClient) DeleteMirrorPod(podFullName string) error {
	if mc.apiserverClient == nil {
		return nil
	}
	name, namespace, err := kubecontainer.ParsePodFullName(podFullName)
	if err != nil {
		glog.Errorf("Failed to parse a pod full name %q", podFullName)
		return err
	}
	glog.V(4).Infof("Deleting a mirror pod %q", podFullName)
	if err := mc.apiserverClient.Core().Pods(namespace).Delete(name, api.NewDeleteOptions(0)); err != nil && !errors.IsNotFound(err) {
		glog.Errorf("Failed deleting a mirror pod %q: %v", podFullName, err)
	}
	return nil
}

func IsStaticPod(pod *api.Pod) bool {
	source, err := kubetypes.GetPodSource(pod)
	return err == nil && source != kubetypes.ApiserverSource
}

func IsMirrorPod(pod *api.Pod) bool {
	_, ok := pod.Annotations[kubetypes.ConfigMirrorAnnotationKey]
	return ok
}

func getHashFromMirrorPod(pod *api.Pod) (string, bool) {
	hash, ok := pod.Annotations[kubetypes.ConfigMirrorAnnotationKey]
	return hash, ok
}

func getPodHash(pod *api.Pod) string {
	// The annotation exists for all static pods.
	return pod.Annotations[kubetypes.ConfigHashAnnotationKey]
}
