package app_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

func topologyServer() *httptest.Server {
	router := mux.NewRouter().SkipClean(true)
	app.RegisterTopologyRoutes(router, app.StaticCollector(fixture.Report), map[string]bool{"foo_capability": true}, "")
	return httptest.NewServer(router)
}

func TestAPIReport(t *testing.T) {
	ts := topologyServer()
	defer ts.Close()

	is404(t, ts, "/api/report/foobar")

	var body = getRawJSON(t, ts, "/api/report")
	// fmt.Printf("Body: %v\n", string(body))
	var r report.Report

	decoder := codec.NewDecoderBytes(body, &codec.JsonHandle{})
	if err := decoder.Decode(&r); err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
}
