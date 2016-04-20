/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package quota

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/util/sets"
)

// Equals returns true if the two lists are equivalent
func Equals(a api.ResourceList, b api.ResourceList) bool {
	for key, value1 := range a {
		value2, found := b[key]
		if !found {
			return false
		}
		if value1.Cmp(value2) != 0 {
			return false
		}
	}
	for key, value1 := range b {
		value2, found := a[key]
		if !found {
			return false
		}
		if value1.Cmp(value2) != 0 {
			return false
		}
	}
	return true
}

// LessThanOrEqual returns true if a < b for each key in b
// If false, it returns the keys in a that exceeded b
func LessThanOrEqual(a api.ResourceList, b api.ResourceList) (bool, []api.ResourceName) {
	result := true
	resourceNames := []api.ResourceName{}
	for key, value := range b {
		if other, found := a[key]; found {
			if other.Cmp(value) > 0 {
				result = false
				resourceNames = append(resourceNames, key)
			}
		}
	}
	return result, resourceNames
}

// Add returns the result of a + b for each named resource
func Add(a api.ResourceList, b api.ResourceList) api.ResourceList {
	result := api.ResourceList{}
	for key, value := range a {
		quantity := *value.Copy()
		if other, found := b[key]; found {
			quantity.Add(other)
		}
		result[key] = quantity
	}
	for key, value := range b {
		if _, found := result[key]; !found {
			quantity := *value.Copy()
			result[key] = quantity
		}
	}
	return result
}

// Subtract returns the result of a - b for each named resource
func Subtract(a api.ResourceList, b api.ResourceList) api.ResourceList {
	result := api.ResourceList{}
	for key, value := range a {
		quantity := *value.Copy()
		if other, found := b[key]; found {
			quantity.Sub(other)
		}
		result[key] = quantity
	}
	for key, value := range b {
		if _, found := result[key]; !found {
			quantity := *value.Copy()
			quantity.Neg(value)
			result[key] = quantity
		}
	}
	return result
}

// Mask returns a new resource list that only has the values with the specified names
func Mask(resources api.ResourceList, names []api.ResourceName) api.ResourceList {
	nameSet := ToSet(names)
	result := api.ResourceList{}
	for key, value := range resources {
		if nameSet.Has(string(key)) {
			result[key] = *value.Copy()
		}
	}
	return result
}

// ResourceNames returns a list of all resource names in the ResourceList
func ResourceNames(resources api.ResourceList) []api.ResourceName {
	result := []api.ResourceName{}
	for resourceName := range resources {
		result = append(result, resourceName)
	}
	return result
}

// Contains returns true if the specified item is in the list of items
func Contains(items []api.ResourceName, item api.ResourceName) bool {
	return ToSet(items).Has(string(item))
}

// Intersection returns the intersection of both list of resources
func Intersection(a []api.ResourceName, b []api.ResourceName) []api.ResourceName {
	setA := ToSet(a)
	setB := ToSet(b)
	setC := setA.Intersection(setB)
	result := []api.ResourceName{}
	for _, resourceName := range setC.List() {
		result = append(result, api.ResourceName(resourceName))
	}
	return result
}

// IsZero returns true if each key maps to the quantity value 0
func IsZero(a api.ResourceList) bool {
	zero := resource.MustParse("0")
	for _, v := range a {
		if v.Cmp(zero) != 0 {
			return false
		}
	}
	return true
}

// ToSet takes a list of resource names and converts to a string set
func ToSet(resourceNames []api.ResourceName) sets.String {
	result := sets.NewString()
	for _, resourceName := range resourceNames {
		result.Insert(string(resourceName))
	}
	return result
}
