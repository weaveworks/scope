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

package testclient

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/watch"
)

// Fake implements ServiceInterface. Meant to be embedded into a struct to get a default
// implementation. This makes faking out just the method you want to test easier.
type FakeServices struct {
	Fake      *Fake
	Namespace string
}

func (c *FakeServices) Get(name string) (*api.Service, error) {
	obj, err := c.Fake.Invokes(NewGetAction("services", c.Namespace, name), &api.Service{})
	if obj == nil {
		return nil, err
	}

	return obj.(*api.Service), err
}

func (c *FakeServices) List(opts api.ListOptions) (*api.ServiceList, error) {
	obj, err := c.Fake.Invokes(NewListAction("services", c.Namespace, opts), &api.ServiceList{})
	if obj == nil {
		return nil, err
	}

	return obj.(*api.ServiceList), err
}

func (c *FakeServices) Create(service *api.Service) (*api.Service, error) {
	obj, err := c.Fake.Invokes(NewCreateAction("services", c.Namespace, service), service)
	if obj == nil {
		return nil, err
	}

	return obj.(*api.Service), err
}

func (c *FakeServices) Update(service *api.Service) (*api.Service, error) {
	obj, err := c.Fake.Invokes(NewUpdateAction("services", c.Namespace, service), service)
	if obj == nil {
		return nil, err
	}

	return obj.(*api.Service), err
}

func (c *FakeServices) UpdateStatus(service *api.Service) (result *api.Service, err error) {
	obj, err := c.Fake.Invokes(NewUpdateSubresourceAction("services", "status", c.Namespace, service), service)
	if obj == nil {
		return nil, err
	}

	return obj.(*api.Service), err
}

func (c *FakeServices) Delete(name string) error {
	_, err := c.Fake.Invokes(NewDeleteAction("services", c.Namespace, name), &api.Service{})
	return err
}

func (c *FakeServices) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.InvokesWatch(NewWatchAction("services", c.Namespace, opts))
}

func (c *FakeServices) ProxyGet(scheme, name, port, path string, params map[string]string) restclient.ResponseWrapper {
	return c.Fake.InvokesProxy(NewProxyGetAction("services", c.Namespace, scheme, name, port, path, params))
}
