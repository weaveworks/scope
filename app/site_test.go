// Basic site layout tests.
package main

import (
	"net/http/httptest"
	"testing"
)

// Test site
func TestSite(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}))
	defer ts.Close()

	is200(t, ts, "/")
	is200(t, ts, "/index.html")
	is404(t, ts, "/index.html/foobar")
}
