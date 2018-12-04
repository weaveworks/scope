// Copyright (c) 2018 Uber Technologies, Inc.
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

package instrument

import "github.com/uber-go/tally"

const (
	resultType        = "result_type"
	resultTypeError   = "error"
	resultTypeSuccess = "success"
	timingSuffix      = "latency"
)

// NewCall returns a Call that instruments a function using a given scope
// and a label to name the metrics.
// The following counters are created excluding {{ and }}:
// {{name}}+result_type=success
// {{name}}+result_type=error
// The following timers are created excluding {{ and }} and replacing . with
// the scope's separator:
// {{name}}.latency
func NewCall(scope tally.Scope, name string) Call {
	return &call{
		err:     scope.Tagged(map[string]string{resultType: resultTypeError}).Counter(name),
		success: scope.Tagged(map[string]string{resultType: resultTypeSuccess}).Counter(name),
		timing:  scope.SubScope(name).Timer(timingSuffix),
	}
}

type call struct {
	scope   tally.Scope
	success tally.Counter
	err     tally.Counter
	timing  tally.Timer
}

func (c *call) Exec(f ExecFn) error {
	sw := c.timing.Start()
	err := f()
	sw.Stop()

	if err != nil {
		c.err.Inc(1)
		return err
	}

	c.success.Inc(1)
	return nil
}
