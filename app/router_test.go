package main

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

type v map[string]string

func TestURLMatcher(t *testing.T) {
	test := func(pattern, path string, match bool, vars v) {
		routeMatch := &mux.RouteMatch{}
		if URLMatcher(pattern)(&http.Request{RequestURI: path}, routeMatch) != match {
			t.Fatalf("'%s' '%s'", pattern, path)
		}
		if match && !reflect.DeepEqual(v(routeMatch.Vars), vars) {
			t.Fatalf("%v != %v", v(routeMatch.Vars), vars)
		}
	}

	test("/a/b/c", "/a/b/c", true, v{})
	test("/a/b/c", "/c/b/a", false, v{})
	test("/{a}/b/c", "/b/b/c", true, v{"a": "b"})
	test("/{a}/b/c", "/b/b/b", false, v{})
	test("/a/b/{c}", "/a/b/b", true, v{"c": "b"})
	test("/a/b/{c}", "/a/b/b%2Fb", true, v{"c": "b/b"})
}
