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

package unversioned

import (
	api "k8s.io/kubernetes/pkg/api"
	watch "k8s.io/kubernetes/pkg/watch"
)

// ServicesGetter has a method to return a ServiceInterface.
// A group's client should implement this interface.
type ServicesGetter interface {
	Services(namespace string) ServiceInterface
}

// ServiceInterface has methods to work with Service resources.
type ServiceInterface interface {
	Create(*api.Service) (*api.Service, error)
	Update(*api.Service) (*api.Service, error)
	UpdateStatus(*api.Service) (*api.Service, error)
	Delete(name string, options *api.DeleteOptions) error
	DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error
	Get(name string) (*api.Service, error)
	List(opts api.ListOptions) (*api.ServiceList, error)
	Watch(opts api.ListOptions) (watch.Interface, error)
	ServiceExpansion
}

// services implements ServiceInterface
type services struct {
	client *CoreClient
	ns     string
}

// newServices returns a Services
func newServices(c *CoreClient, namespace string) *services {
	return &services{
		client: c,
		ns:     namespace,
	}
}

// Create takes the representation of a service and creates it.  Returns the server's representation of the service, and an error, if there is any.
func (c *services) Create(service *api.Service) (result *api.Service, err error) {
	result = &api.Service{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("services").
		Body(service).
		Do().
		Into(result)
	return
}

// Update takes the representation of a service and updates it. Returns the server's representation of the service, and an error, if there is any.
func (c *services) Update(service *api.Service) (result *api.Service, err error) {
	result = &api.Service{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("services").
		Name(service.Name).
		Body(service).
		Do().
		Into(result)
	return
}

func (c *services) UpdateStatus(service *api.Service) (result *api.Service, err error) {
	result = &api.Service{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("services").
		Name(service.Name).
		SubResource("status").
		Body(service).
		Do().
		Into(result)
	return
}

// Delete takes name of the service and deletes it. Returns an error if one occurs.
func (c *services) Delete(name string, options *api.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("services").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *services) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("services").
		VersionedParams(&listOptions, api.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the service, and returns the corresponding service object, and an error if there is any.
func (c *services) Get(name string) (result *api.Service, err error) {
	result = &api.Service{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("services").
		Name(name).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Services that match those selectors.
func (c *services) List(opts api.ListOptions) (result *api.ServiceList, err error) {
	result = &api.ServiceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("services").
		VersionedParams(&opts, api.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested services.
func (c *services) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.client.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("services").
		VersionedParams(&opts, api.ParameterCodec).
		Watch()
}
