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
	extensions "k8s.io/kubernetes/pkg/apis/extensions"
	core "k8s.io/kubernetes/pkg/client/testing/core"
	labels "k8s.io/kubernetes/pkg/labels"
	watch "k8s.io/kubernetes/pkg/watch"
)

// FakeDaemonSets implements DaemonSetInterface
type FakeDaemonSets struct {
	Fake *FakeExtensions
	ns   string
}

func (c *FakeDaemonSets) Create(daemonSet *extensions.DaemonSet) (result *extensions.DaemonSet, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction("daemonsets", c.ns, daemonSet), &extensions.DaemonSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*extensions.DaemonSet), err
}

func (c *FakeDaemonSets) Update(daemonSet *extensions.DaemonSet) (result *extensions.DaemonSet, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction("daemonsets", c.ns, daemonSet), &extensions.DaemonSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*extensions.DaemonSet), err
}

func (c *FakeDaemonSets) UpdateStatus(daemonSet *extensions.DaemonSet) (*extensions.DaemonSet, error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateSubresourceAction("daemonsets", "status", c.ns, daemonSet), &extensions.DaemonSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*extensions.DaemonSet), err
}

func (c *FakeDaemonSets) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction("daemonsets", c.ns, name), &extensions.DaemonSet{})

	return err
}

func (c *FakeDaemonSets) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction("daemonsets", c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &extensions.DaemonSetList{})
	return err
}

func (c *FakeDaemonSets) Get(name string) (result *extensions.DaemonSet, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction("daemonsets", c.ns, name), &extensions.DaemonSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*extensions.DaemonSet), err
}

func (c *FakeDaemonSets) List(opts api.ListOptions) (result *extensions.DaemonSetList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction("daemonsets", c.ns, opts), &extensions.DaemonSetList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &extensions.DaemonSetList{}
	for _, item := range obj.(*extensions.DaemonSetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested daemonSets.
func (c *FakeDaemonSets) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction("daemonsets", c.ns, opts))

}
