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

package labels

import (
	"reflect"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/util/sets"
)

func TestSelectorParse(t *testing.T) {
	testGoodStrings := []string{
		"x=a,y=b,z=c",
		"",
		"x!=a,y=b",
		"x=",
		"x= ",
		"x=,z= ",
		"x= ,z= ",
		"!x",
		"x>1.1",
		"x>1.1,z<5.3",
	}
	testBadStrings := []string{
		"x=a||y=b",
		"x==a==b",
		"!x=a",
		"x<a",
	}
	for _, test := range testGoodStrings {
		lq, err := Parse(test)
		if err != nil {
			t.Errorf("%v: error %v (%#v)\n", test, err, err)
		}
		if strings.Replace(test, " ", "", -1) != lq.String() {
			t.Errorf("%v restring gave: %v\n", test, lq.String())
		}
	}
	for _, test := range testBadStrings {
		_, err := Parse(test)
		if err == nil {
			t.Errorf("%v: did not get expected error\n", test)
		}
	}
}

func TestDeterministicParse(t *testing.T) {
	s1, err := Parse("x=a,a=x")
	s2, err2 := Parse("a=x,x=a")
	if err != nil || err2 != nil {
		t.Errorf("Unexpected parse error")
	}
	if s1.String() != s2.String() {
		t.Errorf("Non-deterministic parse")
	}
}

func expectMatch(t *testing.T, selector string, ls Set) {
	lq, err := Parse(selector)
	if err != nil {
		t.Errorf("Unable to parse %v as a selector\n", selector)
		return
	}
	if !lq.Matches(ls) {
		t.Errorf("Wanted %s to match '%s', but it did not.\n", selector, ls)
	}
}

func expectNoMatch(t *testing.T, selector string, ls Set) {
	lq, err := Parse(selector)
	if err != nil {
		t.Errorf("Unable to parse %v as a selector\n", selector)
		return
	}
	if lq.Matches(ls) {
		t.Errorf("Wanted '%s' to not match '%s', but it did.", selector, ls)
	}
}

func TestEverything(t *testing.T) {
	if !Everything().Matches(Set{"x": "y"}) {
		t.Errorf("Nil selector didn't match")
	}
	if !Everything().Empty() {
		t.Errorf("Everything was not empty")
	}
}

func TestSelectorMatches(t *testing.T) {
	expectMatch(t, "", Set{"x": "y"})
	expectMatch(t, "x=y", Set{"x": "y"})
	expectMatch(t, "x=y,z=w", Set{"x": "y", "z": "w"})
	expectMatch(t, "x!=y,z!=w", Set{"x": "z", "z": "a"})
	expectMatch(t, "notin=in", Set{"notin": "in"}) // in and notin in exactMatch
	expectMatch(t, "x", Set{"x": "z"})
	expectMatch(t, "!x", Set{"y": "z"})
	expectMatch(t, "x>1.1", Set{"x": "1.2"})
	expectMatch(t, "x<1.1", Set{"x": "0.8"})
	expectNoMatch(t, "x=z", Set{})
	expectNoMatch(t, "x=y", Set{"x": "z"})
	expectNoMatch(t, "x=y,z=w", Set{"x": "w", "z": "w"})
	expectNoMatch(t, "x!=y,z!=w", Set{"x": "z", "z": "w"})
	expectNoMatch(t, "x", Set{"y": "z"})
	expectNoMatch(t, "!x", Set{"x": "z"})
	expectNoMatch(t, "x>1.1", Set{"x": "0.8"})
	expectNoMatch(t, "x<1.1", Set{"x": "1.1"})

	labelset := Set{
		"foo": "bar",
		"baz": "blah",
	}
	expectMatch(t, "foo=bar", labelset)
	expectMatch(t, "baz=blah", labelset)
	expectMatch(t, "foo=bar,baz=blah", labelset)
	expectNoMatch(t, "foo=blah", labelset)
	expectNoMatch(t, "baz=bar", labelset)
	expectNoMatch(t, "foo=bar,foobar=bar,baz=blah", labelset)
}

