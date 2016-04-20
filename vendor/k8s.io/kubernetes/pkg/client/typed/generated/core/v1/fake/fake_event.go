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
	v1 "k8s.io/kubernetes/pkg/api/v1"
	core "k8s.io/kubernetes/pkg/client/testing/core"
	labels "k8s.io/kubernetes/pkg/labels"
	watch "k8s.io/kubernetes/pkg/watch"
)

// FakeEvents implements EventInterface
type FakeEvents struct {
	Fake *FakeCore
	ns   string
}

func (c *FakeEvents) Create(event *v1.Event) (result *v1.Event, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction("events", c.ns, event), &v1.Event{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Event), err
}

func (c *FakeEvents) Update(event *v1.Event) (result *v1.Event, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction("events", c.ns, event), &v1.Event{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Event), err
}

func (c *FakeEvents) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction("events", c.ns, name), &v1.Event{})

	return err
}

func (c *FakeEvents) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction("events", c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1.EventList{})
	return err
}

func (c *FakeEvents) Get(name string) (result *v1.Event, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction("events", c.ns, name), &v1.Event{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Event), err
}

func (c *FakeEvents) List(opts api.ListOptions) (result *v1.EventList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction("events", c.ns, opts), &v1.EventList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.EventList{}
	for _, item := range obj.(*v1.EventList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested events.
func (c *FakeEvents) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction("events", c.ns, opts))

}
