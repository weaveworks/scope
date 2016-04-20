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

package recognizer

import (
	"testing"

	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/runtime/serializer/json"
)

type A struct{}

func (A) GetObjectKind() unversioned.ObjectKind { return unversioned.EmptyObjectKind }

func TestRecognizer(t *testing.T) {
	s := runtime.NewScheme()
	s.AddKnownTypes(unversioned.GroupVersion{Version: "v1"}, &A{})
	d := NewDecoder(
		json.NewSerializer(json.DefaultMetaFactory, s, runtime.ObjectTyperToTyper(s), false),
		json.NewYAMLSerializer(json.DefaultMetaFactory, s, runtime.ObjectTyperToTyper(s)),
	)
	out, _, err := d.Decode([]byte(`
kind: A
apiVersion: v1
`), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", out)

	out, _, err = d.Decode([]byte(`
{
  "kind":"A",
  "apiVersion":"v1"
}
`), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", out)
}
