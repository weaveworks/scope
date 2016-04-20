/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/runtime"
)

func TestNewClient(t *testing.T) {
	o := NewObjects(api.Scheme, api.Codecs.UniversalDecoder())
	if err := AddObjectsFromPath("../../../../examples/guestbook/frontend-service.yaml", o, api.Codecs.UniversalDecoder()); err != nil {
		t.Fatal(err)
	}
	client := &Fake{}
	client.AddReactor("*", "*", ObjectReaction(o, testapi.Default.RESTMapper()))
	list, err := client.Services("test").List(api.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("unexpected list %#v", list)
	}

	// When list is invoked a second time, the same results are returned.
	list, err = client.Services("test").List(api.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("unexpected list %#v", list)
	}
	t.Logf("list: %#v", list)
}

func TestErrors(t *testing.T) {
	o := NewObjects(api.Scheme, api.Codecs.UniversalDecoder())
	o.Add(&api.List{
		Items: []runtime.Object{
			// This first call to List will return this error
			&(errors.NewNotFound(api.Resource("ServiceList"), "").(*errors.StatusError).ErrStatus),
			// The second call to List will return this error
			&(errors.NewForbidden(api.Resource("ServiceList"), "", nil).(*errors.StatusError).ErrStatus),
		},
	})
	client := &Fake{}
	client.AddReactor("*", "*", ObjectReaction(o, testapi.Default.RESTMapper()))
	_, err := client.Services("test").List(api.ListOptions{})
	if !errors.IsNotFound(err) {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("error: %#v", err.(*errors.StatusError).Status())
	_, err = client.Services("test").List(api.ListOptions{})
	if !errors.IsForbidden(err) {
		t.Fatalf("unexpected error: %v", err)
	}
}
