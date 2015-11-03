// Basic site layout tests.
package main

import (
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// Test site
func TestSite(t *testing.T) {
	router := mux.NewRouter()
	registerStatic(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	is200(t, ts, "/")
	is200(t, ts, "/index.html")
	is404(t, ts, "/index.html/foobar")
}
