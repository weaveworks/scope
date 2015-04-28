package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type context int

const (
	oneProbe context = iota
	twoProbes
)

var (
	appPort    = 14030
	bridgePort = 14020
	probePort1 = 14010
	probePort2 = 14011
)

func withContext(t *testing.T, c context, tests ...func()) {
	var (
		publish = 10 * time.Millisecond
		batch   = 10 * publish
		wait    = 2 * batch
	)

	switch c {
	case oneProbe:
		probe := start(t, fmt.Sprintf(`fixprobe -listen=:%d -publish.interval=%s %s/test_single_report.json`, probePort1, publish, cwd()))
		defer stop(t, probe)

		time.Sleep(10 * time.Millisecond)
		app := start(t, fmt.Sprintf(`app -http.address=:%d -probes=localhost:%d -batch=%s`, appPort, probePort1, batch))
		defer stop(t, app)

	case twoProbes:
		probe1 := start(t, fmt.Sprintf(`fixprobe -listen=:%d -publish.interval=%s %s/test_single_report.json`, probePort1, publish, cwd()))
		defer stop(t, probe1)

		probe2 := start(t, fmt.Sprintf(`fixprobe -listen=:%d -publish.interval=%s %s/test_extra_report.json`, probePort2, publish, cwd()))
		defer stop(t, probe2)

		time.Sleep(10 * time.Millisecond)
		app := start(t, fmt.Sprintf(`app -http.address=:%d -probes=localhost:%d,localhost:%d -batch=%s`, appPort, probePort1, probePort2, batch))
		defer stop(t, app)

	default:
		t.Fatalf("bad context %v", c)
	}

	time.Sleep(wait)

	for _, f := range tests {
		f()
	}
}

func httpGet(t *testing.T, url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("httpGet: %s", err)
	}
	defer resp.Body.Close()
	if status := resp.StatusCode; status != 200 {
		t.Fatalf("httpGet got status %d, expected 200", status)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("httpGet: %s", err)
	}

	return buf
}
