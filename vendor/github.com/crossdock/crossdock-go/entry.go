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

import "fmt"

// Status represents the result of running a behavior.
type Status string

// Different valid Statuses.
const (
	Passed  Status = "passed"
	Skipped Status = "skipped"
	Failed  Status = "failed"
)

const (
	statusKey = "status"
	outputKey = "output"
)

// Entry is the most basic form of a test result.
type Entry map[string]interface{}

// Status returns the Status stored in the Entry.
func (e Entry) Status() Status {
	switch v := e[statusKey].(type) {
	case string:
		return Status(v)
	case Status:
		return v
	default:
		panic(fmt.Sprintf("invalid status: %v", v))
	}
}

// Output returns the output attached to the entry.
func (e Entry) Output() string {
	s, ok := e[outputKey].(string)
	if ok {
		return s
	}
	return ""
}