func expectMatchDirect(t *testing.T, selector, ls Set) {
	if !SelectorFromSet(selector).Matches(ls) {
		t.Errorf("Wanted %s to match '%s', but it did not.\n", selector, ls)
	}
}

func expectNoMatchDirect(t *testing.T, selector, ls Set) {
	if SelectorFromSet(selector).Matches(ls) {
		t.Errorf("Wanted '%s' to not match '%s', but it did.", selector, ls)
	}
}

func TestSetMatches(t *testing.T) {
	labelset := Set{
		"foo": "bar",
		"baz": "blah",
	}
	expectMatchDirect(t, Set{}, labelset)
	expectMatchDirect(t, Set{"foo": "bar"}, labelset)
	expectMatchDirect(t, Set{"baz": "blah"}, labelset)
	expectMatchDirect(t, Set{"foo": "bar", "baz": "blah"}, labelset)

	//TODO: bad values not handled for the moment in SelectorFromSet
	//expectNoMatchDirect(t, Set{"foo": "=blah"}, labelset)
	//expectNoMatchDirect(t, Set{"baz": "=bar"}, labelset)
	//expectNoMatchDirect(t, Set{"foo": "=bar", "foobar": "bar", "baz": "blah"}, labelset)
}

func TestNilMapIsValid(t *testing.T) {
	selector := Set(nil).AsSelector()
	if selector == nil {
		t.Errorf("Selector for nil set should be Everything")
	}
	if !selector.Empty() {
		t.Errorf("Selector for nil set should be Empty")
	}
}

func TestSetIsEmpty(t *testing.T) {
	if !(Set{}).AsSelector().Empty() {
		t.Errorf("Empty set should be empty")
	}
	if !(NewSelector()).Empty() {
		t.Errorf("Nil Selector should be empty")
	}
}

func TestLexer(t *testing.T) {
	testcases := []struct {
		s string
		t Token
	}{
		{"", EndOfStringToken},
		{",", CommaToken},
		{"notin", NotInToken},
		{"in", InToken},
		{"=", EqualsToken},
		{"==", DoubleEqualsToken},
		{">", GreaterThanToken},
		{"<", LessThanToken},
		//Note that Lex returns the longest valid token found
		{"!", DoesNotExistToken},
		{"!=", NotEqualsToken},
		{"(", OpenParToken},
		{")", ClosedParToken},
		//Non-"special" characters are considered part of an identifier
		{"~", IdentifierToken},
		{"||", IdentifierToken},
	}
	for _, v := range testcases {
		l := &Lexer{s: v.s, pos: 0}
		token, lit := l.Lex()
		if token != v.t {
			t.Errorf("Got %d it should be %d for '%s'", token, v.t, v.s)
		}
		if v.t != ErrorToken && lit != v.s {
			t.Errorf("Got '%s' it should be '%s'", lit, v.s)
		}
	}
}

func min(l, r int) (m int) {
	m = r
	if l < r {
		m = l
	}
	return m
}

