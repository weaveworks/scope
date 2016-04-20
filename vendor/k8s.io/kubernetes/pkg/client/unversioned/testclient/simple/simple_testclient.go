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

package simple

import (
	"net/http/httptest"
	"net/url"
	"path"
	"reflect"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/api/unversioned"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"
	utiltesting "k8s.io/kubernetes/pkg/util/testing"
)

const NameRequiredError = "resource name may not be empty"

type Request struct {
	Method  string
	Path    string
	Header  string
	Query   url.Values
	Body    runtime.Object
	RawBody *string
}

type Response struct {
	StatusCode int
	Body       runtime.Object
	RawBody    *string
}

type Client struct {
	*client.Client
	Clientset *clientset.Clientset
	Request   Request
	Response  Response
	Error     bool
	Created   bool
	server    *httptest.Server
	handler   *utiltesting.FakeHandler
	// For query args, an optional function to validate the contents
	// useful when the contents can change but still be correct.
	// Maps from query arg key to validator.
	// If no validator is present, string equality is used.
	QueryValidator map[string]func(string, string) bool

	// If your object could exist in multiple groups, set this to
	// correspond to the URL you're testing it with.
	ResourceGroup string
}

func (c *Client) Setup(t *testing.T) *Client {
	c.handler = &utiltesting.FakeHandler{
		StatusCode: c.Response.StatusCode,
	}
	if responseBody := c.body(t, c.Response.Body, c.Response.RawBody); responseBody != nil {
		c.handler.ResponseBody = *responseBody
	}
	c.server = httptest.NewServer(c.handler)
	if c.Client == nil {
		c.Client = client.NewOrDie(&restclient.Config{
			Host:          c.server.URL,
			ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Default.GroupVersion()},
		})

		// TODO: caesarxuchao: hacky way to specify version of Experimental client.
		// We will fix this by supporting multiple group versions in Config
		c.AutoscalingClient = client.NewAutoscalingOrDie(&restclient.Config{
			Host:          c.server.URL,
			ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Autoscaling.GroupVersion()},
		})
		c.BatchClient = client.NewBatchOrDie(&restclient.Config{
			Host:          c.server.URL,
			ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Batch.GroupVersion()},
		})
		c.ExtensionsClient = client.NewExtensionsOrDie(&restclient.Config{
			Host:          c.server.URL,
			ContentConfig: restclient.ContentConfig{GroupVersion: testapi.Extensions.GroupVersion()},
		})

		c.Clientset = clientset.NewForConfigOrDie(&restclient.Config{Host: c.server.URL})
	}
	c.QueryValidator = map[string]func(string, string) bool{}
	return c
}

func (c *Client) Close() {
	if c.server != nil {
		// TODO: Uncomment when fix #19254
		// c.server.Close()
	}
}

func (c *Client) ServerURL() string {
	return c.server.URL
}

func (c *Client) Validate(t *testing.T, received runtime.Object, err error) {
	c.ValidateCommon(t, err)

	if c.Response.Body != nil && !api.Semantic.DeepDerivative(c.Response.Body, received) {
		t.Errorf("bad response for request %#v: \nexpected %#v\ngot %#v\n", c.Request, c.Response.Body, received)
	}
}

func (c *Client) ValidateRaw(t *testing.T, received []byte, err error) {
	c.ValidateCommon(t, err)

	if c.Response.Body != nil && !reflect.DeepEqual(c.Response.Body, received) {
		t.Errorf("bad response for request %#v: expected %#v, got %#v", c.Request, c.Response.Body, received)
	}
}

func (c *Client) ValidateCommon(t *testing.T, err error) {
	if c.Error {
		if err == nil {
			t.Errorf("error expected for %#v, got none", c.Request)
		}
		return
	}
	if err != nil {
		t.Errorf("no error expected for %#v, got: %v", c.Request, err)
	}

	if c.handler.RequestReceived == nil {
		t.Errorf("handler had an empty request, %#v", c)
		return
	}

	requestBody := c.body(t, c.Request.Body, c.Request.RawBody)
	actualQuery := c.handler.RequestReceived.URL.Query()
	t.Logf("got query: %v", actualQuery)
	t.Logf("path: %v", c.Request.Path)
	// We check the query manually, so blank it out so that FakeHandler.ValidateRequest
	// won't check it.
	c.handler.RequestReceived.URL.RawQuery = ""
	c.handler.ValidateRequest(t, path.Join(c.Request.Path), c.Request.Method, requestBody)
	for key, values := range c.Request.Query {
		validator, ok := c.QueryValidator[key]
		if !ok {
			switch key {
			case unversioned.LabelSelectorQueryParam(testapi.Default.GroupVersion().String()):
				validator = ValidateLabels
			case unversioned.FieldSelectorQueryParam(testapi.Default.GroupVersion().String()):
				validator = validateFields
			default:
				validator = func(a, b string) bool { return a == b }
			}
		}
		observed := actualQuery.Get(key)
		wanted := strings.Join(values, "")
		if !validator(wanted, observed) {
			t.Errorf("Unexpected query arg for key: %s.  Expected %s, Received %s", key, wanted, observed)
		}
	}
	if c.Request.Header != "" {
		if c.handler.RequestReceived.Header.Get(c.Request.Header) == "" {
			t.Errorf("header %q not found in request %#v", c.Request.Header, c.handler.RequestReceived)
		}
	}

	if expected, received := requestBody, c.handler.RequestBody; expected != nil && *expected != received {
		t.Errorf("bad body for request %#v: expected %s, got %s", c.Request, *expected, received)
	}
}

// buildQueryValues is a convenience function for knowing if a namespace should be in a query param or not
func BuildQueryValues(query url.Values) url.Values {
	v := url.Values{}
	if query != nil {
		for key, values := range query {
			for _, value := range values {
				v.Add(key, value)
			}
		}
	}
	return v
}

func ValidateLabels(a, b string) bool {
	sA, eA := labels.Parse(a)
	if eA != nil {
		return false
	}
	sB, eB := labels.Parse(b)
	if eB != nil {
		return false
	}
	return sA.String() == sB.String()
}

func validateFields(a, b string) bool {
	sA, _ := fields.ParseSelector(a)
	sB, _ := fields.ParseSelector(b)
	return sA.String() == sB.String()
}

func (c *Client) body(t *testing.T, obj runtime.Object, raw *string) *string {
	if obj != nil {
		fqKind, err := api.Scheme.ObjectKind(obj)
		if err != nil {
			t.Errorf("unexpected encoding error: %v", err)
		}
		groupName := fqKind.GroupVersion().Group
		if c.ResourceGroup != "" {
			groupName = c.ResourceGroup
		}
		var bs []byte
		g, found := testapi.Groups[groupName]
		if !found {
			t.Errorf("Group %s is not registered in testapi", groupName)
		}
		bs, err = runtime.Encode(g.Codec(), obj)
		if err != nil {
			t.Errorf("unexpected encoding error: %v", err)
		}
		body := string(bs)
		return &body
	}
	return raw
}
