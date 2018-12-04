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
	"time"

	"github.com/crossdock/crossdock-go/assert"
	"github.com/crossdock/crossdock-go/require"
)

// Assert builds an Assertions object that logs success or failure for all
// operations to the given T. The behavior will continue executing in case of
// failure.
//
// The following will log exactly len(tests) entries.
//
// 	assert := Assert(t)
// 	for _, tt := range tests {
// 		assert.Equals(tt.want, f(tt.give), "expected f(%v) == %v", tt.give, tt.want)
// 	}
func Assert(t T) Assertions {
	return sinkAssertions{t, assert.New(sinkTestingT{t})}
}

// Checks builds an Assertions object that logs only failures to the given T.
// The behavior will continue executing in case of failure.
//
// The following will log only as many entries as invalid test cases.
//
// 	checks := Checks(t)
// 	for _, tt := range tests {
// 		checks.Equals(tt.want, f(tt.give), "expected f(%v) == %v", tt.give, tt.want)
// 	}
func Checks(t T) Assertions {
	return assert.New(sinkTestingT{t})
}

// Require builds an Assertions object that logs success or failure for all
// operations to the given T. Execution of the behavior will be terminated
// immediately on the first failing assertion.
//
// The following will log one entry for each successful test case starting at
// the first one and the first failure that is encountered.
//
// 	require := Require(t)
// 	for _, tt := range tests {
// 		require.Equals(tt.want, f(tt.give), "expected f(%v) == %v", tt.give, tt.want)
// 	}
func Require(t T) Assertions {
	return sinkAssertions{t, requireAssertions{require.New(sinkTestingT{t})}}
}

// Fatals builds an Assertions object that logs only failures to the given
// T. Execution of the behavior will be terminated immediately on the first
// failing assertion.
//
// The following will log the first failure encountered or nothing if all test
// cases were succesful.
//
// 	fatals := Fatals(t)
// 	for _, tt := range tests {
// 		fatals.Equals(tt.want, f(tt.give), "expected f(%v) == %v", tt.give, tt.want)
// 	}
func Fatals(t T) Assertions {
	return requireAssertions{require.New(sinkTestingT{t})}
}

// sinkTestingT adapts a crossdock.T into an {require,assert}.TestingT
type sinkTestingT struct{ t T }

func (st sinkTestingT) FailNow() { st.t.FailNow() }

func (st sinkTestingT) Errorf(format string, args ...interface{}) {
	// We need to prepend a newline because the error message from testify
	// always includes a \r at the start.
	st.t.Errorf("\n"+format, args...)
}

//////////////////////////////////////////////////////////////////////////////
// Assertions