func TestLexerSequence(t *testing.T) {
	testcases := []struct {
		s string
		t []Token
	}{
		{"key in ( value )", []Token{IdentifierToken, InToken, OpenParToken, IdentifierToken, ClosedParToken}},
		{"key notin ( value )", []Token{IdentifierToken, NotInToken, OpenParToken, IdentifierToken, ClosedParToken}},
		{"key in ( value1, value2 )", []Token{IdentifierToken, InToken, OpenParToken, IdentifierToken, CommaToken, IdentifierToken, ClosedParToken}},
		{"key", []Token{IdentifierToken}},
		{"!key", []Token{DoesNotExistToken, IdentifierToken}},
		{"()", []Token{OpenParToken, ClosedParToken}},
		{"x in (),y", []Token{IdentifierToken, InToken, OpenParToken, ClosedParToken, CommaToken, IdentifierToken}},
		{"== != (), = notin", []Token{DoubleEqualsToken, NotEqualsToken, OpenParToken, ClosedParToken, CommaToken, EqualsToken, NotInToken}},
		{"key>1.1", []Token{IdentifierToken, GreaterThanToken, IdentifierToken}},
		{"key<0.8", []Token{IdentifierToken, LessThanToken, IdentifierToken}},
	}
	for _, v := range testcases {
		var literals []string
		var tokens []Token
		l := &Lexer{s: v.s, pos: 0}
		for {
			token, lit := l.Lex()
			if token == EndOfStringToken {
				break
			}
			tokens = append(tokens, token)
			literals = append(literals, lit)
		}
		if len(tokens) != len(v.t) {
			t.Errorf("Bad number of tokens for '%s %d, %d", v.s, len(tokens), len(v.t))
		}
		for i := 0; i < min(len(tokens), len(v.t)); i++ {
			if tokens[i] != v.t[i] {
				t.Errorf("Test '%s': Mismatching in token type found '%v' it should be '%v'", v.s, tokens[i], v.t[i])
			}
		}
	}
}
func TestParserLookahead(t *testing.T) {
	testcases := []struct {
		s string
		t []Token
	}{
		{"key in ( value )", []Token{IdentifierToken, InToken, OpenParToken, IdentifierToken, ClosedParToken, EndOfStringToken}},
		{"key notin ( value )", []Token{IdentifierToken, NotInToken, OpenParToken, IdentifierToken, ClosedParToken, EndOfStringToken}},
		{"key in ( value1, value2 )", []Token{IdentifierToken, InToken, OpenParToken, IdentifierToken, CommaToken, IdentifierToken, ClosedParToken, EndOfStringToken}},
		{"key", []Token{IdentifierToken, EndOfStringToken}},
		{"!key", []Token{DoesNotExistToken, IdentifierToken, EndOfStringToken}},
		{"()", []Token{OpenParToken, ClosedParToken, EndOfStringToken}},
		{"", []Token{EndOfStringToken}},
		{"x in (),y", []Token{IdentifierToken, InToken, OpenParToken, ClosedParToken, CommaToken, IdentifierToken, EndOfStringToken}},
		{"== != (), = notin", []Token{DoubleEqualsToken, NotEqualsToken, OpenParToken, ClosedParToken, CommaToken, EqualsToken, NotInToken, EndOfStringToken}},
		{"key>1.1", []Token{IdentifierToken, GreaterThanToken, IdentifierToken, EndOfStringToken}},
		{"key<0.8", []Token{IdentifierToken, LessThanToken, IdentifierToken, EndOfStringToken}},
	}
	for _, v := range testcases {
		p := &Parser{l: &Lexer{s: v.s, pos: 0}, position: 0}
		p.scan()
		if len(p.scannedItems) != len(v.t) {
			t.Errorf("Expected %d items found %d", len(v.t), len(p.scannedItems))
		}
		for {
			token, lit := p.lookahead(KeyAndOperator)

			token2, lit2 := p.consume(KeyAndOperator)
			if token == EndOfStringToken {
				break
			}
			if token != token2 || lit != lit2 {
				t.Errorf("Bad values")
			}
		}
	}
}

