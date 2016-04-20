// +build proto

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

package protobuf_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	_ "k8s.io/kubernetes/pkg/api/install"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/runtime/serializer/protobuf"
	"k8s.io/kubernetes/pkg/util/diff"
)

type testObject struct {
	gvk *unversioned.GroupVersionKind
}

func (d *testObject) GetObjectKind() unversioned.ObjectKind                 { return d }
func (d *testObject) SetGroupVersionKind(gvk *unversioned.GroupVersionKind) { d.gvk = gvk }
func (d *testObject) GroupVersionKind() *unversioned.GroupVersionKind       { return d.gvk }

type testMarshalable struct {
	testObject
	data []byte
	err  error
}

func (d *testMarshalable) Marshal() ([]byte, error) {
	return d.data, d.err
}

type testBufferedMarshalable struct {
	testObject
	data []byte
	err  error
}

func (d *testBufferedMarshalable) Marshal() ([]byte, error) {
	return nil, fmt.Errorf("not invokable")
}

func (d *testBufferedMarshalable) MarshalTo(data []byte) (int, error) {
	copy(data, d.data)
	return len(d.data), d.err
}

func (d *testBufferedMarshalable) Size() int {
	return len(d.data)
}

func TestRecognize(t *testing.T) {
	s := protobuf.NewSerializer(nil, nil, "application/protobuf")
	ignores := [][]byte{
		nil,
		{},
		[]byte("k8s"),
		{0x6b, 0x38, 0x73, 0x01},
	}
	for i, data := range ignores {
		if ok, err := s.RecognizesData(bytes.NewBuffer(data)); err != nil || ok {
			t.Errorf("%d: should not recognize data: %v", i, err)
		}
	}
	recognizes := [][]byte{
		{0x6b, 0x38, 0x73, 0x00},
		{0x6b, 0x38, 0x73, 0x00, 0x01},
	}
	for i, data := range recognizes {
		if ok, err := s.RecognizesData(bytes.NewBuffer(data)); err != nil || !ok {
			t.Errorf("%d: should recognize data: %v", i, err)
		}
	}
}

func TestEncode(t *testing.T) {
	obj1 := &testMarshalable{testObject: testObject{}, data: []byte{}}
	wire1 := []byte{
		0x6b, 0x38, 0x73, 0x00, // prefix
		0x0a, 0x04,
		0x0a, 0x00, // apiversion
		0x12, 0x00, // kind
		0x12, 0x00, // data
		0x1a, 0x00, // content-type
		0x22, 0x00, // content-encoding
	}
	obj2 := &testMarshalable{
		testObject: testObject{gvk: &unversioned.GroupVersionKind{Kind: "test", Group: "other", Version: "version"}},
		data:       []byte{0x01, 0x02, 0x03},
	}
	wire2 := []byte{
		0x6b, 0x38, 0x73, 0x00, // prefix
		0x0a, 0x15,
		0x0a, 0x0d, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x2f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, // apiversion
		0x12, 0x04, 0x74, 0x65, 0x73, 0x74, // kind
		0x12, 0x03, 0x01, 0x02, 0x03, // data
		0x1a, 0x00, // content-type
		0x22, 0x00, // content-encoding
	}

	err1 := fmt.Errorf("a test error")

	testCases := []struct {
		obj   runtime.Object
		data  []byte
		errFn func(error) bool
	}{
		{
			obj:   &testObject{},
			errFn: protobuf.IsNotMarshalable,
		},
		{
			obj:  obj1,
			data: wire1,
		},
		{
			obj:   &testMarshalable{testObject: obj1.testObject, err: err1},
			errFn: func(err error) bool { return err == err1 },
		},
		{
			// if this test fails, writing the "fast path" marshal is not the same as the "slow path"
			obj:  &testBufferedMarshalable{testObject: obj1.testObject, data: obj1.data},
			data: wire1,
		},
		{
			obj:  obj2,
			data: wire2,
		},
		{
			// if this test fails, writing the "fast path" marshal is not the same as the "slow path"
			obj:  &testBufferedMarshalable{testObject: obj2.testObject, data: obj2.data},
			data: wire2,
		},
		{
			obj:   &testBufferedMarshalable{testObject: obj1.testObject, err: err1},
			errFn: func(err error) bool { return err == err1 },
		},
	}

	for i, test := range testCases {
		s := protobuf.NewSerializer(nil, nil, "application/protobuf")
		data, err := runtime.Encode(s, test.obj)

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
			if data != nil {
				t.Errorf("%d: should not have returned nil data", i)
			}
			continue
		}

		if test.data != nil && !bytes.Equal(test.data, data) {
			t.Errorf("%d: unexpected data:\n%s", i, hex.Dump(data))
			continue
		}

		if ok, err := s.RecognizesData(bytes.NewBuffer(data)); !ok || err != nil {
			t.Errorf("%d: did not recognize data generated by call: %v", i, err)
		}
	}
}

