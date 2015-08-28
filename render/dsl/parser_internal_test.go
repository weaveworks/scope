package dsl

import (
	"errors"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestParseExpression(t *testing.T) {
	for _, test := range []struct {
		input string
		want  Expression
		err   error
	}{
		{
			"",
			Expression{},
			errors.New("bad selector"),
		},
		{
			"ALL",
			Expression{selectAll, transformHighlight},
			nil,
		},
		{
			"HIGHLIGHT",
			Expression{},
			errors.New("bad selector"),
		},
		{
			"CONNECTED REMOVE",
			Expression{selectConnected, transformRemove},
			nil,
		},
		{
			"NOT CONNECTED MERGE",
			Expression{selectNot(selectConnected), transformMerge},
			nil,
		},
	} {
		have, err := ParseExpression(test.input)
		if err == nil && test.err != nil {
			t.Errorf("%q: want error %q, have no error", test.input, test.err.Error())
			continue
		} else if err != nil && test.err == nil {
			t.Errorf("%q: want no error, have error %q", test.input, err.Error())
			continue
		} else if err != nil && test.err != nil && test.err.Error() != err.Error() {
			t.Errorf("%q: want error %q, have %q", test.input, test.err.Error(), err.Error())
			continue
		}
		if want, have := nameof(test.want.selector), nameof(have.selector); want != have {
			t.Errorf("%q: selector: want %v, have %v", test.input, want, have)
		}
		if want, have := nameof(test.want.transformer), nameof(have.transformer); want != have {
			t.Errorf("%q: transformer: want %v, have %v", test.input, want, have)
		}
	}
}

func nameof(i interface{}) string {
	full := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name() // github.com/weaveworks/scope/render/dsl.selectAll
	fields := strings.Split(full, ".")                             // [github com/weaveworks/scope/render/dsl selectAll]
	last := fields[len(fields)-1]                                  // selectAll
	return last
}
