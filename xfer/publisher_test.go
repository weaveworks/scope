package xfer_test

import (
	"compress/gzip"
	"encoding/gob"
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

	p, err := xfer.NewHTTPPublisher(s.URL, token, id)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Publish(rpt); err != nil {
		t.Error(err)
	}

	select {
	case <-done:
	case <-time.After(time.Millisecond):
		t.Error("timeout")
	}
}

func TestMultiPublisher(t *testing.T) {
	var (
		p              = &mockPublisher{}
		factory        = func(string) (xfer.Publisher, error) { return p, nil }
		multiPublisher = xfer.NewMultiPublisher(factory)
	)

	multiPublisher.Add("first")
	if err := multiPublisher.Publish(report.MakeReport()); err != nil {
		t.Error(err)
	}
	if want, have := 1, p.count; want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	multiPublisher.Add("second") // but factory returns same mockPublisher
	if err := multiPublisher.Publish(report.MakeReport()); err != nil {
		t.Error(err)
	}
	if want, have := 3, p.count; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type mockPublisher struct{ count int }

func (p *mockPublisher) Publish(report.Report) error { p.count++; return nil }
func (p *mockPublisher) Stop()                       {}