func TestDecode(t *testing.T) {
	wire1 := []byte{
		0x6b, 0x38, 0x73, 0x00, // prefix
		0x0a, 0x04,
		0x0a, 0x00, // apiversion
		0x12, 0x00, // kind
		0x12, 0x00, // data
		0x1a, 0x00, // content-type
		0x22, 0x00, // content-encoding
	}
	wire2 := []byte{
		0x6b, 0x38, 0x73, 0x00, // prefix
		0x0a, 0x15,
		0x0a, 0x0d, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x2f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, // apiversion
		0x12, 0x04, 0x74, 0x65, 0x73, 0x74, // kind
		0x12, 0x03, 0x01, 0x02, 0x03, // data
		0x1a, 0x00, // content-type
		0x22, 0x00, // content-encoding
	}

	//err1 := fmt.Errorf("a test error")

	testCases := []struct {
		obj   runtime.Object
		data  []byte
		errFn func(error) bool
	}{
		{
			obj:   &runtime.Unknown{},
			errFn: func(err error) bool { return err.Error() == "empty data" },
		},
		{
			data:  []byte{0x6b},
			errFn: func(err error) bool { return strings.Contains(err.Error(), "does not appear to be a protobuf message") },
		},
		{
			obj: &runtime.Unknown{
				ContentType: "application/protobuf",
				Raw:         []byte{},
			},
			data: wire1,
		},
		{
			obj: &runtime.Unknown{
				TypeMeta: runtime.TypeMeta{
					APIVersion: "other/version",
					Kind:       "test",
				},
				ContentType: "application/protobuf",
				Raw:         []byte{0x01, 0x02, 0x03},
			},
			data: wire2,
		},
	}

	for i, test := range testCases {
		s := protobuf.NewSerializer(nil, nil, "application/protobuf")
		unk := &runtime.Unknown{}
		err := runtime.DecodeInto(s, test.data, unk)

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
			continue
		}

		if !reflect.DeepEqual(unk, test.obj) {
			t.Errorf("%d: unexpected object:\n%#v", i, unk)
			continue
		}
	}
}

func TestDecodeObjects(t *testing.T) {
	obj1 := &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: "cool",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "test",
				},
			},
		},
	}
	obj1wire, err := obj1.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	wire1, err := (&runtime.Unknown{
		TypeMeta: runtime.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		Raw:      obj1wire,
	}).Marshal()
	if err != nil {
		t.Fatal(err)
	}

	wire1 = append([]byte{0x6b, 0x38, 0x73, 0x00}, wire1...)

	testCases := []struct {
		obj   runtime.Object
		data  []byte
		errFn func(error) bool
	}{
		{
			obj:  obj1,
			data: wire1,
		},
	}

	for i, test := range testCases {
		s := protobuf.NewSerializer(api.Scheme, runtime.ObjectTyperToTyper(api.Scheme), "application/protobuf")
		obj, err := runtime.Decode(s, test.data)

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
				t.Errorf("%d: should not have returned an object", i)
			}
			continue
		}

		if !api.Semantic.DeepEqual(obj, test.obj) {
			t.Errorf("%d: unexpected object:\n%s", i, diff.ObjectGoPrintDiff(test.obj, obj))
			continue
		}
	}
}
