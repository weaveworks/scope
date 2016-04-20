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

package atomic

import (
	"sync"
)

// TODO(ArtfulCoder)
// sync/atomic/Value was added in golang 1.4
// Once support is dropped for go 1.3, this type must be deprecated in favor of sync/atomic/Value.
// The functions are named Load/Store to match sync/atomic/Value function names.
type Value struct {
	value      interface{}
	valueMutex sync.RWMutex
}

func (at *Value) Store(val interface{}) {
	at.valueMutex.Lock()
	defer at.valueMutex.Unlock()
	at.value = val
}

func (at *Value) Load() interface{} {
	at.valueMutex.RLock()
	defer at.valueMutex.RUnlock()
	return at.value
}
