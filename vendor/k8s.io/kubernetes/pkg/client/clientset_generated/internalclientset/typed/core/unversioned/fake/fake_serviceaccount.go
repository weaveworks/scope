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

// FakeServiceAccounts implements ServiceAccountInterface
type FakeServiceAccounts struct {
	Fake *FakeCore
	ns   string
}

func (c *FakeServiceAccounts) Create(serviceAccount *api.ServiceAccount) (result *api.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction("serviceaccounts", c.ns, serviceAccount), &api.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.ServiceAccount), err
}

func (c *FakeServiceAccounts) Update(serviceAccount *api.ServiceAccount) (result *api.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction("serviceaccounts", c.ns, serviceAccount), &api.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.ServiceAccount), err
}

func (c *FakeServiceAccounts) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction("serviceaccounts", c.ns, name), &api.ServiceAccount{})

	return err
}

func (c *FakeServiceAccounts) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction("serviceaccounts", c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &api.ServiceAccountList{})
	return err
}

func (c *FakeServiceAccounts) Get(name string) (result *api.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction("serviceaccounts", c.ns, name), &api.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*api.ServiceAccount), err
}

func (c *FakeServiceAccounts) List(opts api.ListOptions) (result *api.ServiceAccountList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction("serviceaccounts", c.ns, opts), &api.ServiceAccountList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &api.ServiceAccountList{}
	for _, item := range obj.(*api.ServiceAccountList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceAccounts.
func (c *FakeServiceAccounts) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction("serviceaccounts", c.ns, opts))

}
