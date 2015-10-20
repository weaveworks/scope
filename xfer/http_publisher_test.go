package xfer_test

import (
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/handlers"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/xfer"
)

func TestHTTPPublisher(t *testing.T) {
	var (
		token = "abcdefg"
		id    = "1234567"
		rpt   = report.MakeReport()
		done  = make(chan struct{})
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if want, have := xfer.AuthorizationHeader(token), r.Header.Get("Authorization"); want != have {
			t.Errorf("want %q, have %q", want, have)
		}

		if want, have := id, r.Header.Get(xfer.ScopeProbeIDHeader); want != have {
			t.Errorf("want %q, have %q", want, have)
		}

		if r.URL.Path == "/api" {
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "irrelevant"})
			return
		}

		var have report.Report

		reader := r.Body
		var err error
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			reader, err = gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer reader.Close()
		}

		if err := gob.NewDecoder(reader).Decode(&have); err != nil {
			t.Error(err)
			return
		}
		if want := rpt; !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
			return
		}
		w.WriteHeader(http.StatusOK)
		close(done)
	})

	s := httptest.NewServer(handlers.CompressHandler(handler))
	defer s.Close()

	_, p, err := xfer.NewHTTPPublisher(s.URL, token, id, false)
	if err != nil {
		t.Fatal(err)
	}
	rp := xfer.NewReportPublisher(p)
	if err := rp.Publish(rpt); err != nil {
		t.Error(err)
	}

	select {
	case <-done:
	case <-time.After(time.Millisecond):
		t.Error("timeout")
	}
}
