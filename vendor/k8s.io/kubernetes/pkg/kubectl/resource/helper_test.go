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

package resource

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	apitesting "k8s.io/kubernetes/pkg/api/testing"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/fake"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
)

func objBody(obj runtime.Object) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(runtime.EncodeOrDie(testapi.Default.Codec(), obj))))
}

// splitPath returns the segments for a URL path.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

func TestHelperDelete(t *testing.T) {
	tests := []struct {
		Err     bool
		Req     func(*http.Request) bool
		Resp    *http.Response
		HttpErr error
	}{
		{
			HttpErr: errors.New("failure"),
			Err:     true,
		},
		{
			Resp: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       objBody(&unversioned.Status{Status: unversioned.StatusFailure}),
			},
			Err: true,
		},
		{
			Resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       objBody(&unversioned.Status{Status: unversioned.StatusSuccess}),
			},
			Req: func(req *http.Request) bool {
				if req.Method != "DELETE" {
					t.Errorf("unexpected method: %#v", req)
					return false
				}
				parts := splitPath(req.URL.Path)
				if len(parts) < 3 {
					t.Errorf("expected URL path to have 3 parts: %s", req.URL.Path)
					return false
				}
				if parts[1] != "bar" {
					t.Errorf("url doesn't contain namespace: %#v", req)
					return false
				}
				if parts[2] != "foo" {
					t.Errorf("url doesn't contain name: %#v", req)
					return false
				}
				return true
			},
		},
	}
	for _, test := range tests {
		client := &fake.RESTClient{
			Codec: testapi.Default.Codec(),
			Resp:  test.Resp,
			Err:   test.HttpErr,
		}
		modifier := &Helper{
			RESTClient:      client,
			NamespaceScoped: true,
		}
		err := modifier.Delete("bar", "foo")
		if (err != nil) != test.Err {
			t.Errorf("unexpected error: %t %v", test.Err, err)
		}
		if err != nil {
			continue
		}
		if test.Req != nil && !test.Req(client.Req) {
			t.Errorf("unexpected request: %#v", client.Req)
		}
	}
}

func TestHelperCreate(t *testing.T) {
	expectPost := func(req *http.Request) bool {
		if req.Method != "POST" {
			t.Errorf("unexpected method: %#v", req)
			return false
		}
		parts := splitPath(req.URL.Path)
		if parts[1] != "bar" {
			t.Errorf("url doesn't contain namespace: %#v", req)
			return false
		}
		return true
	}

	tests := []struct {
		Resp    *http.Response
		HttpErr error
		Modify  bool
		Object  runtime.Object

		ExpectObject runtime.Object
		Err          bool
		Req          func(*http.Request) bool
	}{
		{
			HttpErr: errors.New("failure"),
			Err:     true,
		},
		{
			Resp: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       objBody(&unversioned.Status{Status: unversioned.StatusFailure}),
			},
			Err: true,
		},
		{
			Resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       objBody(&unversioned.Status{Status: unversioned.StatusSuccess}),
			},
			Object:       &api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo"}},
			ExpectObject: &api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo"}},
			Req:          expectPost,
		},
		{
			Modify:       false,
			Object:       &api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "10"}},
			ExpectObject: &api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "10"}},
			Resp:         &http.Response{StatusCode: http.StatusOK, Body: objBody(&unversioned.Status{Status: unversioned.StatusSuccess})},
			Req:          expectPost,
		},
		{
			Modify: true,
			Object: &api.Pod{
				ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "10"},
				Spec:       apitesting.DeepEqualSafePodSpec(),
			},
			ExpectObject: &api.Pod{
				ObjectMeta: api.ObjectMeta{Name: "foo"},
				Spec:       apitesting.DeepEqualSafePodSpec(),
			},
			Resp: &http.Response{StatusCode: http.StatusOK, Body: objBody(&unversioned.Status{Status: unversioned.StatusSuccess})},
			Req:  expectPost,
		},
	}
	for i, test := range tests {
		client := &fake.RESTClient{
			Codec: testapi.Default.Codec(),
			Resp:  test.Resp,
			Err:   test.HttpErr,
		}
		modifier := &Helper{
			RESTClient:      client,
			Versioner:       testapi.Default.MetadataAccessor(),
			NamespaceScoped: true,
		}
		_, err := modifier.Create("bar", test.Modify, test.Object)
		if (err != nil) != test.Err {
			t.Errorf("%d: unexpected error: %t %v", i, test.Err, err)
		}
		if err != nil {
			continue
		}
		if test.Req != nil && !test.Req(client.Req) {
			t.Errorf("%d: unexpected request: %#v", i, client.Req)
		}
		body, err := ioutil.ReadAll(client.Req.Body)
		if err != nil {
			t.Fatalf("%d: unexpected error: %#v", i, err)
		}
		t.Logf("got body: %s", string(body))
		expect := []byte{}
		if test.ExpectObject != nil {
			expect = []byte(runtime.EncodeOrDie(testapi.Default.Codec(), test.ExpectObject))
		}
		if !reflect.DeepEqual(expect, body) {
			t.Errorf("%d: unexpected body: %s (expected %s)", i, string(body), string(expect))
		}

	}
}

