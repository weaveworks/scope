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

	internal "k8s.io/kubernetes/pkg/api"
	api "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/runtime"
)

func encodeOrDie(obj runtime.Object) []byte {
	data, err := runtime.Encode(internal.Codecs.LegacyCodec(api.SchemeGroupVersion), obj)
	if err != nil {
		panic(err.Error())
	}
	return data
}

func TestSortingPrinter(t *testing.T) {
	intPtr := func(val int32) *int32 { return &val }

	a := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name: "a",
		},
	}

	b := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name: "b",
		},
	}

	c := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name: "c",
		},
	}

	tests := []struct {
		obj   runtime.Object
		sort  runtime.Object
		field string
		name  string
	}{
		{
			name: "in-order-already",
			obj: &api.PodList{
				Items: []api.Pod{
					{
						ObjectMeta: api.ObjectMeta{
							Name: "a",
						},
					},
					{
						ObjectMeta: api.ObjectMeta{
							Name: "b",
						},
					},
					{
						ObjectMeta: api.ObjectMeta{
							Name: "c",
						},
					},
				},
			},
			sort: &api.PodList{
				Items: []api.Pod{
					{
						ObjectMeta: api.ObjectMeta{
							Name: "a",
						},
					},
					{
						ObjectMeta: api.ObjectMeta{
							Name: "b",
						},
					},
					{
						ObjectMeta: api.ObjectMeta{
							Name: "c",
						},
					},
				},
			},
			field: "{.metadata.name}",
		},
		{
			name: "reverse-order",
			obj: &api.PodList{
				Items: []api.Pod{
					{
						ObjectMeta: api.ObjectMeta{
							Name: "b",
						},
					},
					{
						ObjectMeta: api.ObjectMeta{
							Name: "c",
						},
					},
					{
						ObjectMeta: api.ObjectMeta{
							Name: "a",
						},
					},
				},
			},
			sort: &api.PodList{
				Items: []api.Pod{
					{
						ObjectMeta: api.ObjectMeta{
							Name: "a",
						},
					},
					{
						ObjectMeta: api.ObjectMeta{
							Name: "b",
						},
					},
					{
						ObjectMeta: api.ObjectMeta{
							Name: "c",
						},
					},
				},
			},
			field: "{.metadata.name}",
		},
		{
			name: "random-order-numbers",
			obj: &api.ReplicationControllerList{
				Items: []api.ReplicationController{
					{
						Spec: api.ReplicationControllerSpec{
							Replicas: intPtr(5),
						},
					},
					{
						Spec: api.ReplicationControllerSpec{
							Replicas: intPtr(1),
						},
					},
					{
						Spec: api.ReplicationControllerSpec{
							Replicas: intPtr(9),
						},
					},
				},
			},
			sort: &api.ReplicationControllerList{
				Items: []api.ReplicationController{
					{
						Spec: api.ReplicationControllerSpec{
							Replicas: intPtr(1),
						},
					},
					{
						Spec: api.ReplicationControllerSpec{
							Replicas: intPtr(5),
						},
					},
					{
						Spec: api.ReplicationControllerSpec{
							Replicas: intPtr(9),
						},
					},
				},
			},
			field: "{.spec.replicas}",
		},
		{
			name: "v1.List in order",
			obj: &api.List{
				Items: []runtime.RawExtension{
					{Raw: encodeOrDie(a)},
					{Raw: encodeOrDie(b)},
					{Raw: encodeOrDie(c)},
				},
			},
			sort: &api.List{
				Items: []runtime.RawExtension{
					{Raw: encodeOrDie(a)},
					{Raw: encodeOrDie(b)},
					{Raw: encodeOrDie(c)},
				},
			},
			field: "{.metadata.name}",
		},
		{
			name: "v1.List in reverse",
			obj: &api.List{
				Items: []runtime.RawExtension{
					{Raw: encodeOrDie(c)},
					{Raw: encodeOrDie(b)},
					{Raw: encodeOrDie(a)},
				},
			},
			sort: &api.List{
				Items: []runtime.RawExtension{
					{Raw: encodeOrDie(a)},
					{Raw: encodeOrDie(b)},
					{Raw: encodeOrDie(c)},
				},
			},
			field: "{.metadata.name}",
		},
	}
	for _, test := range tests {
		sort := &SortingPrinter{SortField: test.field, Decoder: internal.Codecs.UniversalDecoder()}
		if err := sort.sortObj(test.obj); err != nil {
			t.Errorf("unexpected error: %v (%s)", err, test.name)
			continue
		}
		if !reflect.DeepEqual(test.obj, test.sort) {
			t.Errorf("[%s]\nexpected:\n%v\nsaw:\n%v", test.name, test.sort, test.obj)
		}
	}
}
