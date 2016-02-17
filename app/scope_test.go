package app_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		tb.Fatalf("%s:%d: "+msg, append([]interface{}{filepath.Base(file), line}, v...)...)
	}
}

// ok errors the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		tb.Errorf("%s:%d: unexpected error: %v", filepath.Base(file), line, err)
	}
}

// equals errors the test if want is not equal to have.
func equals(tb testing.TB, want, have interface{}) {
	if !reflect.DeepEqual(want, have) {
		_, file, line, _ := runtime.Caller(1)
		tb.Errorf("%s:%d: want %#v, have %#v", filepath.Base(file), line, want, have)
	}
}

// checkGet does a GET and returns the response and the body
func checkGet(t *testing.T, ts *httptest.Server, path string) (*http.Response, []byte) {
	return checkRequest(t, ts, "GET", path, nil)
}

// checkRequest does a 'method'-request (e.g. 'GET') and returns the response and the body
func checkRequest(t *testing.T, ts *httptest.Server, method, path string, body []byte) (*http.Response, []byte) {
	fullPath := ts.URL + path
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, fullPath, bodyReader)
	if err != nil {
		t.Fatalf("Error getting %s: %s %s", method, path, err)
	}
	req.Header.Set("Content-Type", "application/msgpack")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("Error getting %s %s: %s", method, path, err)
	}

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatalf("%s %s body read error: %s", method, path, err)
	}
	return res, body
}

// getRawJSON GETs a file, checks it is JSON, and returns the non-parsed body
func getRawJSON(t *testing.T, ts *httptest.Server, path string) []byte {
	res, body := checkGet(t, ts, path)

	_, file, line, _ := runtime.Caller(1)
	file = filepath.Base(file)
	if res.StatusCode != 200 {
		t.Fatalf("%s:%d: Expected status %d, got %d. Path: %s", file, line, 200, res.StatusCode, path)
	}

	foundCtype := res.Header.Get("content-type")
	if foundCtype != "application/json" {
		t.Errorf("%s:%d: Wrong Content-type for JSON: %s", file, line, foundCtype)
	}

	if len(body) == 0 {
		t.Errorf("%s:%d: No response body", file, line)
	}
	// fmt.Printf("Body: %s", body)

	return body
}

// is200 GETs path and verifies the status code. Returns the body
func is200(t *testing.T, ts *httptest.Server, path string) []byte {
	res, body := checkGet(t, ts, path)
	if res.StatusCode != 200 {
		t.Fatalf("Expected status %d, got %d. Path: %s", 200, res.StatusCode, path)
	}
	return body
}

// is404 GETs path and verifies it returns a 404 status code. Returns the body
func is404(t *testing.T, ts *httptest.Server, path string) []byte {
	res, body := checkGet(t, ts, path)
	if res.StatusCode != 404 {
		t.Fatalf("Expected status %d, got %d", 404, res.StatusCode)
	}
	return body
}

// is400 GETs path and verifies it returns a 400 status code. Returns the body
func is400(t *testing.T, ts *httptest.Server, path string) []byte {
	res, body := checkGet(t, ts, path)
	if res.StatusCode != 400 {
		t.Fatalf("Expected status %d, got %d", 400, res.StatusCode)
	}
	return body
}
