package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
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

func TestReportPostHandler(t *testing.T) {
	test := func(contentType string, encoder func(interface{}) ([]byte, error)) {
		b, err := encoder(fixture.Report)
		if err != nil {
			t.Fatalf("Content-Type %s: %s", contentType, err)
		}

		r, _ := http.NewRequest("POST", "/api/report", bytes.NewReader(b))
		r.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		c := NewCollector(1 * time.Minute)
		makeReportPostHandler(c).ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("Content-Type %s: http status: %d\nbody: %s", contentType, w.Code, w.Body.String())
		}
		// Just check a few items, to confirm it parsed. Otherwise
		// reflect.DeepEqual chokes on nil vs empty arrays.
		if want, have := fixture.Report.Endpoint.Nodes, c.Report().Endpoint.Nodes; len(have) == 0 || len(want) != len(have) {
			t.Fatalf("Content-Type %s: %v", contentType, test.Diff(have, want))
		}
	}

	test("", func(v interface{}) ([]byte, error) {
		buf := &bytes.Buffer{}
		err := gob.NewEncoder(buf).Encode(v)
		return buf.Bytes(), err
	})
	test("application/json", json.Marshal)
}
