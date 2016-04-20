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

package horizontalpodautoscaler

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/extensions/validation"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/validation/field"
)

// autoscalerStrategy implements behavior for HorizontalPodAutoscalers
type autoscalerStrategy struct {
	runtime.ObjectTyper
	api.NameGenerator
}

// Strategy is the default logic that applies when creating and updating HorizontalPodAutoscaler
// objects via the REST API.
var Strategy = autoscalerStrategy{api.Scheme, api.SimpleNameGenerator}

// NamespaceScoped is true for autoscaler.
func (autoscalerStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (autoscalerStrategy) PrepareForCreate(obj runtime.Object) {
	newHPA := obj.(*extensions.HorizontalPodAutoscaler)

	// create cannot set status
	newHPA.Status = extensions.HorizontalPodAutoscalerStatus{}
}

// Validate validates a new autoscaler.
func (autoscalerStrategy) Validate(ctx api.Context, obj runtime.Object) field.ErrorList {
	autoscaler := obj.(*extensions.HorizontalPodAutoscaler)
	return validation.ValidateHorizontalPodAutoscaler(autoscaler)
}

// Canonicalize normalizes the object after validation.
func (autoscalerStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for autoscalers.
func (autoscalerStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (autoscalerStrategy) PrepareForUpdate(obj, old runtime.Object) {
	newHPA := obj.(*extensions.HorizontalPodAutoscaler)
	oldHPA := obj.(*extensions.HorizontalPodAutoscaler)
	// Update is not allowed to set status
	newHPA.Status = oldHPA.Status
}

// ValidateUpdate is the default update validation for an end user.
func (autoscalerStrategy) ValidateUpdate(ctx api.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateHorizontalPodAutoscalerUpdate(obj.(*extensions.HorizontalPodAutoscaler), old.(*extensions.HorizontalPodAutoscaler))
}

func (autoscalerStrategy) AllowUnconditionalUpdate() bool {
	return true
}

func AutoscalerToSelectableFields(limitRange *extensions.HorizontalPodAutoscaler) fields.Set {
	return fields.Set{}
}

func MatchAutoscaler(label labels.Selector, field fields.Selector) generic.Matcher {
	return &generic.SelectionPredicate{
		Label: label,
		Field: field,
		GetAttrs: func(obj runtime.Object) (labels.Set, fields.Set, error) {
			hpa, ok := obj.(*extensions.HorizontalPodAutoscaler)
			if !ok {
				return nil, nil, fmt.Errorf("given object is not a horizontal pod autoscaler.")
			}
			return labels.Set(hpa.ObjectMeta.Labels), AutoscalerToSelectableFields(hpa), nil
		},
	}
}

type autoscalerStatusStrategy struct {
	autoscalerStrategy
}

var StatusStrategy = autoscalerStatusStrategy{Strategy}

func (autoscalerStatusStrategy) PrepareForUpdate(obj, old runtime.Object) {
	newAutoscaler := obj.(*extensions.HorizontalPodAutoscaler)
	oldAutoscaler := old.(*extensions.HorizontalPodAutoscaler)
	// status changes are not allowed to update spec
	newAutoscaler.Spec = oldAutoscaler.Spec
}

func (autoscalerStatusStrategy) ValidateUpdate(ctx api.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateHorizontalPodAutoscalerStatusUpdate(obj.(*extensions.HorizontalPodAutoscaler), old.(*extensions.HorizontalPodAutoscaler))
}