func TestRequirementConstructor(t *testing.T) {
	requirementConstructorTests := []struct {
		Key     string
		Op      Operator
		Vals    sets.String
		Success bool
	}{
		{"x", InOperator, nil, false},
		{"x", NotInOperator, sets.NewString(), false},
		{"x", InOperator, sets.NewString("foo"), true},
		{"x", NotInOperator, sets.NewString("foo"), true},
		{"x", ExistsOperator, nil, true},
		{"x", DoesNotExistOperator, nil, true},
		{"1foo", InOperator, sets.NewString("bar"), true},
		{"1234", InOperator, sets.NewString("bar"), true},
		{"y", GreaterThanOperator, sets.NewString("1.1"), true},
		{"z", LessThanOperator, sets.NewString("5.3"), true},
		{"foo", GreaterThanOperator, sets.NewString("bar"), false},
		{"barz", LessThanOperator, sets.NewString("blah"), false},
		{strings.Repeat("a", 254), ExistsOperator, nil, false}, //breaks DNS rule that len(key) <= 253
	}
	for _, rc := range requirementConstructorTests {
		if _, err := NewRequirement(rc.Key, rc.Op, rc.Vals); err == nil && !rc.Success {
			t.Errorf("expected error with key:%#v op:%v vals:%v, got no error", rc.Key, rc.Op, rc.Vals)
		} else if err != nil && rc.Success {
			t.Errorf("expected no error with key:%#v op:%v vals:%v, got:%v", rc.Key, rc.Op, rc.Vals, err)
		}
	}
}

func TestToString(t *testing.T) {
	var req Requirement
	toStringTests := []struct {
		In    *internalSelector
		Out   string
		Valid bool
	}{

		{&internalSelector{
			getRequirement("x", InOperator, sets.NewString("abc", "def"), t),
			getRequirement("y", NotInOperator, sets.NewString("jkl"), t),
			getRequirement("z", ExistsOperator, nil, t)},
			"x in (abc,def),y notin (jkl),z", true},
		{&internalSelector{
			getRequirement("x", NotInOperator, sets.NewString("abc", "def"), t),
			getRequirement("y", NotEqualsOperator, sets.NewString("jkl"), t),
			getRequirement("z", DoesNotExistOperator, nil, t)},
			"x notin (abc,def),y!=jkl,!z", true},
		{&internalSelector{
			getRequirement("x", InOperator, sets.NewString("abc", "def"), t),
			req}, // adding empty req for the trailing ','
			"x in (abc,def),", false},
		{&internalSelector{
			getRequirement("x", NotInOperator, sets.NewString("abc"), t),
			getRequirement("y", InOperator, sets.NewString("jkl", "mno"), t),
			getRequirement("z", NotInOperator, sets.NewString(""), t)},
			"x notin (abc),y in (jkl,mno),z notin ()", true},
		{&internalSelector{
			getRequirement("x", EqualsOperator, sets.NewString("abc"), t),
			getRequirement("y", DoubleEqualsOperator, sets.NewString("jkl"), t),
			getRequirement("z", NotEqualsOperator, sets.NewString("a"), t),
			getRequirement("z", ExistsOperator, nil, t)},
			"x=abc,y==jkl,z!=a,z", true},
		{&internalSelector{
			getRequirement("x", GreaterThanOperator, sets.NewString("2.4"), t),
			getRequirement("y", LessThanOperator, sets.NewString("7.1"), t),
			getRequirement("z", ExistsOperator, nil, t)},
			"x>2.4,y<7.1,z", true},
	}
	for _, ts := range toStringTests {
		if out := ts.In.String(); out == "" && ts.Valid {
			t.Errorf("%+v.String() => '%v' expected no error", ts.In, out)
		} else if out != ts.Out {
			t.Errorf("%+v.String() => '%v' want '%v'", ts.In, out, ts.Out)
		}
	}
}

