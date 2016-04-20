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

package api_test

import (
	"math/rand"
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	apitesting "k8s.io/kubernetes/pkg/api/testing"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/util/diff"

	"github.com/google/gofuzz"
)

func TestDeepCopyApiObjects(t *testing.T) {
	for i := 0; i < *fuzzIters; i++ {
		for _, version := range []unversioned.GroupVersion{testapi.Default.InternalGroupVersion(), *testapi.Default.GroupVersion()} {
			f := apitesting.FuzzerFor(t, version, rand.NewSource(rand.Int63()))
			for kind := range api.Scheme.KnownTypes(version) {
				doDeepCopyTest(t, version.WithKind(kind), f)
			}
		}
	}
}

func doDeepCopyTest(t *testing.T, kind unversioned.GroupVersionKind, f *fuzz.Fuzzer) {
	item, err := api.Scheme.New(kind)
	if err != nil {
		t.Fatalf("Could not create a %v: %s", kind, err)
	}
	f.Fuzz(item)
	itemCopy, err := api.Scheme.DeepCopy(item)
	if err != nil {
		t.Errorf("Could not deep copy a %v: %s", kind, err)
		return
	}

	if !reflect.DeepEqual(item, itemCopy) {
		t.Errorf("\nexpected: %#v\n\ngot:      %#v\n\ndiff:      %v", item, itemCopy, diff.ObjectGoPrintSideBySide(item, itemCopy))
	}
}

func TestDeepCopySingleType(t *testing.T) {
	for i := 0; i < *fuzzIters; i++ {
		for _, version := range []unversioned.GroupVersion{testapi.Default.InternalGroupVersion(), *testapi.Default.GroupVersion()} {
			f := apitesting.FuzzerFor(t, version, rand.NewSource(rand.Int63()))
			doDeepCopyTest(t, version.WithKind("Pod"), f)
		}
	}
}
