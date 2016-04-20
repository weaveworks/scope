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

package fake

import (
	api "k8s.io/kubernetes/pkg/api"
	core "k8s.io/kubernetes/pkg/client/testing/core"
	labels "k8s.io/kubernetes/pkg/labels"
	watch "k8s.io/kubernetes/pkg/watch"
)

// FakePersistentVolumeClaims implements PersistentVolumeClaimInterface
type FakePersistentVolumeClaims struct {
	Fake *FakeCore
	ns   string
}

func (c *FakePersistentVolumeClaims) Create(persistentVolumeClaim *api.PersistentVolumeClaim) (result *api.PersistentVolumeClaim, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction("persistentvolumeclaims", c.ns, persistentVolumeClaim), &api.PersistentVolumeClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.PersistentVolumeClaim), err
}

func (c *FakePersistentVolumeClaims) Update(persistentVolumeClaim *api.PersistentVolumeClaim) (result *api.PersistentVolumeClaim, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction("persistentvolumeclaims", c.ns, persistentVolumeClaim), &api.PersistentVolumeClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.PersistentVolumeClaim), err
}

func (c *FakePersistentVolumeClaims) UpdateStatus(persistentVolumeClaim *api.PersistentVolumeClaim) (*api.PersistentVolumeClaim, error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateSubresourceAction("persistentvolumeclaims", "status", c.ns, persistentVolumeClaim), &api.PersistentVolumeClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.PersistentVolumeClaim), err
}

func (c *FakePersistentVolumeClaims) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction("persistentvolumeclaims", c.ns, name), &api.PersistentVolumeClaim{})

	return err
}

func (c *FakePersistentVolumeClaims) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction("persistentvolumeclaims", c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &api.PersistentVolumeClaimList{})
	return err
}

func (c *FakePersistentVolumeClaims) Get(name string) (result *api.PersistentVolumeClaim, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction("persistentvolumeclaims", c.ns, name), &api.PersistentVolumeClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.PersistentVolumeClaim), err
}

func (c *FakePersistentVolumeClaims) List(opts api.ListOptions) (result *api.PersistentVolumeClaimList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction("persistentvolumeclaims", c.ns, opts), &api.PersistentVolumeClaimList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &api.PersistentVolumeClaimList{}
	for _, item := range obj.(*api.PersistentVolumeClaimList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested persistentVolumeClaims.
func (c *FakePersistentVolumeClaims) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction("persistentvolumeclaims", c.ns, opts))

}
