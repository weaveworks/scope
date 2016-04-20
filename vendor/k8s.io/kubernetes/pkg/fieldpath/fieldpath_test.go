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

package fieldpath

import (
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
)

func TestExtractFieldPathAsString(t *testing.T) {
	cases := []struct {
		name                    string
		fieldPath               string
		obj                     interface{}
		expectedValue           string
		expectedMessageFragment string
	}{
		{
			name:      "not an API object",
			fieldPath: "metadata.name",
			obj:       "",
			expectedMessageFragment: "expected struct",
		},
		{
			name:      "ok - namespace",
			fieldPath: "metadata.namespace",
			obj: &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Namespace: "object-namespace",
				},
			},
			expectedValue: "object-namespace",
		},
		{
			name:      "ok - name",
			fieldPath: "metadata.name",
			obj: &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Name: "object-name",
				},
			},
			expectedValue: "object-name",
		},
		{
			name:      "ok - labels",
			fieldPath: "metadata.labels",
			obj: &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Labels: map[string]string{"key": "value"},
				},
			},
			expectedValue: "key=\"value\"\n",
		},
		{
			name:      "ok - labels bslash n",
			fieldPath: "metadata.labels",
			obj: &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Labels: map[string]string{"key": "value\n"},
				},
			},
			expectedValue: "key=\"value\\n\"\n",
		},
		{
			name:      "ok - annotations",
			fieldPath: "metadata.annotations",
			obj: &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Annotations: map[string]string{"builder": "john-doe"},
				},
			},
			expectedValue: "builder=\"john-doe\"\n",
		},

		{
			name:      "invalid expression",
			fieldPath: "metadata.whoops",
			obj: &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Namespace: "object-namespace",
				},
			},
			expectedMessageFragment: "Unsupported fieldPath",
		},
	}

	for _, tc := range cases {
		actual, err := ExtractFieldPathAsString(tc.obj, tc.fieldPath)
		if err != nil {
			if tc.expectedMessageFragment != "" {
				if !strings.Contains(err.Error(), tc.expectedMessageFragment) {
					t.Errorf("%v: Unexpected error message: %q, expected to contain %q", tc.name, err, tc.expectedMessageFragment)
				}
			} else {
				t.Errorf("%v: unexpected error: %v", tc.name, err)
			}
		} else if e := tc.expectedValue; e != "" && e != actual {
			t.Errorf("%v: Unexpected result; got %q, expected %q", tc.name, actual, e)
		}
	}
}
