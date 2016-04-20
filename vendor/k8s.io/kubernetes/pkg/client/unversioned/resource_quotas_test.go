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

package unversioned_test

import (
	"net/url"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/client/unversioned/testclient/simple"
)

func getResourceQuotasResoureName() string {
	return "resourcequotas"
}

func TestResourceQuotaCreate(t *testing.T) {
	ns := api.NamespaceDefault
	resourceQuota := &api.ResourceQuota{
		ObjectMeta: api.ObjectMeta{
			Name:      "abc",
			Namespace: "foo",
		},
		Spec: api.ResourceQuotaSpec{
			Hard: api.ResourceList{
				api.ResourceCPU:                    resource.MustParse("100"),
				api.ResourceMemory:                 resource.MustParse("10000"),
				api.ResourcePods:                   resource.MustParse("10"),
				api.ResourceServices:               resource.MustParse("10"),
				api.ResourceReplicationControllers: resource.MustParse("10"),
				api.ResourceQuotas:                 resource.MustParse("10"),
			},
		},
	}
	c := &simple.Client{
		Request: simple.Request{
			Method: "POST",
			Path:   testapi.Default.ResourcePath(getResourceQuotasResoureName(), ns, ""),
			Query:  simple.BuildQueryValues(nil),
			Body:   resourceQuota,
		},
		Response: simple.Response{StatusCode: 200, Body: resourceQuota},
	}

	response, err := c.Setup(t).ResourceQuotas(ns).Create(resourceQuota)
	defer c.Close()
	c.Validate(t, response, err)
}

func TestResourceQuotaGet(t *testing.T) {
	ns := api.NamespaceDefault
	resourceQuota := &api.ResourceQuota{
		ObjectMeta: api.ObjectMeta{
			Name:      "abc",
			Namespace: "foo",
		},
		Spec: api.ResourceQuotaSpec{
			Hard: api.ResourceList{
				api.ResourceCPU:                    resource.MustParse("100"),
				api.ResourceMemory:                 resource.MustParse("10000"),
				api.ResourcePods:                   resource.MustParse("10"),
				api.ResourceServices:               resource.MustParse("10"),
				api.ResourceReplicationControllers: resource.MustParse("10"),
				api.ResourceQuotas:                 resource.MustParse("10"),
			},
		},
	}
	c := &simple.Client{
		Request: simple.Request{
			Method: "GET",
			Path:   testapi.Default.ResourcePath(getResourceQuotasResoureName(), ns, "abc"),
			Query:  simple.BuildQueryValues(nil),
			Body:   nil,
		},
		Response: simple.Response{StatusCode: 200, Body: resourceQuota},
	}

	response, err := c.Setup(t).ResourceQuotas(ns).Get("abc")
	defer c.Close()
	c.Validate(t, response, err)
}

func TestResourceQuotaList(t *testing.T) {
	ns := api.NamespaceDefault

	resourceQuotaList := &api.ResourceQuotaList{
		Items: []api.ResourceQuota{
			{
				ObjectMeta: api.ObjectMeta{Name: "foo"},
			},
		},
	}
	c := &simple.Client{
		Request: simple.Request{
			Method: "GET",
			Path:   testapi.Default.ResourcePath(getResourceQuotasResoureName(), ns, ""),
			Query:  simple.BuildQueryValues(nil),
			Body:   nil,
		},
		Response: simple.Response{StatusCode: 200, Body: resourceQuotaList},
	}
	response, err := c.Setup(t).ResourceQuotas(ns).List(api.ListOptions{})
	defer c.Close()
	c.Validate(t, response, err)
}

func TestResourceQuotaUpdate(t *testing.T) {
	ns := api.NamespaceDefault
	resourceQuota := &api.ResourceQuota{
		ObjectMeta: api.ObjectMeta{
			Name:            "abc",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: api.ResourceQuotaSpec{
			Hard: api.ResourceList{
				api.ResourceCPU:                    resource.MustParse("100"),
				api.ResourceMemory:                 resource.MustParse("10000"),
				api.ResourcePods:                   resource.MustParse("10"),
				api.ResourceServices:               resource.MustParse("10"),
				api.ResourceReplicationControllers: resource.MustParse("10"),
				api.ResourceQuotas:                 resource.MustParse("10"),
			},
		},
	}
	c := &simple.Client{
		Request:  simple.Request{Method: "PUT", Path: testapi.Default.ResourcePath(getResourceQuotasResoureName(), ns, "abc"), Query: simple.BuildQueryValues(nil)},
		Response: simple.Response{StatusCode: 200, Body: resourceQuota},
	}
	response, err := c.Setup(t).ResourceQuotas(ns).Update(resourceQuota)
	defer c.Close()
	c.Validate(t, response, err)
}

func TestResourceQuotaStatusUpdate(t *testing.T) {
	ns := api.NamespaceDefault
	resourceQuota := &api.ResourceQuota{
		ObjectMeta: api.ObjectMeta{
			Name:            "abc",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Status: api.ResourceQuotaStatus{
			Hard: api.ResourceList{
				api.ResourceCPU:                    resource.MustParse("100"),
				api.ResourceMemory:                 resource.MustParse("10000"),
				api.ResourcePods:                   resource.MustParse("10"),
				api.ResourceServices:               resource.MustParse("10"),
				api.ResourceReplicationControllers: resource.MustParse("10"),
				api.ResourceQuotas:                 resource.MustParse("10"),
			},
		},
	}
	c := &simple.Client{
		Request: simple.Request{
			Method: "PUT",
			Path:   testapi.Default.ResourcePath(getResourceQuotasResoureName(), ns, "abc") + "/status",
			Query:  simple.BuildQueryValues(nil)},
		Response: simple.Response{StatusCode: 200, Body: resourceQuota},
	}
	response, err := c.Setup(t).ResourceQuotas(ns).UpdateStatus(resourceQuota)
	defer c.Close()
	c.Validate(t, response, err)
}

func TestResourceQuotaDelete(t *testing.T) {
	ns := api.NamespaceDefault
	c := &simple.Client{
		Request:  simple.Request{Method: "DELETE", Path: testapi.Default.ResourcePath(getResourceQuotasResoureName(), ns, "foo"), Query: simple.BuildQueryValues(nil)},
		Response: simple.Response{StatusCode: 200},
	}
	err := c.Setup(t).ResourceQuotas(ns).Delete("foo")
	defer c.Close()
	c.Validate(t, nil, err)
}

func TestResourceQuotaWatch(t *testing.T) {
	c := &simple.Client{
		Request: simple.Request{
			Method: "GET",
			Path:   testapi.Default.ResourcePathWithPrefix("watch", getResourceQuotasResoureName(), "", ""),
			Query:  url.Values{"resourceVersion": []string{}}},
		Response: simple.Response{StatusCode: 200},
	}
	_, err := c.Setup(t).ResourceQuotas(api.NamespaceAll).Watch(api.ListOptions{})
	defer c.Close()
	c.Validate(t, nil, err)
}