func TestHelperGet(t *testing.T) {
	tests := []struct {
		Err     bool
		Req     func(*http.Request) bool
		Resp    *http.Response
		HttpErr error
	}{
		{
			HttpErr: errors.New("failure"),
			Err:     true,
		},
		{
			Resp: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       objBody(&unversioned.Status{Status: unversioned.StatusFailure}),
			},
			Err: true,
		},
		{
			Resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       objBody(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo"}}),
			},
			Req: func(req *http.Request) bool {
				if req.Method != "GET" {
					t.Errorf("unexpected method: %#v", req)
					return false
				}
				parts := splitPath(req.URL.Path)
				if parts[1] != "bar" {
					t.Errorf("url doesn't contain namespace: %#v", req)
					return false
				}
				if parts[2] != "foo" {
					t.Errorf("url doesn't contain name: %#v", req)
					return false
				}
				return true
			},
		},
	}
	for _, test := range tests {
		client := &fake.RESTClient{
			Codec: testapi.Default.Codec(),
			Resp:  test.Resp,
			Err:   test.HttpErr,
		}
		modifier := &Helper{
			RESTClient:      client,
			NamespaceScoped: true,
		}
		obj, err := modifier.Get("bar", "foo", false)
		if (err != nil) != test.Err {
			t.Errorf("unexpected error: %t %v", test.Err, err)
		}
		if err != nil {
			continue
		}
		if obj.(*api.Pod).Name != "foo" {
			t.Errorf("unexpected object: %#v", obj)
		}
		if test.Req != nil && !test.Req(client.Req) {
			t.Errorf("unexpected request: %#v", client.Req)
		}
	}
}

func TestHelperList(t *testing.T) {
	tests := []struct {
		Err     bool
		Req     func(*http.Request) bool
		Resp    *http.Response
		HttpErr error
	}{
		{
			HttpErr: errors.New("failure"),
			Err:     true,
		},
		{
			Resp: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       objBody(&unversioned.Status{Status: unversioned.StatusFailure}),
			},
			Err: true,
		},
		{
			Resp: &http.Response{
				StatusCode: http.StatusOK,
				Body: objBody(&api.PodList{
					Items: []api.Pod{{
						ObjectMeta: api.ObjectMeta{Name: "foo"},
					},
					},
				}),
			},
			Req: func(req *http.Request) bool {
				if req.Method != "GET" {
					t.Errorf("unexpected method: %#v", req)
					return false
				}
				if req.URL.Path != "/namespaces/bar" {
					t.Errorf("url doesn't contain name: %#v", req.URL)
					return false
				}
				if req.URL.Query().Get(unversioned.LabelSelectorQueryParam(testapi.Default.GroupVersion().String())) != labels.SelectorFromSet(labels.Set{"foo": "baz"}).String() {
					t.Errorf("url doesn't contain query parameters: %#v", req.URL)
					return false
				}
				return true
			},
		},
	}
	for _, test := range tests {
		client := &fake.RESTClient{
			Codec: testapi.Default.Codec(),
			Resp:  test.Resp,
			Err:   test.HttpErr,
		}
		modifier := &Helper{
			RESTClient:      client,
			NamespaceScoped: true,
		}
		obj, err := modifier.List("bar", testapi.Default.GroupVersion().String(), labels.SelectorFromSet(labels.Set{"foo": "baz"}), false)
		if (err != nil) != test.Err {
			t.Errorf("unexpected error: %t %v", test.Err, err)
		}
		if err != nil {
			continue
		}
		if obj.(*api.PodList).Items[0].Name != "foo" {
			t.Errorf("unexpected object: %#v", obj)
		}
		if test.Req != nil && !test.Req(client.Req) {
			t.Errorf("unexpected request: %#v", client.Req)
		}
	}
}

