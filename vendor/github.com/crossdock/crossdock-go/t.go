// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package crossdock

import (
	"fmt"
	"runtime"
)

// T records the result of calling different behaviors.
type T interface {
	Behavior() string

	// Look up a behavior parameter.
	Param(key string) string

	// Tag adds the given key-value pair to all entries emitted by this T from
	// this point onwards.
	//
	// If a key with the same name already exists, it will be overwritten.
	// If value is empty, the tag will be deleted from all entries that follow.
	//
	// key MUST NOT be "stauts" or "output".
	Tag(key, value string)

	// Log a failure and continue running the behavior.
	Errorf(format string, args ...interface{})

	// Log a skipped test and continue running the behavior.
	Skipf(format string, args ...interface{})

	// Log a success and continue running the behavior.
	Successf(format string, args ...interface{})

	// Log a failure and stop executing this behavior immediately.
	Fatalf(format string, args ...interface{})

	// Stop executing this behavior immediately.
	FailNow()

	// Put logs an entry with the given status and output. Usually, you'll want
	// to use Errorf, Skipf, Successf or Fatalf instead.
	Put(status Status, output string)
}

// Params represents args to a test
type Params map[string]string

// entryT is a sink that keeps track of entries in-order
type entryT struct {
	behavior string
	params   Params
	tags     map[string]string
	entries  []Entry
}

// Behavior returns the test to dispatch on
func (t entryT) Behavior() string {
	return t.behavior
}

// Param gets a key out of the params map
func (t entryT) Param(key string) string {
	return t.params[key]
}

func (t *entryT) Tag(key, value string) {
	if key == statusKey || key == outputKey {
		panic(fmt.Sprintf("tag %q is reserved", key))
	}

	if t.tags == nil {
		t.tags = make(map[string]string)
	}

	if value == "" {
		delete(t.tags, key)
	} else {
		t.tags[key] = value
	}
}

func (t *entryT) Put(status Status, output string) {
	e := make(Entry)
	e[statusKey] = status
	e[outputKey] = output
	for k, v := range t.tags {
		e[k] = v
	}

	t.entries = append(t.entries, e)
}

func (*entryT) FailNow() {
	// Exit this goroutine and call any deferred functions
	runtime.Goexit()
}

// Skipf records a skipped test.
//
// This may be called multiple times if multiple tests inside a behavior were
// skipped.
func (t *entryT) Skipf(format string, args ...interface{}) {
	t.Put(Skipped, fmt.Sprintf(format, args...))
}

// Errorf records a failed test.
//
// This may be called multiple times if multiple tests inside a behavior
// failed.
func (t *entryT) Errorf(format string, args ...interface{}) {
	t.Put(Failed, fmt.Sprintf(format, args...))
}

// Successf records a successful test.
//
// This may be called multiple times for multiple successful tests inside a
// behavior.
func (t *entryT) Successf(format string, args ...interface{}) {
	t.Put(Passed, fmt.Sprintf(format, args...))
}

// Fatalf records a failed test and fails immediately
func (t *entryT) Fatalf(format string, args ...interface{}) {
	t.Errorf(format, args...)
	t.FailNow()
}