func TestRequirementSelectorMatching(t *testing.T) {
	var req Requirement
	labelSelectorMatchingTests := []struct {
		Set   Set
		Sel   Selector
		Match bool
	}{
		{Set{"x": "foo", "y": "baz"}, &internalSelector{
			req,
		}, false},
		{Set{"x": "foo", "y": "baz"}, &internalSelector{
			getRequirement("x", InOperator, sets.NewString("foo"), t),
			getRequirement("y", NotInOperator, sets.NewString("alpha"), t),
		}, true},
		{Set{"x": "foo", "y": "baz"}, &internalSelector{
			getRequirement("x", InOperator, sets.NewString("foo"), t),
			getRequirement("y", InOperator, sets.NewString("alpha"), t),
		}, false},
		{Set{"y": ""}, &internalSelector{
			getRequirement("x", NotInOperator, sets.NewString(""), t),
			getRequirement("y", ExistsOperator, nil, t),
		}, true},
		{Set{"y": ""}, &internalSelector{
			getRequirement("x", DoesNotExistOperator, nil, t),
			getRequirement("y", ExistsOperator, nil, t),
		}, true},
		{Set{"y": ""}, &internalSelector{
			getRequirement("x", NotInOperator, sets.NewString(""), t),
			getRequirement("y", DoesNotExistOperator, nil, t),
		}, false},
		{Set{"y": "baz"}, &internalSelector{
			getRequirement("x", InOperator, sets.NewString(""), t),
		}, false},
		{Set{"z": "1.2"}, &internalSelector{
			getRequirement("z", GreaterThanOperator, sets.NewString("1.0"), t),
		}, true},
		{Set{"z": "v1.2"}, &internalSelector{
			getRequirement("z", GreaterThanOperator, sets.NewString("1.0"), t),
		}, false},
	}
	for _, lsm := range labelSelectorMatchingTests {
		if match := lsm.Sel.Matches(lsm.Set); match != lsm.Match {
			t.Errorf("%+v.Matches(%#v) => %v, want %v", lsm.Sel, lsm.Set, match, lsm.Match)
		}
	}
}

func TestSetSelectorParser(t *testing.T) {
	setSelectorParserTests := []struct {
		In    string
		Out   Selector
		Match bool
		Valid bool
	}{
		{"", NewSelector(), true, true},
		{"\rx", internalSelector{
			getRequirement("x", ExistsOperator, nil, t),
		}, true, true},
		{"this-is-a-dns.domain.com/key-with-dash", internalSelector{
			getRequirement("this-is-a-dns.domain.com/key-with-dash", ExistsOperator, nil, t),
		}, true, true},
		{"this-is-another-dns.domain.com/key-with-dash in (so,what)", internalSelector{
			getRequirement("this-is-another-dns.domain.com/key-with-dash", InOperator, sets.NewString("so", "what"), t),
		}, true, true},
		{"0.1.2.domain/99 notin (10.10.100.1, tick.tack.clock)", internalSelector{
			getRequirement("0.1.2.domain/99", NotInOperator, sets.NewString("10.10.100.1", "tick.tack.clock"), t),
		}, true, true},
		{"foo  in	 (abc)", internalSelector{
			getRequirement("foo", InOperator, sets.NewString("abc"), t),
		}, true, true},
		{"x notin\n (abc)", internalSelector{
			getRequirement("x", NotInOperator, sets.NewString("abc"), t),
		}, true, true},
		{"x  notin	\t	(abc,def)", internalSelector{
			getRequirement("x", NotInOperator, sets.NewString("abc", "def"), t),
		}, true, true},
		{"x in (abc,def)", internalSelector{
			getRequirement("x", InOperator, sets.NewString("abc", "def"), t),
		}, true, true},
		{"x in (abc,)", internalSelector{
			getRequirement("x", InOperator, sets.NewString("abc", ""), t),
		}, true, true},
		{"x in ()", internalSelector{
			getRequirement("x", InOperator, sets.NewString(""), t),
		}, true, true},
		{"x notin (abc,,def),bar,z in (),w", internalSelector{
			getRequirement("bar", ExistsOperator, nil, t),
			getRequirement("w", ExistsOperator, nil, t),
			getRequirement("x", NotInOperator, sets.NewString("abc", "", "def"), t),
			getRequirement("z", InOperator, sets.NewString(""), t),
		}, true, true},
		{"x,y in (a)", internalSelector{
			getRequirement("y", InOperator, sets.NewString("a"), t),
			getRequirement("x", ExistsOperator, nil, t),
		}, false, true},
		{"x=a", internalSelector{
			getRequirement("x", EqualsOperator, sets.NewString("a"), t),
		}, true, true},
		{"x>1.1", internalSelector{
			getRequirement("x", GreaterThanOperator, sets.NewString("1.1"), t),
		}, true, true},
		{"x<7.1", internalSelector{
			getRequirement("x", LessThanOperator, sets.NewString("7.1"), t),
		}, true, true},
		{"x=a,y!=b", internalSelector{
			getRequirement("x", EqualsOperator, sets.NewString("a"), t),
			getRequirement("y", NotEqualsOperator, sets.NewString("b"), t),
		}, true, true},
		{"x=a,y!=b,z in (h,i,j)", internalSelector{
			getRequirement("x", EqualsOperator, sets.NewString("a"), t),
			getRequirement("y", NotEqualsOperator, sets.NewString("b"), t),
			getRequirement("z", InOperator, sets.NewString("h", "i", "j"), t),
		}, true, true},
		{"x=a||y=b", internalSelector{}, false, false},
		{"x,,y", nil, true, false},
		{",x,y", nil, true, false},
		{"x nott in (y)", nil, true, false},
		{"x notin ( )", internalSelector{
			getRequirement("x", NotInOperator, sets.NewString(""), t),
		}, true, true},
		{"x notin (, a)", internalSelector{
			getRequirement("x", NotInOperator, sets.NewString("", "a"), t),
		}, true, true},
		{"a in (xyz),", nil, true, false},
		{"a in (xyz)b notin ()", nil, true, false},
		{"a ", internalSelector{
			getRequirement("a", ExistsOperator, nil, t),
		}, true, true},
		{"a in (x,y,notin, z,in)", internalSelector{
			getRequirement("a", InOperator, sets.NewString("in", "notin", "x", "y", "z"), t),
		}, true, true}, // operator 'in' inside list of identifiers
		{"a in (xyz abc)", nil, false, false}, // no comma
		{"a notin(", nil, true, false},        // bad formed
		{"a (", nil, false, false},            // cpar
		{"(", nil, false, false},              // opar
	}

	for _, ssp := range setSelectorParserTests {
		if sel, err := Parse(ssp.In); err != nil && ssp.Valid {
			t.Errorf("Parse(%s) => %v expected no error", ssp.In, err)
		} else if err == nil && !ssp.Valid {
			t.Errorf("Parse(%s) => %+v expected error", ssp.In, sel)
		} else if ssp.Match && !reflect.DeepEqual(sel, ssp.Out) {
			t.Errorf("Parse(%s) => parse output '%#v' doesn't match '%#v' expected match", ssp.In, sel, ssp.Out)
		}
	}
}

