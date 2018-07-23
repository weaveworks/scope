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
	"runtime"
	"runtime/debug"
)

// Run the given function inside a behavior context and return the entries
// logged by it.
//
// Functions like Fatalf won't work if the behavior is not executed inside a
// Run context.
func Run(params Params, f func(T)) []Entry {
	behavior := params[BehaviorParam]
	delete(params, BehaviorParam)
	t := entryT{
		params:   params,
		behavior: behavior,
	}

	done := make(chan struct{})

	// We run the function inside a goroutine so that Fatalf can simply call
	// runtime.Goexit to stop execution.
	go func() {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("%v\n%s", err, string(debug.Stack()))
			}
			close(done)
		}()

		if runtime.Version() == "go1.5" {
			// Gnarly workaround for https://github.com/golang/go/issues/12253
			//
			// In short: A bug in Go 1.5 causes a specific form of comparison
			// (runtime.assertE2T2) to leave an invalid pointer on the stack instead
			// of zeroing it out. Usually, this isn't a problem because the function
			// returns afterwards. However, we're consistently hitting a case where
			// another function call is causing the stack to be grown while the
			// pointer is still invalid. The scanner responsible for copying the
			// stack as part of growing it runs into the invalid pointer and
			// crashes.
			//
			// To work around this, we grow the stack significantly beforehand to
			// reduce the likelihood of another growth attempt while the pointer is
			// invalid.
			growStack([1024]int64{})
		}

		f(&t)
	}()

	<-done
	return t.entries
}

func growStack([1024]int64) {
	// Nothing to do
}
