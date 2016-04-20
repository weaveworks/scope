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

// FakeEndpoints implements EndpointsInterface
type FakeEndpoints struct {
	Fake *FakeCore
	ns   string
}

func (c *FakeEndpoints) Create(endpoints *v1.Endpoints) (result *v1.Endpoints, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction("endpoints", c.ns, endpoints), &v1.Endpoints{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Endpoints), err
}

func (c *FakeEndpoints) Update(endpoints *v1.Endpoints) (result *v1.Endpoints, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction("endpoints", c.ns, endpoints), &v1.Endpoints{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Endpoints), err
}

func (c *FakeEndpoints) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction("endpoints", c.ns, name), &v1.Endpoints{})

	return err
}

func (c *FakeEndpoints) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction("endpoints", c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1.EndpointsList{})
	return err
}

func (c *FakeEndpoints) Get(name string) (result *v1.Endpoints, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction("endpoints", c.ns, name), &v1.Endpoints{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Endpoints), err
}

func (c *FakeEndpoints) List(opts api.ListOptions) (result *v1.EndpointsList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction("endpoints", c.ns, opts), &v1.EndpointsList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.EndpointsList{}
	for _, item := range obj.(*v1.EndpointsList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested endpoints.
func (c *FakeEndpoints) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction("endpoints", c.ns, opts))

}