func getRequirement(key string, op Operator, vals sets.String, t *testing.T) Requirement {
	req, err := NewRequirement(key, op, vals)
	if err != nil {
		t.Errorf("NewRequirement(%v, %v, %v) resulted in error:%v", key, op, vals, err)
		return Requirement{}
	}
	return *req
}

func TestAdd(t *testing.T) {
	testCases := []struct {
		name        string
		sel         Selector
		key         string
		operator    Operator
		values      []string
		refSelector Selector
	}{
		{
			"keyInOperator",
			internalSelector{},
			"key",
			InOperator,
			[]string{"value"},
			internalSelector{Requirement{"key", InOperator, sets.NewString("value")}},
		},
		{
			"keyEqualsOperator",
			internalSelector{Requirement{"key", InOperator, sets.NewString("value")}},
			"key2",
			EqualsOperator,
			[]string{"value2"},
			internalSelector{
				Requirement{"key", InOperator, sets.NewString("value")},
				Requirement{"key2", EqualsOperator, sets.NewString("value2")},
			},
		},
	}
	for _, ts := range testCases {
		req, err := NewRequirement(ts.key, ts.operator, sets.NewString(ts.values...))
		if err != nil {
			t.Errorf("%s - Unable to create labels.Requirement", ts.name)
		}
		ts.sel = ts.sel.Add(*req)
		if !reflect.DeepEqual(ts.sel, ts.refSelector) {
			t.Errorf("%s - Expected %v found %v", ts.name, ts.refSelector, ts.sel)
		}
	}
}
