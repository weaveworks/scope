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
	v1beta1 "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	core "k8s.io/kubernetes/pkg/client/testing/core"
	labels "k8s.io/kubernetes/pkg/labels"
	watch "k8s.io/kubernetes/pkg/watch"
)

// FakeDeployments implements DeploymentInterface
type FakeDeployments struct {
	Fake *FakeExtensions
	ns   string
}

func (c *FakeDeployments) Create(deployment *v1beta1.Deployment) (result *v1beta1.Deployment, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction("deployments", c.ns, deployment), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}

func (c *FakeDeployments) Update(deployment *v1beta1.Deployment) (result *v1beta1.Deployment, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction("deployments", c.ns, deployment), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}

func (c *FakeDeployments) UpdateStatus(deployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateSubresourceAction("deployments", "status", c.ns, deployment), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}

func (c *FakeDeployments) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction("deployments", c.ns, name), &v1beta1.Deployment{})

	return err
}

func (c *FakeDeployments) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction("deployments", c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.DeploymentList{})
	return err
}

func (c *FakeDeployments) Get(name string) (result *v1beta1.Deployment, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction("deployments", c.ns, name), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}

func (c *FakeDeployments) List(opts api.ListOptions) (result *v1beta1.DeploymentList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction("deployments", c.ns, opts), &v1beta1.DeploymentList{})

	if obj == nil {
		return nil, err
	}

	label := opts.LabelSelector
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.DeploymentList{}
	for _, item := range obj.(*v1beta1.DeploymentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested deployments.
func (c *FakeDeployments) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction("deployments", c.ns, opts))

}
