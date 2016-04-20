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

package testing

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type Simple struct {
	unversioned.TypeMeta `json:",inline"`
	api.ObjectMeta       `json:"metadata"`
	Other                string            `json:"other,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
}

func (obj *Simple) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }

type SimpleRoot struct {
	unversioned.TypeMeta `json:",inline"`
	api.ObjectMeta       `json:"metadata"`
	Other                string            `json:"other,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
}

func (obj *SimpleRoot) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }

type SimpleGetOptions struct {
	unversioned.TypeMeta `json:",inline"`
	Param1               string `json:"param1"`
	Param2               string `json:"param2"`
	Path                 string `json:"atAPath"`
}

func (SimpleGetOptions) SwaggerDoc() map[string]string {
	return map[string]string{
		"param1": "description for param1",
		"param2": "description for param2",
	}
}

func (obj *SimpleGetOptions) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }

type SimpleList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,inline"`
	Items                []Simple `json:"items,omitempty"`
}

func (obj *SimpleList) GetObjectKind() unversioned.ObjectKind { return &obj.TypeMeta }
