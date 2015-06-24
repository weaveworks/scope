package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/weaveworks/scope/test"
)

func TestAPIOriginHost(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}))
	defer ts.Close()

	is404(t, ts, "/api/origin/foobar")
	is404(t, ts, "/api/origin/host/foobar")

	{
		// Origin
		body := getRawJSON(t, ts, fmt.Sprintf("/api/origin/host/%s", test.ServerHostNodeID))
		var o OriginHost
		if err := json.Unmarshal(body, &o); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		if want, have := "Linux", o.OS; want != have {
			t.Errorf("Origin error. Want %v, have %v", want, have)
		}
		if want, have := "0.01 0.01 0.01", o.Load; want != have {
			t.Errorf("Origin error. Want %v, have %v", want, have)
		}
	}
}
