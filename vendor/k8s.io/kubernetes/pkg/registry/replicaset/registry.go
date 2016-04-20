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

package replicaset

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/watch"
)

// Registry is an interface for things that know how to store ReplicaSets.
type Registry interface {
	ListReplicaSets(ctx api.Context, options *api.ListOptions) (*extensions.ReplicaSetList, error)
	WatchReplicaSets(ctx api.Context, options *api.ListOptions) (watch.Interface, error)
	GetReplicaSet(ctx api.Context, replicaSetID string) (*extensions.ReplicaSet, error)
	CreateReplicaSet(ctx api.Context, replicaSet *extensions.ReplicaSet) (*extensions.ReplicaSet, error)
	UpdateReplicaSet(ctx api.Context, replicaSet *extensions.ReplicaSet) (*extensions.ReplicaSet, error)
	DeleteReplicaSet(ctx api.Context, replicaSetID string) error
}

// storage puts strong typing around storage calls
type storage struct {
	rest.StandardStorage
}

// NewRegistry returns a new Registry interface for the given Storage. Any mismatched
// types will panic.
func NewRegistry(s rest.StandardStorage) Registry {
	return &storage{s}
}

func (s *storage) ListReplicaSets(ctx api.Context, options *api.ListOptions) (*extensions.ReplicaSetList, error) {
	if options != nil && options.FieldSelector != nil && !options.FieldSelector.Empty() {
		return nil, fmt.Errorf("field selector not supported yet")
	}
	obj, err := s.List(ctx, options)
	if err != nil {
		return nil, err
	}
	return obj.(*extensions.ReplicaSetList), err
}

func (s *storage) WatchReplicaSets(ctx api.Context, options *api.ListOptions) (watch.Interface, error) {
	return s.Watch(ctx, options)
}

func (s *storage) GetReplicaSet(ctx api.Context, replicaSetID string) (*extensions.ReplicaSet, error) {
	obj, err := s.Get(ctx, replicaSetID)
	if err != nil {
		return nil, err
	}
	return obj.(*extensions.ReplicaSet), nil
}

func (s *storage) CreateReplicaSet(ctx api.Context, replicaSet *extensions.ReplicaSet) (*extensions.ReplicaSet, error) {
	obj, err := s.Create(ctx, replicaSet)
	if err != nil {
		return nil, err
	}
	return obj.(*extensions.ReplicaSet), nil
}

func (s *storage) UpdateReplicaSet(ctx api.Context, replicaSet *extensions.ReplicaSet) (*extensions.ReplicaSet, error) {
	obj, _, err := s.Update(ctx, replicaSet)
	if err != nil {
		return nil, err
	}
	return obj.(*extensions.ReplicaSet), nil
}

func (s *storage) DeleteReplicaSet(ctx api.Context, replicaSetID string) error {
	_, err := s.Delete(ctx, replicaSetID, nil)
	return err
}
