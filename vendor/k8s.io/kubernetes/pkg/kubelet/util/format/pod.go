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

package format

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/api"
)

type podHandler func(*api.Pod) string

// Pod returns a string reprenetating a pod in a human readable format,
// with pod UID as part of the string.
func Pod(pod *api.Pod) string {
	// Use underscore as the delimiter because it is not allowed in pod name
	// (DNS subdomain format), while allowed in the container name format.
	return fmt.Sprintf("%s_%s(%s)", pod.Name, pod.Namespace, pod.UID)
}

// Pods returns a string representating a list of pods in a human
// readable format.
func Pods(pods []*api.Pod) string {
	return aggregatePods(pods, Pod)
}

func aggregatePods(pods []*api.Pod, handler podHandler) string {
	podStrings := make([]string, 0, len(pods))
	for _, pod := range pods {
		podStrings = append(podStrings, handler(pod))
	}
	return fmt.Sprintf(strings.Join(podStrings, ", "))
}
