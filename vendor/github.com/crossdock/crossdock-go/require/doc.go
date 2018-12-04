// Copyright (c) 2012 - 2013 Mat Ryer and Tyler Bunnell
//
// Please consider promoting this project if you find it useful.
//
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated documentation
// files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge,
// publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT
// OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// Package require implements the same assertions as the `assert` package but
// stops test execution when a test fails.
//
// Example Usage
//
// The following is a complete example using require in a standard test function:
//    import (
//      "testing"
//      "github.com/crossdock/crossdock-go/require"
//    )
//
//    func TestSomething(t *testing.T) {
//
//      var a string = "Hello"
//      var b string = "Hello"
//
//      require.Equal(t, a, b, "The two words should be the same.")
//
//    }
//
// Assertions
//
// The `require` package have same global functions as in the `assert` package,
// but instead of returning a boolean result they call `t.FailNow()`.
//
// Every assertion function also takes an optional string message as the final argument,
// allowing custom error messages to be appended to the message the assertion method outputs.
package require
