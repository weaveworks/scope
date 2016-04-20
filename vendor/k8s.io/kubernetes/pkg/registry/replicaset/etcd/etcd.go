/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

// If you make changes to this file, you should also make the corresponding change in ReplicationController.

package etcd

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/apis/extensions"
	extvalidation "k8s.io/kubernetes/pkg/apis/extensions/validation"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/cachesize"
	"k8s.io/kubernetes/pkg/registry/generic"
	etcdgeneric "k8s.io/kubernetes/pkg/registry/generic/etcd"
	"k8s.io/kubernetes/pkg/registry/replicaset"
	"k8s.io/kubernetes/pkg/runtime"
)

// ReplicaSetStorage includes dummy storage for ReplicaSets and for Scale subresource.
type ReplicaSetStorage struct {
	ReplicaSet *REST
	Status     *StatusREST
	Scale      *ScaleREST
}

func NewStorage(opts generic.RESTOptions) ReplicaSetStorage {
	replicaSetRest, replicaSetStatusRest := NewREST(opts)
	replicaSetRegistry := replicaset.NewRegistry(replicaSetRest)

	return ReplicaSetStorage{
		ReplicaSet: replicaSetRest,
		Status:     replicaSetStatusRest,
		Scale:      &ScaleREST{registry: replicaSetRegistry},
	}
}

type REST struct {
	*etcdgeneric.Etcd
}

// NewREST returns a RESTStorage object that will work against ReplicaSet.
func NewREST(opts generic.RESTOptions) (*REST, *StatusREST) {
	prefix := "/replicasets"

	newListFunc := func() runtime.Object { return &extensions.ReplicaSetList{} }
	storageInterface := opts.Decorator(
		opts.Storage, cachesize.GetWatchCacheSizeByResource(cachesize.Replicasets), &extensions.ReplicaSet{}, prefix, replicaset.Strategy, newListFunc)

	store := &etcdgeneric.Etcd{
		NewFunc: func() runtime.Object { return &extensions.ReplicaSet{} },

		// NewListFunc returns an object capable of storing results of an etcd list.
		NewListFunc: newListFunc,
		// Produces a path that etcd understands, to the root of the resource
		// by combining the namespace in the context with the given prefix
		KeyRootFunc: func(ctx api.Context) string {
			return etcdgeneric.NamespaceKeyRootFunc(ctx, prefix)
		},
		// Produces a path that etcd understands, to the resource by combining
		// the namespace in the context with the given prefix
		KeyFunc: func(ctx api.Context, name string) (string, error) {
			return etcdgeneric.NamespaceKeyFunc(ctx, prefix, name)
		},
		// Retrieve the name field of a ReplicaSet
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*extensions.ReplicaSet).Name, nil
		},
		// Used to match objects based on labels/fields for list and watch
		PredicateFunc: func(label labels.Selector, field fields.Selector) generic.Matcher {
			return replicaset.MatchReplicaSet(label, field)
		},
		QualifiedResource:       api.Resource("replicasets"),
		DeleteCollectionWorkers: opts.DeleteCollectionWorkers,

		// Used to validate ReplicaSet creation
		CreateStrategy: replicaset.Strategy,

		// Used to validate ReplicaSet updates
		UpdateStrategy: replicaset.Strategy,
		DeleteStrategy: replicaset.Strategy,

		Storage: storageInterface,
	}
	statusStore := *store
	statusStore.UpdateStrategy = replicaset.StatusStrategy

	return &REST{store}, &StatusREST{store: &statusStore}
}

// StatusREST implements the REST endpoint for changing the status of a ReplicaSet
type StatusREST struct {
	store *etcdgeneric.Etcd
}

func (r *StatusREST) New() runtime.Object {
	return &extensions.ReplicaSet{}
}

// Update alters the status subset of an object.
func (r *StatusREST) Update(ctx api.Context, obj runtime.Object) (runtime.Object, bool, error) {
	return r.store.Update(ctx, obj)
}

type ScaleREST struct {
	registry replicaset.Registry
}

// ScaleREST implements Patcher
var _ = rest.Patcher(&ScaleREST{})

// New creates a new Scale object
func (r *ScaleREST) New() runtime.Object {
	return &extensions.Scale{}
}

func (r *ScaleREST) Get(ctx api.Context, name string) (runtime.Object, error) {
	rs, err := r.registry.GetReplicaSet(ctx, name)
	if err != nil {
		return nil, errors.NewNotFound(extensions.Resource("replicasets/scale"), name)
	}
	scale, err := scaleFromReplicaSet(rs)
	if err != nil {
		return nil, errors.NewBadRequest(fmt.Sprintf("%v", err))
	}
	return scale, err
}

func (r *ScaleREST) Update(ctx api.Context, obj runtime.Object) (runtime.Object, bool, error) {
	if obj == nil {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("nil update passed to Scale"))
	}
	scale, ok := obj.(*extensions.Scale)
	if !ok {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("wrong object passed to Scale update: %v", obj))
	}

	if errs := extvalidation.ValidateScale(scale); len(errs) > 0 {
		return nil, false, errors.NewInvalid(extensions.Kind("Scale"), scale.Name, errs)
	}

	rs, err := r.registry.GetReplicaSet(ctx, scale.Name)
	if err != nil {
		return nil, false, errors.NewNotFound(extensions.Resource("replicasets/scale"), scale.Name)
	}
	rs.Spec.Replicas = scale.Spec.Replicas
	rs.ResourceVersion = scale.ResourceVersion
	rs, err = r.registry.UpdateReplicaSet(ctx, rs)
	if err != nil {
		return nil, false, err
	}
	newScale, err := scaleFromReplicaSet(rs)
	if err != nil {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("%v", err))
	}
	return newScale, false, err
}

// scaleFromReplicaSet returns a scale subresource for a replica set.
func scaleFromReplicaSet(rs *extensions.ReplicaSet) (*extensions.Scale, error) {
	return &extensions.Scale{
		// TODO: Create a variant of ObjectMeta type that only contains the fields below.
		ObjectMeta: api.ObjectMeta{
			Name:              rs.Name,
			Namespace:         rs.Namespace,
			UID:               rs.UID,
			ResourceVersion:   rs.ResourceVersion,
			CreationTimestamp: rs.CreationTimestamp,
		},
		Spec: extensions.ScaleSpec{
			Replicas: rs.Spec.Replicas,
		},
		Status: extensions.ScaleStatus{
			Replicas: rs.Status.Replicas,
			Selector: rs.Spec.Selector,
		},
	}, nil
}
