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
	persistentVolumeClaimStorage, statusStorage := NewREST(restOptions)
	return persistentVolumeClaimStorage, statusStorage, server
}

func validNewPersistentVolumeClaim(name, ns string) *api.PersistentVolumeClaim {
	pv := &api.PersistentVolumeClaim{
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: api.PersistentVolumeClaimSpec{
			AccessModes: []api.PersistentVolumeAccessMode{api.ReadWriteOnce},
			Resources: api.ResourceRequirements{
				Requests: api.ResourceList{
					api.ResourceName(api.ResourceStorage): resource.MustParse("10G"),
				},
			},
		},
		Status: api.PersistentVolumeClaimStatus{
			Phase: api.ClaimPending,
		},
	}
	return pv
}

func TestCreate(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	pv := validNewPersistentVolumeClaim("foo", api.NamespaceDefault)
	pv.ObjectMeta = api.ObjectMeta{}
	test.TestCreate(
		// valid
		pv,
		// invalid
		&api.PersistentVolumeClaim{
			ObjectMeta: api.ObjectMeta{Name: "*BadName!"},
		},
	)
}

func TestUpdate(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	test.TestUpdate(
		// valid
		validNewPersistentVolumeClaim("foo", api.NamespaceDefault),
		// updateFunc
		func(obj runtime.Object) runtime.Object {
			object := obj.(*api.PersistentVolumeClaim)
			object.Spec.Resources = api.ResourceRequirements{
				Requests: api.ResourceList{
					api.ResourceName(api.ResourceStorage): resource.MustParse("20G"),
				},
			}
			return object
		},
	)
}

func TestDelete(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd).ReturnDeletedObject()
	test.TestDelete(validNewPersistentVolumeClaim("foo", api.NamespaceDefault))
}

func TestGet(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	test.TestGet(validNewPersistentVolumeClaim("foo", api.NamespaceDefault))
}

func TestList(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	test.TestList(validNewPersistentVolumeClaim("foo", api.NamespaceDefault))
}

func TestWatch(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	test.TestWatch(
		validNewPersistentVolumeClaim("foo", api.NamespaceDefault),
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
	ctx := api.NewDefaultContext()

	key, _ := storage.KeyFunc(ctx, "foo")
	key = etcdtest.AddPrefix(key)
	pvcStart := validNewPersistentVolumeClaim("foo", api.NamespaceDefault)
	err := storage.Storage.Set(ctx, key, pvcStart, nil, 0)

	pvc := &api.PersistentVolumeClaim{
		ObjectMeta: api.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: api.PersistentVolumeClaimSpec{
			AccessModes: []api.PersistentVolumeAccessMode{api.ReadWriteOnce},
			Resources: api.ResourceRequirements{
				Requests: api.ResourceList{
					api.ResourceName(api.ResourceStorage): resource.MustParse("3Gi"),
				},
			},
		},
		Status: api.PersistentVolumeClaimStatus{
			Phase: api.ClaimBound,
		},
	}

	_, _, err = statusStorage.Update(ctx, pvc)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	obj, err := storage.Get(ctx, "foo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	pvcOut := obj.(*api.PersistentVolumeClaim)
	// only compare relevant changes b/c of difference in metadata
	if !api.Semantic.DeepEqual(pvc.Status, pvcOut.Status) {
		t.Errorf("unexpected object: %s", diff.ObjectDiff(pvc.Status, pvcOut.Status))
	}
}