func TestHelperReplace(t *testing.T) {
	expectPut := func(path string, req *http.Request) bool {
		if req.Method != "PUT" {
			t.Errorf("unexpected method: %#v", req)
			return false
		}
		if req.URL.Path != path {
			t.Errorf("unexpected url: %v", req.URL)
			return false
		}
		return true
	}

	tests := []struct {
		Resp            *http.Response
		HTTPClient      *http.Client
		HttpErr         error
		Overwrite       bool
		Object          runtime.Object
		Namespace       string
		NamespaceScoped bool

		ExpectPath   string
		ExpectObject runtime.Object
		Err          bool
		Req          func(string, *http.Request) bool
	}{
		{
			Namespace:       "bar",
			NamespaceScoped: true,
			HttpErr:         errors.New("failure"),
			Err:             true,
		},
		{
			Namespace:       "bar",
			NamespaceScoped: true,
			Object:          &api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo"}},
			Resp: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       objBody(&unversioned.Status{Status: unversioned.StatusFailure}),
			},
			Err: true,
		},
		{
			Namespace:       "bar",
			NamespaceScoped: true,
			Object:          &api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo"}},
			ExpectPath:      "/namespaces/bar/foo",
			ExpectObject:    &api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo"}},
			Resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       objBody(&unversioned.Status{Status: unversioned.StatusSuccess}),
			},
			Req: expectPut,
		},
		// namespace scoped resource
		{
			Namespace:       "bar",
			NamespaceScoped: true,
			Object: &api.Pod{
				ObjectMeta: api.ObjectMeta{Name: "foo"},
				Spec:       apitesting.DeepEqualSafePodSpec(),
			},
			ExpectPath: "/namespaces/bar/foo",
			ExpectObject: &api.Pod{
				ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "10"},
				Spec:       apitesting.DeepEqualSafePodSpec(),
			},
			Overwrite: true,
			HTTPClient: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
				if req.Method == "PUT" {
					return &http.Response{StatusCode: http.StatusOK, Body: objBody(&unversioned.Status{Status: unversioned.StatusSuccess})}, nil
				}
				return &http.Response{StatusCode: http.StatusOK, Body: objBody(&api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "10"}})}, nil
			}),
			Req: expectPut,
		},
		// cluster scoped resource
		{
			Object: &api.Node{
				ObjectMeta: api.ObjectMeta{Name: "foo"},
			},
			ExpectObject: &api.Node{
				ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "10"},
			},
			Overwrite:  true,
			ExpectPath: "/foo",
			HTTPClient: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
				if req.Method == "PUT" {
					return &http.Response{StatusCode: http.StatusOK, Body: objBody(&unversioned.Status{Status: unversioned.StatusSuccess})}, nil
				}
				return &http.Response{StatusCode: http.StatusOK, Body: objBody(&api.Node{ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "10"}})}, nil
			}),
			Req: expectPut,
		},
		{
			Namespace:       "bar",
			NamespaceScoped: true,
			Object:          &api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "10"}},
			ExpectPath:      "/namespaces/bar/foo",
			ExpectObject:    &api.Pod{ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "10"}},
			Resp:            &http.Response{StatusCode: http.StatusOK, Body: objBody(&unversioned.Status{Status: unversioned.StatusSuccess})},
			Req:             expectPut,
		},
	}
	for i, test := range tests {
		client := &fake.RESTClient{
			Client: test.HTTPClient,
			Codec:  testapi.Default.Codec(),
			Resp:   test.Resp,
			Err:    test.HttpErr,
		}
		modifier := &Helper{
			RESTClient:      client,
			Versioner:       testapi.Default.MetadataAccessor(),
			NamespaceScoped: test.NamespaceScoped,
		}
		_, err := modifier.Replace(test.Namespace, "foo", test.Overwrite, test.Object)
		if (err != nil) != test.Err {
			t.Errorf("%d: unexpected error: %t %v", i, test.Err, err)
		}
		if err != nil {
			continue
		}
		if test.Req != nil && !test.Req(test.ExpectPath, client.Req) {
			t.Errorf("%d: unexpected request: %#v", i, client.Req)
		}
		body, err := ioutil.ReadAll(client.Req.Body)
		if err != nil {
			t.Fatalf("%d: unexpected error: %#v", i, err)
		}
		expect := []byte{}
		if test.ExpectObject != nil {
			expect = []byte(runtime.EncodeOrDie(testapi.Default.Codec(), test.ExpectObject))
		}
		if !reflect.DeepEqual(expect, body) {
			t.Errorf("%d: unexpected body: %s", i, string(body))
		}
	}
}
