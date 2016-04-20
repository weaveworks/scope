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

package etcd

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/registry/registrytest"
	"k8s.io/kubernetes/pkg/storage"
	"k8s.io/kubernetes/pkg/storage/etcd/etcdtest"
	etcdtesting "k8s.io/kubernetes/pkg/storage/etcd/testing"
)

func newStorage(t *testing.T) (*ScaleREST, *etcdtesting.EtcdTestServer, storage.Interface) {
	etcdStorage, server := registrytest.NewEtcdStorage(t, "")
	restOptions := generic.RESTOptions{Storage: etcdStorage, Decorator: generic.UndecoratedStorage, DeleteCollectionWorkers: 1}
	return NewStorage(restOptions).Scale, server, etcdStorage
}

var validPodTemplate = api.PodTemplate{
	Template: api.PodTemplateSpec{
		ObjectMeta: api.ObjectMeta{
			Labels: map[string]string{"a": "b"},
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{
					Name:            "test",
					Image:           "test_image",
					ImagePullPolicy: api.PullIfNotPresent,
				},
			},
			RestartPolicy: api.RestartPolicyAlways,
			DNSPolicy:     api.DNSClusterFirst,
		},
	},
}

var validReplicas = 8

var validControllerSpec = api.ReplicationControllerSpec{
	Replicas: validReplicas,
	Selector: validPodTemplate.Template.Labels,
	Template: &validPodTemplate.Template,
}

var validController = api.ReplicationController{
	ObjectMeta: api.ObjectMeta{Name: "foo", Namespace: "test"},
	Spec:       validControllerSpec,
}

var validScale = extensions.Scale{
	ObjectMeta: api.ObjectMeta{Name: "foo", Namespace: "test"},
	Spec: extensions.ScaleSpec{
		Replicas: validReplicas,
	},
	Status: extensions.ScaleStatus{
		Replicas: 0,
		Selector: &unversioned.LabelSelector{
			MatchLabels: validPodTemplate.Template.Labels,
		},
	},
}

func TestGet(t *testing.T) {
	storage, server, si := newStorage(t)
	defer server.Terminate(t)

	ctx := api.WithNamespace(api.NewContext(), "test")
	key := etcdtest.AddPrefix("/controllers/test/foo")
	if err := si.Set(ctx, key, &validController, nil, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj, err := storage.Get(ctx, "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	scale := obj.(*extensions.Scale)
	if scale.Spec.Replicas != validReplicas {
		t.Errorf("wrong replicas count expected: %d got: %d", validReplicas, scale.Spec.Replicas)
	}
}

func TestUpdate(t *testing.T) {
	storage, server, si := newStorage(t)
	defer server.Terminate(t)

	ctx := api.WithNamespace(api.NewContext(), "test")
	key := etcdtest.AddPrefix("/controllers/test/foo")
	if err := si.Set(ctx, key, &validController, nil, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	replicas := 12
	update := extensions.Scale{
		ObjectMeta: api.ObjectMeta{Name: "foo", Namespace: "test"},
		Spec: extensions.ScaleSpec{
			Replicas: replicas,
		},
	}

	if _, _, err := storage.Update(ctx, &update); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj, err := storage.Get(ctx, "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := obj.(*extensions.Scale)
	if updated.Spec.Replicas != replicas {
		t.Errorf("wrong replicas count expected: %d got: %d", replicas, updated.Spec.Replicas)
	}
}
