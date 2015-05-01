package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestAPIOrigin(t *testing.T) {
	ts := httptest.NewServer(Router(StaticReport{}, nil))
	defer ts.Close()

	is404(t, ts, "/api/origin/foobar")

	{
		// Origin
		body := getRawJSON(t, ts, "/api/origin/hostA")
		var o Origin
		if err := json.Unmarshal(body, &o); err != nil {
			t.Fatalf("JSON parse error: %s", err)
		}
		if want, have := "Linux", o.OS; want != have {
			t.Errorf("Origin error. Want %v, have %v", want, have)
		}
		if want, have := 3.1415, o.LoadOne; want != have {
			t.Errorf("Origin error. Want %v, have %v", want, have)
		}
		if want, have := 2.7182, o.LoadFive; want != have {
			t.Errorf("Origin error. Want %v, have %v", want, have)
		}
		if want, have := 1.6180, o.LoadFifteen; want != have {
			t.Errorf("Origin error. Want %v, have %v", want, have)
		}
	}
}
