package app_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/ugorji/go/codec"

	"$GITHUB_URI/app"
	"$GITHUB_URI/report"
)

func topologyServer() *httptest.Server {
	router := mux.NewRouter().SkipClean(true)
	app.RegisterTopologyRoutes(router, StaticReport{})
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
