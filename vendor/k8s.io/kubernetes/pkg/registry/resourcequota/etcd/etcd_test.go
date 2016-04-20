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
	"k8s.io/kubernetes/pkg/storage/etcd/etcdtest"
	etcdtesting "k8s.io/kubernetes/pkg/storage/etcd/testing"
	"k8s.io/kubernetes/pkg/util/diff"
)

func newStorage(t *testing.T) (*REST, *StatusREST, *etcdtesting.EtcdTestServer) {
	etcdStorage, server := registrytest.NewEtcdStorage(t, "")
	restOptions := generic.RESTOptions{Storage: etcdStorage, Decorator: generic.UndecoratedStorage, DeleteCollectionWorkers: 1}
	resourceQuotaStorage, statusStorage := NewREST(restOptions)
	return resourceQuotaStorage, statusStorage, server
}

func validNewResourceQuota() *api.ResourceQuota {
	return &api.ResourceQuota{
		ObjectMeta: api.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Spec: api.ResourceQuotaSpec{
			Hard: api.ResourceList{
				api.ResourceCPU:                    resource.MustParse("100"),
				api.ResourceMemory:                 resource.MustParse("4Gi"),
				api.ResourcePods:                   resource.MustParse("10"),
				api.ResourceServices:               resource.MustParse("10"),
				api.ResourceReplicationControllers: resource.MustParse("10"),
				api.ResourceQuotas:                 resource.MustParse("1"),
			},
		},
	}
}

func TestCreate(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	resourcequota := validNewResourceQuota()
	resourcequota.ObjectMeta = api.ObjectMeta{}
	test.TestCreate(
		// valid
		resourcequota,
		// invalid
		&api.ResourceQuota{
			ObjectMeta: api.ObjectMeta{Name: "_-a123-a_"},
		},
	)
}

func TestCreateSetsFields(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	ctx := api.NewDefaultContext()
	resourcequota := validNewResourceQuota()
	_, err := storage.Create(api.NewDefaultContext(), resourcequota)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	object, err := storage.Get(ctx, "foo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	actual := object.(*api.ResourceQuota)
	if actual.Name != resourcequota.Name {
		t.Errorf("unexpected resourcequota: %#v", actual)
	}
	if len(actual.UID) == 0 {
		t.Errorf("expected resourcequota UID to be set: %#v", actual)
	}
}

func TestDelete(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd).ReturnDeletedObject()
	test.TestDelete(validNewResourceQuota())
}

func TestGet(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	test.TestGet(validNewResourceQuota())
}

func TestList(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	test.TestList(validNewResourceQuota())
}

func TestWatch(t *testing.T) {
	storage, _, server := newStorage(t)
	defer server.Terminate(t)
	test := registrytest.New(t, storage.Etcd)
	test.TestWatch(
		validNewResourceQuota(),
		// matching labels
		[]labels.Set{},
		// not matching labels
		[]labels.Set{
			{"foo": "bar"},
		},
		// matching fields
		[]fields.Set{
			{"metadata.name": "foo"},
		},
		// not matchin fields
		[]fields.Set{
			{"metadata.name": "bar"},
		},
	)
}

func TestUpdateStatus(t *testing.T) {
	storage, status, server := newStorage(t)
	defer server.Terminate(t)
	ctx := api.NewDefaultContext()

	key, _ := storage.KeyFunc(ctx, "foo")
	key = etcdtest.AddPrefix(key)
	resourcequotaStart := validNewResourceQuota()
	err := storage.Storage.Set(ctx, key, resourcequotaStart, nil, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resourcequotaIn := &api.ResourceQuota{
		ObjectMeta: api.ObjectMeta{
			Name:      "foo",
			Namespace: api.NamespaceDefault,
		},
		Status: api.ResourceQuotaStatus{
			Used: api.ResourceList{
				api.ResourceCPU:                    resource.MustParse("1"),
				api.ResourceMemory:                 resource.MustParse("1Gi"),
				api.ResourcePods:                   resource.MustParse("1"),
				api.ResourceServices:               resource.MustParse("1"),
				api.ResourceReplicationControllers: resource.MustParse("1"),
				api.ResourceQuotas:                 resource.MustParse("1"),
			},
			Hard: api.ResourceList{
				api.ResourceCPU:                    resource.MustParse("100"),
				api.ResourceMemory:                 resource.MustParse("4Gi"),
				api.ResourcePods:                   resource.MustParse("10"),
				api.ResourceServices:               resource.MustParse("10"),
				api.ResourceReplicationControllers: resource.MustParse("10"),
				api.ResourceQuotas:                 resource.MustParse("1"),
			},
		},
	}

	_, _, err = status.Update(ctx, resourcequotaIn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	obj, err := storage.Get(ctx, "foo")
	rqOut := obj.(*api.ResourceQuota)
	// only compare the meaningful update b/c we can't compare due to metadata
	if !api.Semantic.DeepEqual(resourcequotaIn.Status, rqOut.Status) {
		t.Errorf("unexpected object: %s", diff.ObjectDiff(resourcequotaIn, rqOut))
	}
}