// Assertions provides helpers to assert conditions in crossdock behaviors.
//
// All assertions can include informative error messages formatted using
// fmt.Sprintf style,
//
// 	assert := Assert(t)
// 	assert.Contains(foo, "bar", "expected to find 'bar' in %q", foo)
//
// All assert operations return true if the condition was met and false
// otherwise. This allows gating operations that would otherwise panic behind
// preconditions.
//
// 	if assert.Error(t, err, "expected failure") {
// 		assert.Contains(t, err.Error(), "something went wrong", "error message mismatch")
// 	}
//
// Additionally, in case of failure, all Assertions make an attempt to
// provide a stack trace in the error message.
//
// Four kinds of Assertions objects are offered via the corresponding
// functions:
//
// Assert(T): All asserts will result in a success or failure being logged to the
// crossdock.T. Execution will continue on failure.
//
// Checks(T): Only failures will be logged to crossdock.T. Execution will
// continue on failure.
//
// Require(T): All asserts will result in a success or failure being logged to
// the crossdock.T. Execution of the behavior will be terminated immediately
// on failure.
//
// Fatals(t): Only failures will be logged to crossdock.T. Execution of the
// behavior will be temrinated immediately on failure.
//
// 	                     +--------+--------+---------+--------+
// 	                     | Assert | Checks | Require | Fatals |
// 	+--------------------+--------+--------+---------+--------+
// 	| Log on success     | Yes    | No     | Yes     | No     |
// 	+--------------------+--------+--------+---------+--------+
// 	| Continue execution | Yes    | Yes    | No      | No     |
// 	| on failure         |        |        |         |        |
// 	+--------------------+--------+--------+---------+--------+
//
type Assertions interface {
	Condition(comp assert.Comparison, msgAndArgs ...interface{}) bool
	Contains(s interface{}, contains interface{}, msgAndArgs ...interface{}) bool
	Empty(object interface{}, msgAndArgs ...interface{}) bool
	Equal(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool
	EqualError(theError error, errString string, msgAndArgs ...interface{}) bool
	EqualValues(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool
	Error(err error, msgAndArgs ...interface{}) bool
	Exactly(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool
	Fail(failureMessage string, msgAndArgs ...interface{}) bool
	FailNow(failureMessage string, msgAndArgs ...interface{}) bool
	False(value bool, msgAndArgs ...interface{}) bool
	Implements(interfaceObject interface{}, object interface{}, msgAndArgs ...interface{}) bool
	InDelta(expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{}) bool
	InDeltaSlice(expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{}) bool
	InEpsilon(expected interface{}, actual interface{}, epsilon float64, msgAndArgs ...interface{}) bool
	InEpsilonSlice(expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{}) bool
	IsType(expectedType interface{}, object interface{}, msgAndArgs ...interface{}) bool
	JSONEq(expected string, actual string, msgAndArgs ...interface{}) bool
	Len(object interface{}, length int, msgAndArgs ...interface{}) bool
	Nil(object interface{}, msgAndArgs ...interface{}) bool
	NoError(err error, msgAndArgs ...interface{}) bool
	NotContains(s interface{}, contains interface{}, msgAndArgs ...interface{}) bool
	NotEmpty(object interface{}, msgAndArgs ...interface{}) bool
	NotEqual(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool
	NotNil(object interface{}, msgAndArgs ...interface{}) bool
	NotPanics(f assert.PanicTestFunc, msgAndArgs ...interface{}) bool
	NotRegexp(rx interface{}, str interface{}, msgAndArgs ...interface{}) bool
	NotZero(i interface{}, msgAndArgs ...interface{}) bool
	Panics(f assert.PanicTestFunc, msgAndArgs ...interface{}) bool
	Regexp(rx interface{}, str interface{}, msgAndArgs ...interface{}) bool
	True(value bool, msgAndArgs ...interface{}) bool
	WithinDuration(expected time.Time, actual time.Time, delta time.Duration, msgAndArgs ...interface{}) bool
	Zero(i interface{}, msgAndArgs ...interface{}) bool
}

//////////////////////////////////////////////////////////////////////////////
// T => TestingT

func formatMsgAndArgs(msgAndArgs []interface{}) string {
	if len(msgAndArgs) == 0 {
		return ""
	}
	if len(msgAndArgs) == 1 {
		return msgAndArgs[0].(string)
	}
	return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
}

type sinkAssertions struct {
	// We need to wrap assert rather than using it as-is because we need to
	// log success messages.
	t T
	a Assertions
}

var _ Assertions = (*sinkAssertions)(nil)

func (sa sinkAssertions) Condition(comp assert.Comparison, msgAndArgs ...interface{}) bool {
	if sa.a.Condition(comp, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Contains(s interface{}, contains interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.Contains(s, contains, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Empty(object interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.Empty(object, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Equal(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.Equal(expected, actual, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) EqualError(theError error, errString string, msgAndArgs ...interface{}) bool {
	if sa.a.EqualError(theError, errString, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) EqualValues(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.EqualValues(expected, actual, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Error(err error, msgAndArgs ...interface{}) bool {
	if sa.a.Error(err, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Exactly(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.Exactly(expected, actual, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Fail(failureMessage string, msgAndArgs ...interface{}) bool {
	if sa.a.Fail(failureMessage, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) FailNow(failureMessage string, msgAndArgs ...interface{}) bool {
	if sa.a.FailNow(failureMessage, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) False(value bool, msgAndArgs ...interface{}) bool {
	if sa.a.False(value, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Implements(interfaceObject interface{}, object interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.Implements(interfaceObject, object, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) InDelta(expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{}) bool {
	if sa.a.InDelta(expected, actual, delta, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) InDeltaSlice(expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{}) bool {
	if sa.a.InDeltaSlice(expected, actual, delta, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) InEpsilon(expected interface{}, actual interface{}, epsilon float64, msgAndArgs ...interface{}) bool {
	if sa.a.InEpsilon(expected, actual, epsilon, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) InEpsilonSlice(expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{}) bool {
	if sa.a.InEpsilonSlice(expected, actual, delta, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) IsType(expectedType interface{}, object interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.IsType(expectedType, object, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) JSONEq(expected string, actual string, msgAndArgs ...interface{}) bool {
	if sa.a.JSONEq(expected, actual, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Len(object interface{}, length int, msgAndArgs ...interface{}) bool {
	if sa.a.Len(object, length, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Nil(object interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.Nil(object, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) NoError(err error, msgAndArgs ...interface{}) bool {
	if sa.a.NoError(err, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) NotContains(s interface{}, contains interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.NotContains(s, contains, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) NotEmpty(object interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.NotEmpty(object, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) NotEqual(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.NotEqual(expected, actual, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) NotNil(object interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.NotNil(object, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) NotPanics(f assert.PanicTestFunc, msgAndArgs ...interface{}) bool {
	if sa.a.NotPanics(f, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) NotRegexp(rx interface{}, str interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.NotRegexp(rx, str, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) NotZero(i interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.NotZero(i, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Panics(f assert.PanicTestFunc, msgAndArgs ...interface{}) bool {
	if sa.a.Panics(f, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Regexp(rx interface{}, str interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.Regexp(rx, str, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) True(value bool, msgAndArgs ...interface{}) bool {
	if sa.a.True(value, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) WithinDuration(expected time.Time, actual time.Time, delta time.Duration, msgAndArgs ...interface{}) bool {
	if sa.a.WithinDuration(expected, actual, delta, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

func (sa sinkAssertions) Zero(i interface{}, msgAndArgs ...interface{}) bool {
	if sa.a.Zero(i, msgAndArgs...) {
		sa.t.Successf(formatMsgAndArgs(msgAndArgs))
		return true
	}
	return false
}

//////////////////////////////////////////////////////////////////////////////
// Require

// Adapts a require.Assertions into an Assertions. This simply returns true
// for all cases because execution just stops in case of failure.
type requireAssertions struct{ r *require.Assertions }

var _ Assertions = (*requireAssertions)(nil)

func (r requireAssertions) Condition(comp assert.Comparison, msgAndArgs ...interface{}) bool {
	r.r.Condition(comp, msgAndArgs...)
	return true
}

func (r requireAssertions) Contains(s interface{}, contains interface{}, msgAndArgs ...interface{}) bool {
	r.r.Contains(s, contains, msgAndArgs...)
	return true
}

func (r requireAssertions) Empty(object interface{}, msgAndArgs ...interface{}) bool {
	r.r.Empty(object, msgAndArgs...)
	return true
}

func (r requireAssertions) Equal(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	r.r.Equal(expected, actual, msgAndArgs...)
	return true
}

func (r requireAssertions) EqualError(theError error, errString string, msgAndArgs ...interface{}) bool {
	r.r.EqualError(theError, errString, msgAndArgs...)
	return true
}

func (r requireAssertions) EqualValues(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	r.r.EqualValues(expected, actual, msgAndArgs...)
	return true
}

func (r requireAssertions) Error(err error, msgAndArgs ...interface{}) bool {
	r.r.Error(err, msgAndArgs...)
	return true
}

func (r requireAssertions) Exactly(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	r.r.Exactly(expected, actual, msgAndArgs...)
	return true
}

func (r requireAssertions) Fail(failureMessage string, msgAndArgs ...interface{}) bool {
	r.r.Fail(failureMessage, msgAndArgs...)
	return true
}

func (r requireAssertions) FailNow(failureMessage string, msgAndArgs ...interface{}) bool {
	r.r.FailNow(failureMessage, msgAndArgs...)
	return true
}

func (r requireAssertions) False(value bool, msgAndArgs ...interface{}) bool {
	r.r.False(value, msgAndArgs...)
	return true
}

func (r requireAssertions) Implements(interfaceObject interface{}, object interface{}, msgAndArgs ...interface{}) bool {
	r.r.Implements(interfaceObject, object, msgAndArgs...)
	return true
}

func (r requireAssertions) InDelta(expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{}) bool {
	r.r.InDelta(expected, actual, delta, msgAndArgs...)
	return true
}

func (r requireAssertions) InDeltaSlice(expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{}) bool {
	r.r.InDeltaSlice(expected, actual, delta, msgAndArgs...)
	return true
}

func (r requireAssertions) InEpsilon(expected interface{}, actual interface{}, epsilon float64, msgAndArgs ...interface{}) bool {
	r.r.InEpsilon(expected, actual, epsilon, msgAndArgs...)
	return true
}

func (r requireAssertions) InEpsilonSlice(expected interface{}, actual interface{}, delta float64, msgAndArgs ...interface{}) bool {
	r.r.InEpsilonSlice(expected, actual, delta, msgAndArgs...)
	return true
}

func (r requireAssertions) IsType(expectedType interface{}, object interface{}, msgAndArgs ...interface{}) bool {
	r.r.IsType(expectedType, object, msgAndArgs...)
	return true
}

func (r requireAssertions) JSONEq(expected string, actual string, msgAndArgs ...interface{}) bool {
	r.r.JSONEq(expected, actual, msgAndArgs...)
	return true
}

func (r requireAssertions) Len(object interface{}, length int, msgAndArgs ...interface{}) bool {
	r.r.Len(object, length, msgAndArgs...)
	return true
}

func (r requireAssertions) Nil(object interface{}, msgAndArgs ...interface{}) bool {
	r.r.Nil(object, msgAndArgs...)
	return true
}

func (r requireAssertions) NoError(err error, msgAndArgs ...interface{}) bool {
	r.r.NoError(err, msgAndArgs...)
	return true
}

func (r requireAssertions) NotContains(s interface{}, contains interface{}, msgAndArgs ...interface{}) bool {
	r.r.NotContains(s, contains, msgAndArgs...)
	return true
}

func (r requireAssertions) NotEmpty(object interface{}, msgAndArgs ...interface{}) bool {
	r.r.NotEmpty(object, msgAndArgs...)
	return true
}

func (r requireAssertions) NotEqual(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	r.r.NotEqual(expected, actual, msgAndArgs...)
	return true
}

func (r requireAssertions) NotNil(object interface{}, msgAndArgs ...interface{}) bool {
	r.r.NotNil(object, msgAndArgs...)
	return true
}

func (r requireAssertions) NotPanics(f assert.PanicTestFunc, msgAndArgs ...interface{}) bool {
	r.r.NotPanics(f, msgAndArgs...)
	return true
}

func (r requireAssertions) NotRegexp(rx interface{}, str interface{}, msgAndArgs ...interface{}) bool {
	r.r.NotRegexp(rx, str, msgAndArgs...)
	return true
}

func (r requireAssertions) NotZero(i interface{}, msgAndArgs ...interface{}) bool {
	r.r.NotZero(i, msgAndArgs...)
	return true
}

func (r requireAssertions) Panics(f assert.PanicTestFunc, msgAndArgs ...interface{}) bool {
	r.r.Panics(f, msgAndArgs...)
	return true
}

func (r requireAssertions) Regexp(rx interface{}, str interface{}, msgAndArgs ...interface{}) bool {
	r.r.Regexp(rx, str, msgAndArgs...)
	return true
}

func (r requireAssertions) True(value bool, msgAndArgs ...interface{}) bool {
	r.r.True(value, msgAndArgs...)
	return true
}

func (r requireAssertions) WithinDuration(expected time.Time, actual time.Time, delta time.Duration, msgAndArgs ...interface{}) bool {
	r.r.WithinDuration(expected, actual, delta, msgAndArgs...)
	return true
}

func (r requireAssertions) Zero(i interface{}, msgAndArgs ...interface{}) bool {
	r.r.Zero(i, msgAndArgs...)
	return true
}
