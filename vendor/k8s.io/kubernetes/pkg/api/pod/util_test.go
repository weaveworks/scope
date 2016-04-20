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
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/util/intstr"
)

func TestFindPort(t *testing.T) {
	testCases := []struct {
		name       string
		containers []api.Container
		port       intstr.IntOrString
		expected   int
		pass       bool
	}{{
		name:       "valid int, no ports",
		containers: []api.Container{{}},
		port:       intstr.FromInt(93),
		expected:   93,
		pass:       true,
	}, {
		name: "valid int, with ports",
		containers: []api.Container{{Ports: []api.ContainerPort{{
			Name:          "",
			ContainerPort: 11,
			Protocol:      "TCP",
		}, {
			Name:          "p",
			ContainerPort: 22,
			Protocol:      "TCP",
		}}}},
		port:     intstr.FromInt(93),
		expected: 93,
		pass:     true,
	}, {
		name:       "valid str, no ports",
		containers: []api.Container{{}},
		port:       intstr.FromString("p"),
		expected:   0,
		pass:       false,
	}, {
		name: "valid str, one ctr with ports",
		containers: []api.Container{{Ports: []api.ContainerPort{{
			Name:          "",
			ContainerPort: 11,
			Protocol:      "UDP",
		}, {
			Name:          "p",
			ContainerPort: 22,
			Protocol:      "TCP",
		}, {
			Name:          "q",
			ContainerPort: 33,
			Protocol:      "TCP",
		}}}},
		port:     intstr.FromString("q"),
		expected: 33,
		pass:     true,
	}, {
		name: "valid str, two ctr with ports",
		containers: []api.Container{{}, {Ports: []api.ContainerPort{{
			Name:          "",
			ContainerPort: 11,
			Protocol:      "UDP",
		}, {
			Name:          "p",
			ContainerPort: 22,
			Protocol:      "TCP",
		}, {
			Name:          "q",
			ContainerPort: 33,
			Protocol:      "TCP",
		}}}},
		port:     intstr.FromString("q"),
		expected: 33,
		pass:     true,
	}}

	for _, tc := range testCases {
		port, err := FindPort(&api.Pod{Spec: api.PodSpec{Containers: tc.containers}},
			&api.ServicePort{Protocol: "TCP", TargetPort: tc.port})
		if err != nil && tc.pass {
			t.Errorf("unexpected error for %s: %v", tc.name, err)
		}
		if err == nil && !tc.pass {
			t.Errorf("unexpected non-error for %s: %d", tc.name, port)
		}
		if port != tc.expected {
			t.Errorf("wrong result for %s: expected %d, got %d", tc.name, tc.expected, port)
		}
	}
}
