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

package thirdpartyresourcedata

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/watch"
)

// Registry is an interface implemented by things that know how to store ThirdPartyResourceData objects.
type Registry interface {
	ListThirdPartyResourceData(ctx api.Context, options *api.ListOptions) (*extensions.ThirdPartyResourceDataList, error)
	WatchThirdPartyResourceData(ctx api.Context, options *api.ListOptions) (watch.Interface, error)
	GetThirdPartyResourceData(ctx api.Context, name string) (*extensions.ThirdPartyResourceData, error)
	CreateThirdPartyResourceData(ctx api.Context, resource *extensions.ThirdPartyResourceData) (*extensions.ThirdPartyResourceData, error)
	UpdateThirdPartyResourceData(ctx api.Context, resource *extensions.ThirdPartyResourceData) (*extensions.ThirdPartyResourceData, error)
	DeleteThirdPartyResourceData(ctx api.Context, name string) error
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

func (s *storage) ListThirdPartyResourceData(ctx api.Context, options *api.ListOptions) (*extensions.ThirdPartyResourceDataList, error) {
	obj, err := s.List(ctx, options)
	if err != nil {
		return nil, err
	}
	return obj.(*extensions.ThirdPartyResourceDataList), nil
}

func (s *storage) WatchThirdPartyResourceData(ctx api.Context, options *api.ListOptions) (watch.Interface, error) {
	return s.Watch(ctx, options)
}

func (s *storage) GetThirdPartyResourceData(ctx api.Context, name string) (*extensions.ThirdPartyResourceData, error) {
	obj, err := s.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return obj.(*extensions.ThirdPartyResourceData), nil
}

func (s *storage) CreateThirdPartyResourceData(ctx api.Context, ThirdPartyResourceData *extensions.ThirdPartyResourceData) (*extensions.ThirdPartyResourceData, error) {
	obj, err := s.Create(ctx, ThirdPartyResourceData)
	return obj.(*extensions.ThirdPartyResourceData), err
}

func (s *storage) UpdateThirdPartyResourceData(ctx api.Context, ThirdPartyResourceData *extensions.ThirdPartyResourceData) (*extensions.ThirdPartyResourceData, error) {
	obj, _, err := s.Update(ctx, ThirdPartyResourceData)
	return obj.(*extensions.ThirdPartyResourceData), err
}

func (s *storage) DeleteThirdPartyResourceData(ctx api.Context, name string) error {
	_, err := s.Delete(ctx, name, nil)
	return err
}
