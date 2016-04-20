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

package unversioned

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"
)

type PersistentVolumesInterface interface {
	PersistentVolumes() PersistentVolumeInterface
}

// PersistentVolumeInterface has methods to work with PersistentVolume resources.
type PersistentVolumeInterface interface {
	List(opts api.ListOptions) (*api.PersistentVolumeList, error)
	Get(name string) (*api.PersistentVolume, error)
	Create(volume *api.PersistentVolume) (*api.PersistentVolume, error)
	Update(volume *api.PersistentVolume) (*api.PersistentVolume, error)
	UpdateStatus(persistentVolume *api.PersistentVolume) (*api.PersistentVolume, error)
	Delete(name string) error
	Watch(opts api.ListOptions) (watch.Interface, error)
}

// persistentVolumes implements PersistentVolumesInterface
type persistentVolumes struct {
	client *Client
}

func newPersistentVolumes(c *Client) *persistentVolumes {
	return &persistentVolumes{c}
}

func (c *persistentVolumes) List(opts api.ListOptions) (result *api.PersistentVolumeList, err error) {
	result = &api.PersistentVolumeList{}
	err = c.client.Get().
		Resource("persistentVolumes").
		VersionedParams(&opts, api.ParameterCodec).
		Do().
		Into(result)

	return result, err
}

func (c *persistentVolumes) Get(name string) (result *api.PersistentVolume, err error) {
	result = &api.PersistentVolume{}
	err = c.client.Get().Resource("persistentVolumes").Name(name).Do().Into(result)
	return
}

func (c *persistentVolumes) Create(volume *api.PersistentVolume) (result *api.PersistentVolume, err error) {
	result = &api.PersistentVolume{}
	err = c.client.Post().Resource("persistentVolumes").Body(volume).Do().Into(result)
	return
}

func (c *persistentVolumes) Update(volume *api.PersistentVolume) (result *api.PersistentVolume, err error) {
	result = &api.PersistentVolume{}
	err = c.client.Put().Resource("persistentVolumes").Name(volume.Name).Body(volume).Do().Into(result)
	return
}

func (c *persistentVolumes) UpdateStatus(volume *api.PersistentVolume) (result *api.PersistentVolume, err error) {
	result = &api.PersistentVolume{}
	err = c.client.Put().Resource("persistentVolumes").Name(volume.Name).SubResource("status").Body(volume).Do().Into(result)
	return
}

func (c *persistentVolumes) Delete(name string) error {
	return c.client.Delete().Resource("persistentVolumes").Name(name).Do().Error()
}

func (c *persistentVolumes) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.client.Get().
		Prefix("watch").
		Resource("persistentVolumes").
		VersionedParams(&opts, api.ParameterCodec).
		Watch()
}
