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

package genericapiserver

import (
	"testing"

	"k8s.io/kubernetes/pkg/api/unversioned"
)

func TestDisabledVersion(t *testing.T) {
	g1v1 := unversioned.GroupVersion{Group: "group1", Version: "version1"}
	g1v2 := unversioned.GroupVersion{Group: "group1", Version: "version2"}
	g2v1 := unversioned.GroupVersion{Group: "group2", Version: "version1"}
	g3v1 := unversioned.GroupVersion{Group: "group3", Version: "version1"}

	resourceType := "the-resource"
	disabledResourceType := "the-disabled-resource"

	config := NewResourceConfig()

	config.DisableVersions(g1v1)
	config.EnableVersions(g1v2, g3v1)
	config.EnableResources(g1v1.WithResource(resourceType), g2v1.WithResource(resourceType))
	config.DisableResources(g1v2.WithResource(disabledResourceType))

	expectedEnabledResources := []unversioned.GroupVersionResource{
		g1v2.WithResource(resourceType),
		g2v1.WithResource(resourceType),
	}
	expectedDisabledResources := []unversioned.GroupVersionResource{
		g1v1.WithResource(resourceType), g1v1.WithResource(disabledResourceType),
		g1v2.WithResource(disabledResourceType),
		g2v1.WithResource(disabledResourceType),
	}

	for _, expectedResource := range expectedEnabledResources {
		if !config.ResourceEnabled(expectedResource) {
			t.Errorf("expected enabled for %v, from %v", expectedResource, config)
		}
	}
	for _, expectedResource := range expectedDisabledResources {
		if config.ResourceEnabled(expectedResource) {
			t.Errorf("expected disabled for %v, from %v", expectedResource, config)
		}
	}

	if e, a := false, config.AnyResourcesForVersionEnabled(g1v1); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := false, config.AllResourcesForVersionEnabled(g1v1); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := true, config.AnyResourcesForVersionEnabled(g1v2); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := false, config.AllResourcesForVersionEnabled(g1v2); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := true, config.AnyResourcesForVersionEnabled(g3v1); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
	if e, a := true, config.AllResourcesForVersionEnabled(g3v1); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	expectedEnabledAnyVersionResources := []unversioned.GroupResource{
		{Group: "group1", Resource: resourceType},
	}
	expectedDisabledAnyResources := []unversioned.GroupResource{
		{Group: "group1", Resource: disabledResourceType},
	}

	for _, expectedResource := range expectedEnabledAnyVersionResources {
		if !config.AnyVersionOfResourceEnabled(expectedResource) {
			t.Errorf("expected enabled for %v, from %v", expectedResource, config)
		}
	}
	for _, expectedResource := range expectedDisabledAnyResources {
		if config.AnyVersionOfResourceEnabled(expectedResource) {
			t.Errorf("expected disabled for %v, from %v", expectedResource, config)
		}
	}

}
