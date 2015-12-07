package xfer

import (
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/handlers"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

type publisherFunc func(io.Reader) error

func (p publisherFunc) Publish(r io.Reader) error {
	return p(r)
}

func (publisherFunc) Stop() {}

func dummyServer(t *testing.T, expectedToken, expectedID string, expectedReport report.Report, done chan struct{}) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if have := r.Header.Get("Authorization"); fmt.Sprintf("Scope-Probe token=%s", expectedToken) != have {
			t.Errorf("want %q, have %q", expectedToken, have)
		}

		if have := r.Header.Get(ScopeProbeIDHeader); expectedID != have {
			t.Errorf("want %q, have %q", expectedID, have)
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
		if !reflect.DeepEqual(expectedReport, have) {
			t.Error(test.Diff(expectedReport, have))
			return
		}
		w.WriteHeader(http.StatusOK)
		close(done)
	})

	return httptest.NewServer(handlers.CompressHandler(handler))
}

func TestAppClientPublishInternal(t *testing.T) {
	var (
		token = "abcdefg"
		id    = "1234567"
		rpt   = report.MakeReport()
		done  = make(chan struct{})
	)

	s := dummyServer(t, token, id, rpt, done)
	defer s.Close()

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	pc := ProbeConfig{
		Token:    token,
		ProbeID:  id,
		Insecure: false,
	}

	p, err := NewAppClient(pc, u.Host, s.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()
	rp := NewReportPublisher(publisherFunc(p.(*appClient).publish))
	if err := rp.Publish(rpt); err != nil {
		t.Error(err)
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestAppClientDetails(t *testing.T) {
	var (
		token = "abcdefg"
		id    = "1234567"
		rpt   = report.MakeReport()
		done  = make(chan struct{})
	)

	s := dummyServer(t, token, id, rpt, done)
	defer s.Close()

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	pc := ProbeConfig{
		Token:    token,
		ProbeID:  id,
		Insecure: false,
	}

	p, err := NewAppClient(pc, u.Host, s.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()

	// First few reports might be dropped as the client is spinning up.
	rp := NewReportPublisher(p)
	for i := 0; i < 3; i++ {
		if err := rp.Publish(rpt); err != nil {
			t.Error(err)
		}
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestAppClientPublish(t *testing.T) {
	var (
		id      = "foobarbaz"
		version = "imalittleteapot"
		want    = Details{ID: id, Version: version}
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(want); err != nil {
			t.Fatal(err)
		}
	})

	s := httptest.NewServer(handlers.CompressHandler(handler))
	defer s.Close()

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	pc := ProbeConfig{
		Token:    "",
		ProbeID:  "",
		Insecure: false,
	}
	p, err := NewAppClient(pc, u.Host, s.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()

	have, err := p.Details()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
		return
	}
}
