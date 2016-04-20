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

package latest

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	_ "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api/v1"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/runtime/serializer/json"
	"k8s.io/kubernetes/pkg/runtime/serializer/versioning"
)

// Version is the string that represents the current external default version.
const Version = "v1"

var ExternalVersion = unversioned.GroupVersion{Group: "", Version: "v1"}

// OldestVersion is the string that represents the oldest server version supported,
// for client code that wants to hardcode the lowest common denominator.
const OldestVersion = "v1"

// Versions is the list of versions that are recognized in code. The order provided
// may be assumed to be least feature rich to most feature rich, and clients may
// choose to prefer the latter items in the list over the former items when presented
// with a set of versions to choose.
var Versions = []string{"v1"}

var Codec runtime.Codec

func init() {
	yamlSerializer := json.NewYAMLSerializer(json.DefaultMetaFactory, api.Scheme, runtime.ObjectTyperToTyper(api.Scheme))
	Codec = versioning.NewCodecForScheme(
		api.Scheme,
		yamlSerializer,
		yamlSerializer,
		[]unversioned.GroupVersion{{Version: Version}},
		[]unversioned.GroupVersion{{Version: runtime.APIVersionInternal}},
	)
}
