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
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/util/intstr"
)

func TestGenerateService(t *testing.T) {
	tests := []struct {
		generator Generator
		params    map[string]interface{}
		expected  api.Service
	}{
		{
			generator: ServiceGeneratorV2{},
			params: map[string]interface{}{
				"selector":       "foo=bar,baz=blah",
				"name":           "test",
				"port":           "80",
				"protocol":       "TCP",
				"container-port": "1234",
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
						"baz": "blah",
					},
					Ports: []api.ServicePort{
						{
							Port:       80,
							Protocol:   "TCP",
							TargetPort: intstr.FromInt(1234),
						},
					},
				},
			},
		},
		{

			generator: ServiceGeneratorV2{},
			params: map[string]interface{}{
				"selector":       "foo=bar,baz=blah",
				"name":           "test",
				"port":           "80",
				"protocol":       "UDP",
				"container-port": "foobar",
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
						"baz": "blah",
					},
					Ports: []api.ServicePort{
						{
							Port:       80,
							Protocol:   "UDP",
							TargetPort: intstr.FromString("foobar"),
						},
					},
				},
			},
		},
		{
			generator: ServiceGeneratorV2{},
			params: map[string]interface{}{
				"selector":       "foo=bar,baz=blah",
				"labels":         "key1=value1,key2=value2",
				"name":           "test",
				"port":           "80",
				"protocol":       "TCP",
				"container-port": "1234",
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
						"baz": "blah",
					},
					Ports: []api.ServicePort{
						{
							Port:       80,
							Protocol:   "TCP",
							TargetPort: intstr.FromInt(1234),
						},
					},
				},
			},
		},
		{
			generator: ServiceGeneratorV2{},
			params: map[string]interface{}{
				"selector":       "foo=bar,baz=blah",
				"name":           "test",
				"port":           "80",
				"protocol":       "UDP",
				"container-port": "foobar",
				"external-ip":    "1.2.3.4",
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
						"baz": "blah",
					},
					Ports: []api.ServicePort{
						{
							Port:       80,
							Protocol:   "UDP",
							TargetPort: intstr.FromString("foobar"),
						},
					},
					ExternalIPs: []string{"1.2.3.4"},
				},
			},
		},
		{
			generator: ServiceGeneratorV2{},
			params: map[string]interface{}{
				"selector":                      "foo=bar,baz=blah",
				"name":                          "test",
				"port":                          "80",
				"protocol":                      "UDP",
				"container-port":                "foobar",
				"external-ip":                   "1.2.3.4",
				"create-external-load-balancer": "true",
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
						"baz": "blah",
					},
					Ports: []api.ServicePort{
						{
							Port:       80,
							Protocol:   "UDP",
							TargetPort: intstr.FromString("foobar"),
						},
					},
					Type:        api.ServiceTypeLoadBalancer,
					ExternalIPs: []string{"1.2.3.4"},
				},
			},
		},
		{
			generator: ServiceGeneratorV2{},
			params: map[string]interface{}{
				"selector":       "foo=bar,baz=blah",
				"name":           "test",
				"port":           "80",
				"protocol":       "UDP",
				"container-port": "foobar",
				"type":           string(api.ServiceTypeNodePort),
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
						"baz": "blah",
					},
					Ports: []api.ServicePort{
						{
							Port:       80,
							Protocol:   "UDP",
							TargetPort: intstr.FromString("foobar"),
						},
					},
					Type: api.ServiceTypeNodePort,
				},
			},
		},
		{
			generator: ServiceGeneratorV2{},
			params: map[string]interface{}{
				"selector":                      "foo=bar,baz=blah",
				"name":                          "test",
				"port":                          "80",
				"protocol":                      "UDP",
				"container-port":                "foobar",
				"create-external-load-balancer": "true", // ignored when type is present
				"type": string(api.ServiceTypeNodePort),
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
						"baz": "blah",
					},
					Ports: []api.ServicePort{
						{
							Port:       80,
							Protocol:   "UDP",
							TargetPort: intstr.FromString("foobar"),
						},
					},
					Type: api.ServiceTypeNodePort,
				},
			},
		},
		{
			generator: ServiceGeneratorV1{},
			params: map[string]interface{}{
				"selector":       "foo=bar,baz=blah",
				"name":           "test",
				"port":           "80",
				"protocol":       "TCP",
				"container-port": "1234",
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
						"baz": "blah",
					},
					Ports: []api.ServicePort{
						{
							Name:       "default",
							Port:       80,
							Protocol:   "TCP",
							TargetPort: intstr.FromInt(1234),
						},
					},
				},
			},
		},
		{
			generator: ServiceGeneratorV1{},
			params: map[string]interface{}{
				"selector":         "foo=bar,baz=blah",
				"name":             "test",
				"port":             "80",
				"protocol":         "TCP",
				"container-port":   "1234",
				"session-affinity": "ClientIP",
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
						"baz": "blah",
					},
					Ports: []api.ServicePort{
						{
							Name:       "default",
							Port:       80,
							Protocol:   "TCP",
							TargetPort: intstr.FromInt(1234),
						},
					},
					SessionAffinity: api.ServiceAffinityClientIP,
				},
			},
		},
		{
			generator: ServiceGeneratorV1{},
			params: map[string]interface{}{
				"selector":       "foo=bar",
				"name":           "test",
				"ports":          "80,443",
				"protocol":       "TCP",
				"container-port": "foobar",
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
					},
					Ports: []api.ServicePort{
						{
							Name:       "port-1",
							Port:       80,
							Protocol:   api.ProtocolTCP,
							TargetPort: intstr.FromString("foobar"),
						},
						{
							Name:       "port-2",
							Port:       443,
							Protocol:   api.ProtocolTCP,
							TargetPort: intstr.FromString("foobar"),
						},
					},
				},
			},
		},
		{
			generator: ServiceGeneratorV2{},
			params: map[string]interface{}{
				"selector":    "foo=bar",
				"name":        "test",
				"ports":       "80,443",
				"protocol":    "UDP",
				"target-port": "1234",
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
					},
					Ports: []api.ServicePort{
						{
							Name:       "port-1",
							Port:       80,
							Protocol:   api.ProtocolUDP,
							TargetPort: intstr.FromInt(1234),
						},
						{
							Name:       "port-2",
							Port:       443,
							Protocol:   api.ProtocolUDP,
							TargetPort: intstr.FromInt(1234),
						},
					},
				},
			},
		},
		{
			generator: ServiceGeneratorV2{},
			params: map[string]interface{}{
				"selector": "foo=bar",
				"name":     "test",
				"ports":    "80,443",
				"protocol": "TCP",
			},
			expected: api.Service{
				ObjectMeta: api.ObjectMeta{
					Name: "test",
				},
				Spec: api.ServiceSpec{
					Selector: map[string]string{
						"foo": "bar",
					},
					Ports: []api.ServicePort{
						{
							Name:       "port-1",
							Port:       80,
							Protocol:   api.ProtocolTCP,
							TargetPort: intstr.FromInt(80),
						},
						{
							Name:       "port-2",
							Port:       443,
							Protocol:   api.ProtocolTCP,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		obj, err := test.generator.Generate(test.params)
		if !reflect.DeepEqual(obj, &test.expected) {
			t.Errorf("expected:\n%#v\ngot\n%#v\n", &test.expected, obj)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
}
