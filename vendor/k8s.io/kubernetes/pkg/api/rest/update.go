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

package rest

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/validation"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/validation/field"
)

// RESTUpdateStrategy defines the minimum validation, accepted input, and
// name generation behavior to update an object that follows Kubernetes
// API conventions. A resource may have many UpdateStrategies, depending on
// the call pattern in use.
type RESTUpdateStrategy interface {
	runtime.ObjectTyper
	// NamespaceScoped returns true if the object must be within a namespace.
	NamespaceScoped() bool
	// AllowCreateOnUpdate returns true if the object can be created by a PUT.
	AllowCreateOnUpdate() bool
	// PrepareForUpdate is invoked on update before validation to normalize
	// the object.  For example: remove fields that are not to be persisted,
	// sort order-insensitive list fields, etc.  This should not remove fields
	// whose presence would be considered a validation error.
	PrepareForUpdate(obj, old runtime.Object)
	// ValidateUpdate is invoked after default fields in the object have been
	// filled in before the object is persisted.  This method should not mutate
	// the object.
	ValidateUpdate(ctx api.Context, obj, old runtime.Object) field.ErrorList
	// Canonicalize is invoked after validation has succeeded but before the
	// object has been persisted.  This method may mutate the object.
	Canonicalize(obj runtime.Object)
	// AllowUnconditionalUpdate returns true if the object can be updated
	// unconditionally (irrespective of the latest resource version), when
	// there is no resource version specified in the object.
	AllowUnconditionalUpdate() bool
}

// TODO: add other common fields that require global validation.
func validateCommonFields(obj, old runtime.Object) (field.ErrorList, error) {
	allErrs := field.ErrorList{}
	objectMeta, err := api.ObjectMetaFor(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get new object metadata: %v", err)
	}
	oldObjectMeta, err := api.ObjectMetaFor(old)
	if err != nil {
		return nil, fmt.Errorf("failed to get old object metadata: %v", err)
	}
	allErrs = append(allErrs, validation.ValidateObjectMetaUpdate(objectMeta, oldObjectMeta, field.NewPath("metadata"))...)

	return allErrs, nil
}

// BeforeUpdate ensures that common operations for all resources are performed on update. It only returns
// errors that can be converted to api.Status. It will invoke update validation with the provided existing
// and updated objects.
func BeforeUpdate(strategy RESTUpdateStrategy, ctx api.Context, obj, old runtime.Object) error {
	objectMeta, kind, kerr := objectMetaAndKind(strategy, obj)
	if kerr != nil {
		return kerr
	}
	if strategy.NamespaceScoped() {
		if !api.ValidNamespace(ctx, objectMeta) {
			return errors.NewBadRequest("the namespace of the provided object does not match the namespace sent on the request")
		}
	} else {
		objectMeta.Namespace = api.NamespaceNone
	}

	strategy.PrepareForUpdate(obj, old)

	// Ensure some common fields, like UID, are validated for all resources.
	errs, err := validateCommonFields(obj, old)
	if err != nil {
		return errors.NewInternalError(err)
	}

	errs = append(errs, strategy.ValidateUpdate(ctx, obj, old)...)
	if len(errs) > 0 {
		return errors.NewInvalid(kind.GroupKind(), objectMeta.Name, errs)
	}

	strategy.Canonicalize(obj)

	return nil
}
