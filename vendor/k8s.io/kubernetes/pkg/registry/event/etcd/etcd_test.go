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
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/registry/registrytest"
	"k8s.io/kubernetes/pkg/runtime"
	etcdtesting "k8s.io/kubernetes/pkg/storage/etcd/testing"
)

var testTTL uint64 = 60

func newStorage(t *testing.T) (*REST, *etcdtesting.EtcdTestServer) {
	etcdStorage, server := registrytest.NewEtcdStorage(t, "")
	restOptions := generic.RESTOptions{Storage: etcdStorage, Decorator: generic.UndecoratedStorage, DeleteCollectionWorkers: 1}
	return NewREST(restOptions, testTTL), server
}

func validNewEvent(namespace string) *api.Event {
	return &api.Event{
		ObjectMeta: api.ObjectMeta{
			Name:      "foo",
			Namespace: namespace,
		},
		Reason: "forTesting",
		InvolvedObject: api.ObjectReference{
			Name:      "bar",
			Namespace: namespace,
		},
	}
}

func TestCreate(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	event := validNewEvent(test.TestNamespace())
	event.ObjectMeta = api.ObjectMeta{}
	test.TestCreate(
		// valid
		event,
		// invalid
		&api.Event{},
	)
}

func TestUpdate(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd).AllowCreateOnUpdate()
	test.TestUpdate(
		// valid
		validNewEvent(test.TestNamespace()),
		// valid updateFunc
		func(obj runtime.Object) runtime.Object {
			object := obj.(*api.Event)
			object.Reason = "forDifferentTesting"
			return object
		},
		// invalid updateFunc
		func(obj runtime.Object) runtime.Object {
			object := obj.(*api.Event)
			object.InvolvedObject.Namespace = "different-namespace"
			return object
		},
	)
}

func TestDelete(t *testing.T) {
	storage, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	test.TestDelete(validNewEvent(test.TestNamespace()))
}
