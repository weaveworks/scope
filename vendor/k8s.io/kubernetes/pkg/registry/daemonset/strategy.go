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

package daemonset

import (
	"fmt"
	"reflect"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/extensions/validation"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/validation/field"
)

// daemonSetStrategy implements verification logic for daemon sets.
type daemonSetStrategy struct {
	runtime.ObjectTyper
	api.NameGenerator
}

// Strategy is the default logic that applies when creating and updating DaemonSet objects.
var Strategy = daemonSetStrategy{api.Scheme, api.SimpleNameGenerator}

// NamespaceScoped returns true because all DaemonSets need to be within a namespace.
func (daemonSetStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears the status of a daemon set before creation.
func (daemonSetStrategy) PrepareForCreate(obj runtime.Object) {
	daemonSet := obj.(*extensions.DaemonSet)
	daemonSet.Status = extensions.DaemonSetStatus{}

	daemonSet.Generation = 1
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (daemonSetStrategy) PrepareForUpdate(obj, old runtime.Object) {
	newDaemonSet := obj.(*extensions.DaemonSet)
	oldDaemonSet := old.(*extensions.DaemonSet)

	// update is not allowed to set status
	newDaemonSet.Status = oldDaemonSet.Status

	// Any changes to the spec increment the generation number, any changes to the
	// status should reflect the generation number of the corresponding object. We push
	// the burden of managing the status onto the clients because we can't (in general)
	// know here what version of spec the writer of the status has seen. It may seem like
	// we can at first -- since obj contains spec -- but in the future we will probably make
	// status its own object, and even if we don't, writes may be the result of a
	// read-update-write loop, so the contents of spec may not actually be the spec that
	// the manager has *seen*.
	//
	// TODO: Any changes to a part of the object that represents desired state (labels,
	// annotations etc) should also increment the generation.
	if !reflect.DeepEqual(oldDaemonSet.Spec, newDaemonSet.Spec) {
		newDaemonSet.Generation = oldDaemonSet.Generation + 1
	}
}

// Validate validates a new daemon set.
func (daemonSetStrategy) Validate(ctx api.Context, obj runtime.Object) field.ErrorList {
	daemonSet := obj.(*extensions.DaemonSet)
	return validation.ValidateDaemonSet(daemonSet)
}

// Canonicalize normalizes the object after validation.
func (daemonSetStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for daemon set; this means a POST is
// needed to create one
func (daemonSetStrategy) AllowCreateOnUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user.
func (daemonSetStrategy) ValidateUpdate(ctx api.Context, obj, old runtime.Object) field.ErrorList {
	validationErrorList := validation.ValidateDaemonSet(obj.(*extensions.DaemonSet))
	updateErrorList := validation.ValidateDaemonSetUpdate(obj.(*extensions.DaemonSet), old.(*extensions.DaemonSet))
	return append(validationErrorList, updateErrorList...)
}

// AllowUnconditionalUpdate is the default update policy for daemon set objects.
func (daemonSetStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// DaemonSetToSelectableFields returns a field set that represents the object.
func DaemonSetToSelectableFields(daemon *extensions.DaemonSet) fields.Set {
	return generic.ObjectMetaFieldsSet(daemon.ObjectMeta, true)
}

// MatchSetDaemon is the filter used by the generic etcd backend to route
// watch events from etcd to clients of the apiserver only interested in specific
// labels/fields.
func MatchDaemonSet(label labels.Selector, field fields.Selector) generic.Matcher {
	return &generic.SelectionPredicate{
		Label: label,
		Field: field,
		GetAttrs: func(obj runtime.Object) (labels.Set, fields.Set, error) {
			ds, ok := obj.(*extensions.DaemonSet)
			if !ok {
				return nil, nil, fmt.Errorf("given object is not a ds.")
			}
			return labels.Set(ds.ObjectMeta.Labels), DaemonSetToSelectableFields(ds), nil
		},
	}
}

type daemonSetStatusStrategy struct {
	daemonSetStrategy
}

var StatusStrategy = daemonSetStatusStrategy{Strategy}

func (daemonSetStatusStrategy) PrepareForUpdate(obj, old runtime.Object) {
	newDaemonSet := obj.(*extensions.DaemonSet)
	oldDaemonSet := old.(*extensions.DaemonSet)
	newDaemonSet.Spec = oldDaemonSet.Spec
}

func (daemonSetStatusStrategy) ValidateUpdate(ctx api.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateDaemonSetStatusUpdate(obj.(*extensions.DaemonSet), old.(*extensions.DaemonSet))
}
