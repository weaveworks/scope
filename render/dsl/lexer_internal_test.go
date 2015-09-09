package dsl

import (
	"errors"
	"testing"
)

func TestLexer(t *testing.T) {
	for _, test := range []struct {
		input string
		want  []item
		err   error
	}{
		{
			"",
			[]item{},
			errors.New("bad selector"),
		},
		{
			"foo",
			[]item{},
			errors.New("bad selector"),
		},
		{
			"ALL",
			[]item{{itemAll, keywordAll}},
			errors.New("bad transformer"),
		},
		{
			"NOT ALL",
			[]item{{itemNot, keywordNot}, {itemAll, keywordAll}},
			errors.New("bad transformer"),
		},
		{
			"ALL HIGHLIGHT",
			[]item{{itemAll, keywordAll}, {itemHighlight, keywordHighlight}},
			nil,
		},
		{
			"WITH {{pid}} REMOVE",
			[]item{{itemWith, keywordWith}, {itemKeyValue, "pid"}, {itemRemove, keywordRemove}},
			nil,
		},
	} {
		_, c := lex(test.input)
		for item := range c {
			if item.itemType == itemError {
				if test.err == nil {
					t.Errorf("%q: unexpected error: %v", test.input, item.literal)
					break
				}
				if want, have := test.err.Error(), item.literal; want != have {
					t.Errorf("%q: want error %q, have %q", test.input, want, have)
					break
				}
				t.Logf("%q: got expected error %v", test.input, test.err)
				break
			}

			if len(test.want) <= 0 {
				t.Errorf("%q: got too many items", test.input)
				break
			}

			want := test.want[0]
			test.want = test.want[1:]

			if want, have := want.itemType, item.itemType; want != have {
				t.Errorf("%q: unexpected item: want %v, have %v", test.input, want, have)
				break
			}

			t.Logf("%s: lex %s (%q) OK", test.input, item.itemType, item.literal)
		}
	}
}
