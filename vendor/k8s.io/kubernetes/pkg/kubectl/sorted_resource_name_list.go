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

package kubectl

import (
	"k8s.io/kubernetes/pkg/api"
)

type SortableResourceNames []api.ResourceName

func (list SortableResourceNames) Len() int {
	return len(list)
}

func (list SortableResourceNames) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableResourceNames) Less(i, j int) bool {
	return list[i] < list[j]
}

type SortableResourceQuotas []api.ResourceQuota

func (list SortableResourceQuotas) Len() int {
	return len(list)
}

func (list SortableResourceQuotas) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list SortableResourceQuotas) Less(i, j int) bool {
	return list[i].Name < list[j].Name
}
