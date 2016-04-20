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

package registered

import (
	"testing"

	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apimachinery"
)

func TestAllPreferredGroupVersions(t *testing.T) {
	testCases := []struct {
		groupMetas []apimachinery.GroupMeta
		expect     string
	}{
		{
			groupMetas: []apimachinery.GroupMeta{
				{
					GroupVersion: unversioned.GroupVersion{Group: "group1", Version: "v1"},
				},
				{
					GroupVersion: unversioned.GroupVersion{Group: "group2", Version: "v2"},
				},
				{
					GroupVersion: unversioned.GroupVersion{Group: "", Version: "v1"},
				},
			},
			expect: "group1/v1,group2/v2,v1",
		},
		{
			groupMetas: []apimachinery.GroupMeta{
				{
					GroupVersion: unversioned.GroupVersion{Group: "", Version: "v1"},
				},
			},
			expect: "v1",
		},
		{
			groupMetas: []apimachinery.GroupMeta{},
			expect:     "",
		},
	}
	for _, testCase := range testCases {
		for _, groupMeta := range testCase.groupMetas {
			RegisterGroup(groupMeta)
		}
		output := AllPreferredGroupVersions()
		if testCase.expect != output {
			t.Errorf("Error. expect: %s, got: %s", testCase.expect, output)
		}
		reset()
	}
}
