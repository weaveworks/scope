package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/cello/report"
)

// Test /api/report
func TestAPIReport(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}, nil))
	defer ts.Close()

	is404(t, ts, "/api/report/foobar")

	var body = getRawJSON(t, ts, "/api/report")
	// fmt.Printf("Body: %v\n", string(body))
	var r report.Report
	err := json.Unmarshal(body, &r)
	if err != nil {
		t.Fatalf("JSON parse error: %s", err)
	}
}
