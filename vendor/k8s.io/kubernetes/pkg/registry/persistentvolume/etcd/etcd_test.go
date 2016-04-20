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

package etcd

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/registry/registrytest"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/storage/etcd/etcdtest"
	etcdtesting "k8s.io/kubernetes/pkg/storage/etcd/testing"
	"k8s.io/kubernetes/pkg/util/diff"
)

func newStorage(t *testing.T) (*REST, *StatusREST, *etcdtesting.EtcdTestServer) {
	etcdStorage, server := registrytest.NewEtcdStorage(t, "")
	restOptions := generic.RESTOptions{Storage: etcdStorage, Decorator: generic.UndecoratedStorage, DeleteCollectionWorkers: 1}
	persistentVolumeStorage, statusStorage := NewREST(restOptions)
	return persistentVolumeStorage, statusStorage, server
}

func validNewPersistentVolume(name string) *api.PersistentVolume {
	pv := &api.PersistentVolume{
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Spec: api.PersistentVolumeSpec{
			Capacity: api.ResourceList{
				api.ResourceName(api.ResourceStorage): resource.MustParse("10G"),
			},
			AccessModes: []api.PersistentVolumeAccessMode{api.ReadWriteOnce},
			PersistentVolumeSource: api.PersistentVolumeSource{
				HostPath: &api.HostPathVolumeSource{Path: "/foo"},
			},
			PersistentVolumeReclaimPolicy: api.PersistentVolumeReclaimRetain,
		},
		Status: api.PersistentVolumeStatus{
			Phase:   api.VolumePending,
			Message: "bar",
			Reason:  "foo",
		},
	}
	return pv
}

func validChangedPersistentVolume() *api.PersistentVolume {
	pv := validNewPersistentVolume("foo")
	return pv
}

func TestCreate(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd).ClusterScope()
	pv := validNewPersistentVolume("foo")
	pv.ObjectMeta = api.ObjectMeta{GenerateName: "foo"}
	test.TestCreate(
		// valid
		pv,
		// invalid
		&api.PersistentVolume{
			ObjectMeta: api.ObjectMeta{Name: "*BadName!"},
		},
	)
}

func TestUpdate(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd).ClusterScope()
	test.TestUpdate(
		// valid
		validNewPersistentVolume("foo"),
		// updateFunc
		func(obj runtime.Object) runtime.Object {
			object := obj.(*api.PersistentVolume)
			object.Spec.Capacity = api.ResourceList{
				api.ResourceName(api.ResourceStorage): resource.MustParse("20G"),
			}
			return object
		},
	)
}

func TestDelete(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd).ClusterScope().ReturnDeletedObject()
	test.TestDelete(validNewPersistentVolume("foo"))
}

func TestGet(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd).ClusterScope()
	test.TestGet(validNewPersistentVolume("foo"))
}

func TestList(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd).ClusterScope()
	test.TestList(validNewPersistentVolume("foo"))
}

func TestWatch(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd).ClusterScope()
	test.TestWatch(
		validNewPersistentVolume("foo"),
		// matching labels
		[]labels.Set{},
		// not matching labels
		[]labels.Set{
			{"foo": "bar"},
		},
		// matching fields
		[]fields.Set{
			{"metadata.name": "foo"},
			{"name": "foo"},
		},
		// not matching fields
		[]fields.Set{
			{"metadata.name": "bar"},
		},
	)
}

func TestUpdateStatus(t *testing.T) {
	storage, statusStorage, server := newStorage(t)
	defer server.Terminate(t)
	ctx := api.NewContext()
	key, _ := storage.KeyFunc(ctx, "foo")
	key = etcdtest.AddPrefix(key)
	pvStart := validNewPersistentVolume("foo")
	err := storage.Storage.Set(ctx, key, pvStart, nil, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	pvIn := &api.PersistentVolume{
		ObjectMeta: api.ObjectMeta{
			Name: "foo",
		},
		Status: api.PersistentVolumeStatus{
			Phase: api.VolumeBound,
		},
	}

	_, _, err = statusStorage.Update(ctx, pvIn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	obj, err := storage.Get(ctx, "foo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	pvOut := obj.(*api.PersistentVolume)
	// only compare the relevant change b/c metadata will differ
	if !api.Semantic.DeepEqual(pvIn.Status, pvOut.Status) {
		t.Errorf("unexpected object: %s", diff.ObjectDiff(pvIn.Status, pvOut.Status))
	}
}
