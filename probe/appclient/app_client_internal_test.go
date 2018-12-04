package appclient

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/handlers"
	"github.com/ugorji/go/codec"
	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/report"
)

func dummyServer(t *testing.T, expectedToken, expectedID string, expectedVersion string, expectedReport report.Report, done chan struct{}) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if have := r.Header.Get("Authorization"); fmt.Sprintf("Scope-Probe token=%s", expectedToken) != have {
			t.Errorf("want %q, have %q", expectedToken, have)
		}

		if have := r.Header.Get(xfer.ScopeProbeIDHeader); expectedID != have {
			t.Errorf("want %q, have %q", expectedID, have)
		}

		if have := r.Header.Get(xfer.ScopeProbeVersionHeader); expectedVersion != have {
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

		decoder := codec.NewDecoder(reader, &codec.MsgpackHandle{})
		if err := decoder.Decode(&have); err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(expectedReport, have) {
			t.Error(test.Diff(expectedReport, have))
			return
		}
		w.WriteHeader(http.StatusOK)
		done <- struct{}{}
	})

	return httptest.NewServer(handlers.CompressHandler(handler))
}

func TestAppClientPublish(t *testing.T) {
	var (
		token   = "abcdefg"
		id      = "1234567"
		version = "0.18"
		rpt     = report.MakeReport()
		done    = make(chan struct{}, 10)
	)

	// marshalling->unmarshaling is not idempotent due to `json:"omitempty"`
	// tags, transforming empty slices into nils. So, we make DeepEqual
	// happy by setting empty `json:"omitempty"` entries to nil
	rpt.WalkTopologies(func(to *report.Topology) {
		*to = report.MakeTopology()
		to.Controls = nil
	})

	s := dummyServer(t, token, id, version, rpt, done)
	defer s.Close()

	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	pc := ProbeConfig{
		Token:        token,
		ProbeVersion: version,
		ProbeID:      id,
		Insecure:     false,
	}

	url, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewAppClient(pc, u.Host, *url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()

	// First few reports might be dropped as the client is spinning up.
	for i := 0; i < 10; i++ {
		buf, _ := rpt.WriteBinary()
		if err := p.Publish(buf, false); err != nil {
			t.Error(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestAppClientDetails(t *testing.T) {
	var (
		id      = "foobarbaz"
		version = "imalittleteapot"
		want    = xfer.Details{ID: id, Version: version}
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoder := codec.NewEncoder(w, &codec.JsonHandle{})
		if err := encoder.Encode(want); err != nil {
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
	url, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewAppClient(pc, u.Host, *url, nil)
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

// Make sure Stopping a client works even if the connection or the remote app
// gets stuck for whatever reason.
// See https://github.com/weaveworks/scope/issues/1576
func TestStop(t *testing.T) {
	var (
		rpt            = report.MakeReport()
		stopHanging    = make(chan struct{})
		receivedReport = make(chan struct{})
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(receivedReport)
		<-stopHanging
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

	url, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewAppClient(pc, u.Host, *url, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the app received our report and is stuck
	for done := false; !done; {
		select {
		case <-receivedReport:
			done = true
		default:
			buf, _ := rpt.WriteBinary()
			if err := p.Publish(buf, false); err != nil {
				t.Error(err)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Close the client while the app is stuck
	p.Stop()

	// Let the server go so that the test can end
	close(stopHanging)
}
