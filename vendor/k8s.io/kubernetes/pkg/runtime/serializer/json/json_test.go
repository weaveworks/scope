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

package json_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/runtime/serializer/json"
	"k8s.io/kubernetes/pkg/util/diff"
)

type testDecodable struct {
	Other string
	Value int `json:"value"`
	gvk   *unversioned.GroupVersionKind
}

func (d *testDecodable) GetObjectKind() unversioned.ObjectKind                 { return d }
func (d *testDecodable) SetGroupVersionKind(gvk *unversioned.GroupVersionKind) { d.gvk = gvk }
func (d *testDecodable) GroupVersionKind() *unversioned.GroupVersionKind       { return d.gvk }

func TestDecode(t *testing.T) {
	testCases := []struct {
		creater runtime.ObjectCreater
		typer   runtime.Typer
		yaml    bool
		pretty  bool

		data       []byte
		defaultGVK *unversioned.GroupVersionKind
		into       runtime.Object

		errFn          func(error) bool
		expectedObject runtime.Object
		expectedGVK    *unversioned.GroupVersionKind
	}{
		{
			data: []byte("{}"),

			expectedGVK: &unversioned.GroupVersionKind{},
			errFn:       func(err error) bool { return strings.Contains(err.Error(), "Object 'Kind' is missing in") },
		},
		{
			data:       []byte("{}"),
			defaultGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			creater:    &mockCreater{err: fmt.Errorf("fake error")},

			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			errFn:       func(err error) bool { return err.Error() == "fake error" },
		},
		{
			data:       []byte("{}"),
			defaultGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			creater:    &mockCreater{err: fmt.Errorf("fake error")},

			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			errFn:       func(err error) bool { return err.Error() == "fake error" },
		},
		{
			data:       []byte("{}"),
			defaultGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			creater:    &mockCreater{obj: &testDecodable{}},
			expectedObject: &testDecodable{
				gvk: nil, // json serializer does NOT set GVK
			},
			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
		},

		// version without group is not defaulted
		{
			data:       []byte(`{"apiVersion":"blah"}`),
			defaultGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			creater:    &mockCreater{obj: &testDecodable{}},
			expectedObject: &testDecodable{
				gvk: nil, // json serializer does NOT set GVK
			},
			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "", Version: "blah"},
		},
		// group without version is defaulted
		{
			data:       []byte(`{"apiVersion":"other/"}`),
			defaultGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			creater:    &mockCreater{obj: &testDecodable{}},
			expectedObject: &testDecodable{
				gvk: nil, // json serializer does NOT set GVK
			},
			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
		},

		// accept runtime.Unknown as into and bypass creator
		{
			data: []byte(`{}`),
			into: &runtime.Unknown{},

			expectedGVK: &unversioned.GroupVersionKind{},
			expectedObject: &runtime.Unknown{
				Raw:         []byte(`{}`),
				ContentType: runtime.ContentTypeJSON,
			},
		},
		{
			data: []byte(`{"test":"object"}`),
			into: &runtime.Unknown{},

			expectedGVK: &unversioned.GroupVersionKind{},
			expectedObject: &runtime.Unknown{
				Raw:         []byte(`{"test":"object"}`),
				ContentType: runtime.ContentTypeJSON,
			},
		},
		{
			data:        []byte(`{"test":"object"}`),
			into:        &runtime.Unknown{},
			defaultGVK:  &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			expectedObject: &runtime.Unknown{
				TypeMeta:    runtime.TypeMeta{APIVersion: "other/blah", Kind: "Test"},
				Raw:         []byte(`{"test":"object"}`),
				ContentType: runtime.ContentTypeJSON,
			},
		},

		// unregistered objects can be decoded into directly
		{
			data:        []byte(`{"kind":"Test","apiVersion":"other/blah","value":1,"Other":"test"}`),
			into:        &testDecodable{},
			typer:       &mockTyper{err: runtime.NewNotRegisteredErr(unversioned.GroupVersionKind{}, nil)},
			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			expectedObject: &testDecodable{
				Other: "test",
				Value: 1,
			},
		},
		// registered types get defaulted by the into object kind
		{
			data:        []byte(`{"value":1,"Other":"test"}`),
			into:        &testDecodable{},
			typer:       &mockTyper{gvk: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"}},
			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			expectedObject: &testDecodable{
				Other: "test",
				Value: 1,
			},
		},
		// registered types get defaulted by the into object kind even without version, but return an error
		{
			data:        []byte(`{"value":1,"Other":"test"}`),
			into:        &testDecodable{},
			typer:       &mockTyper{gvk: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: ""}},
			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: ""},
			errFn:       func(err error) bool { return strings.Contains(err.Error(), "Object 'apiVersion' is missing in") },
			expectedObject: &testDecodable{
				Other: "test",
				Value: 1,
			},
		},

		// runtime.VersionedObjects are decoded
		{
			data:        []byte(`{"value":1,"Other":"test"}`),
			into:        &runtime.VersionedObjects{Objects: []runtime.Object{}},
			creater:     &mockCreater{obj: &testDecodable{}},
			typer:       &mockTyper{gvk: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"}},
			defaultGVK:  &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			expectedObject: &runtime.VersionedObjects{
				Objects: []runtime.Object{
					&testDecodable{
						Other: "test",
						Value: 1,
					},
				},
			},
		},
		// runtime.VersionedObjects with an object are decoded into
		{
			data:        []byte(`{"Other":"test"}`),
			into:        &runtime.VersionedObjects{Objects: []runtime.Object{&testDecodable{Value: 2}}},
			typer:       &mockTyper{gvk: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"}},
			expectedGVK: &unversioned.GroupVersionKind{Kind: "Test", Group: "other", Version: "blah"},
			expectedObject: &runtime.VersionedObjects{
				Objects: []runtime.Object{
					&testDecodable{
						Other: "test",
						Value: 2,
					},
				},
			},
		},
	}

	for i, test := range testCases {
		var s runtime.Serializer
		if test.yaml {
			s = json.NewYAMLSerializer(json.DefaultMetaFactory, test.creater, test.typer)
		} else {
			s = json.NewSerializer(json.DefaultMetaFactory, test.creater, test.typer, test.pretty)
		}
		obj, gvk, err := s.Decode([]byte(test.data), test.defaultGVK, test.into)

		if !reflect.DeepEqual(test.expectedGVK, gvk) {
			t.Errorf("%d: unexpected GVK: %v", i, gvk)
		}

		switch {
		case err == nil && test.errFn != nil:
			t.Errorf("%d: failed: %v", i, err)
			continue
		case err != nil && test.errFn == nil:
			t.Errorf("%d: failed: %v", i, err)
			continue
		case err != nil:
			if !test.errFn(err) {
				t.Errorf("%d: failed: %v", i, err)
			}
			if obj != nil {
				t.Errorf("%d: should have returned nil object", i)
			}
			continue
		}

		if test.into != nil && test.into != obj {
			t.Errorf("%d: expected into to be returned: %v", i, obj)
			continue
		}

		if !reflect.DeepEqual(test.expectedObject, obj) {
			t.Errorf("%d: unexpected object:\n%s", i, diff.ObjectGoPrintSideBySide(test.expectedObject, obj))
		}
	}
}

type mockCreater struct {
	apiVersion string
	kind       string
	err        error
	obj        runtime.Object
}

func (c *mockCreater) New(kind unversioned.GroupVersionKind) (runtime.Object, error) {
	c.apiVersion, c.kind = kind.GroupVersion().String(), kind.Kind
	return c.obj, c.err
}

type mockTyper struct {
	gvk *unversioned.GroupVersionKind
	err error
}

func (t *mockTyper) ObjectKind(obj runtime.Object) (*unversioned.GroupVersionKind, bool, error) {
	return t.gvk, false, t.err
}
