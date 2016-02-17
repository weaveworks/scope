package app_test

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
)

type v map[string]string

func TestURLMatcher(t *testing.T) {
	test := func(pattern, path string, match bool, vars v) {
		routeMatch := &mux.RouteMatch{}
		if app.URLMatcher(pattern)(&http.Request{RequestURI: path}, routeMatch) != match {
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
		router := mux.NewRouter()
		c := app.NewCollector(1 * time.Minute)
		app.RegisterReportPostHandler(c, router)
		ts := httptest.NewServer(router)
		defer ts.Close()

		b, err := encoder(fixture.Report)
		if err != nil {
			t.Fatalf("Content-Type %s: %s", contentType, err)
		}

		req, err := http.NewRequest("POST", ts.URL+"/api/report", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("Error posting report: %v", err)
		}
		req.Header.Set("Content-Type", contentType)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Error posting report %v", err)
		}

		_, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatalf("Error posting report: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Error posting report: %d", resp.StatusCode)
		}

		if want, have := fixture.Report.Endpoint.Nodes, c.Report().Endpoint.Nodes; len(have) == 0 || len(want) != len(have) {
			t.Fatalf("Content-Type %s: %v", contentType, test.Diff(have, want))
		}
	}

	test("", func(v interface{}) ([]byte, error) {
		buf := &bytes.Buffer{}
		err := gob.NewEncoder(buf).Encode(v)
		return buf.Bytes(), err
	})
	test("application/json", func(v interface{}) ([]byte, error) {
		buf := &bytes.Buffer{}
		err := codec.NewEncoder(buf, &codec.JsonHandle{}).Encode(v)
		return buf.Bytes(), err
	})
	test("application/msgpack", func(v interface{}) ([]byte, error) {
		buf := &bytes.Buffer{}
		err := codec.NewEncoder(buf, &codec.MsgpackHandle{}).Encode(v)
		return buf.Bytes(), err
	})
}
