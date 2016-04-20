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

// FakeReplicationControllers implements ReplicationControllerInterface
type FakeReplicationControllers struct {
	Fake *FakeCore
	ns   string
}

func (c *FakeReplicationControllers) Create(replicationController *api.ReplicationController) (result *api.ReplicationController, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction("replicationcontrollers", c.ns, replicationController), &api.ReplicationController{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.ReplicationController), err
}

func (c *FakeReplicationControllers) Update(replicationController *api.ReplicationController) (result *api.ReplicationController, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction("replicationcontrollers", c.ns, replicationController), &api.ReplicationController{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.ReplicationController), err
}

func (c *FakeReplicationControllers) UpdateStatus(replicationController *api.ReplicationController) (*api.ReplicationController, error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateSubresourceAction("replicationcontrollers", "status", c.ns, replicationController), &api.ReplicationController{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.ReplicationController), err
}

func (c *FakeReplicationControllers) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction("replicationcontrollers", c.ns, name), &api.ReplicationController{})

	return err
}

func (c *FakeReplicationControllers) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction("replicationcontrollers", c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &api.ReplicationControllerList{})
	return err
}

func (c *FakeReplicationControllers) Get(name string) (result *api.ReplicationController, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction("replicationcontrollers", c.ns, name), &api.ReplicationController{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.ReplicationController), err
}

func (c *FakeReplicationControllers) List(opts api.ListOptions) (result *api.ReplicationControllerList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction("replicationcontrollers", c.ns, opts), &api.ReplicationControllerList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &api.ReplicationControllerList{}
	for _, item := range obj.(*api.ReplicationControllerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested replicationControllers.
func (c *FakeReplicationControllers) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction("replicationcontrollers", c.ns, opts))

}
