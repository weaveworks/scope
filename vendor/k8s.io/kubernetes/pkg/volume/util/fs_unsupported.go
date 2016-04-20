// +build !linux,!darwin

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

package util

import (
	"errors"
	"fmt"

	"k8s.io/kubernetes/pkg/api/resource"
)

// FSInfo unsupported returns 0 values for available and capacity and an error.
func FsInfo(path string) (int64, int64, error) {
	return 0, 0, errors.New("FsInfo not supported for this build.")
}

func Du(path string) (*resource.Quantity, error) {
	return nil, fmt.Errorf("Du not support for this build.")
}
