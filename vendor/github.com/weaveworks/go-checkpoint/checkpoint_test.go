package checkpoint

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestCheck(t *testing.T) {
	expected := &CheckResponse{
		CurrentVersion:      "1.0.0",
		CurrentReleaseDate:  1460459932, // 2016-04-12 11:18:52
		CurrentDownloadURL:  "https://test-app.used-for-testing",
		CurrentChangelogURL: "https://test-app.used-for-testing",
		ProjectWebsite:      "https://test-app.used-for-testing",
		Outdated:            false,
		Alerts:              nil,
	}

	actual, err := Check(&CheckParams{
		Product: "test-app",
		Version: "1.0.0",
	})

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestCheck_flags(t *testing.T) {
	expected := &CheckResponse{
		CurrentVersion:      "1.0.0",
		CurrentReleaseDate:  1460459932, // 2016-04-12 11:18:52
		CurrentDownloadURL:  "https://test-app.used-for-testing",
		CurrentChangelogURL: "https://test-app.used-for-testing",
		ProjectWebsite:      "https://test-app.used-for-testing",
		Outdated:            false,
		Alerts:              nil,
	}

	actual, err := Check(&CheckParams{
		Product: "test-app",
		Version: "1.0.0",
		Flags: map[string]string{
			"flag1": "value1",
			"flag2": "value2",
		},
	})

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestCheck_disabled(t *testing.T) {
	os.Setenv("CHECKPOINT_DISABLE", "1")
	defer os.Setenv("CHECKPOINT_DISABLE", "")

	expected := &CheckResponse{}

	actual, err := Check(&CheckParams{
		Product: "test-app",
		Version: "1.0.0",
	})

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v to equal %+v", actual, expected)
	}
}

func TestCheck_cache(t *testing.T) {
	dir, err := ioutil.TempDir("", "checkpoint")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := &CheckResponse{
		CurrentVersion:      "1.0.0",
		CurrentReleaseDate:  1460459932, // 2016-04-12 11:18:52
		CurrentDownloadURL:  "https://test-app.used-for-testing",
		CurrentChangelogURL: "https://test-app.used-for-testing",
		ProjectWebsite:      "https://test-app.used-for-testing",
		Outdated:            false,
		Alerts:              nil,
	}

	var actual *CheckResponse
	for i := 0; i < 5; i++ {
		var err error
		actual, err = Check(&CheckParams{
			Product:   "test-app",
			Version:   "1.0.0",
			CacheFile: filepath.Join(dir, "cache"),
		})
		if err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v %#v", actual, expected)
	}
}

func TestCheck_cacheNested(t *testing.T) {
	dir, err := ioutil.TempDir("", "checkpoint")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := &CheckResponse{
		CurrentVersion:      "1.0.0",
		CurrentReleaseDate:  1460459932, // 2016-04-12 11:18:52
		CurrentDownloadURL:  "https://test-app.used-for-testing",
		CurrentChangelogURL: "https://test-app.used-for-testing",
		ProjectWebsite:      "https://test-app.used-for-testing",
		Outdated:            false,
		Alerts:              nil,
	}

	var actual *CheckResponse
	for i := 0; i < 5; i++ {
		var err error
		actual, err = Check(&CheckParams{
			Product:   "test-app",
			Version:   "1.0.0",
			CacheFile: filepath.Join(dir, "nested", "cache"),
		})
		if err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestCheckInterval(t *testing.T) {
	expected := &CheckResponse{
		CurrentVersion:      "1.0.0",
		CurrentReleaseDate:  1460459932, // 2016-04-12 11:18:52
		CurrentDownloadURL:  "https://test-app.used-for-testing",
		CurrentChangelogURL: "https://test-app.used-for-testing",
		ProjectWebsite:      "https://test-app.used-for-testing",
		Outdated:            false,
		Alerts:              nil,
	}

	params := &CheckParams{
		Product: "test-app",
		Version: "1.0.0",
	}

	calledCh := make(chan struct{})
	checkFn := func(actual *CheckResponse, err error) {
		defer close(calledCh)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("bad: %#v", actual)
		}
	}

	st := CheckInterval(params, 500*time.Millisecond, checkFn)
	defer st.Stop()

	select {
	case <-calledCh:
	case <-time.After(time.Second):
		t.Fatalf("timeout")
	}
}

func TestCheckInterval_disabled(t *testing.T) {
	os.Setenv("CHECKPOINT_DISABLE", "1")
	defer os.Setenv("CHECKPOINT_DISABLE", "")

	params := &CheckParams{
		Product: "test-app",
		Version: "1.0.0",
	}

	calledCh := make(chan struct{})
	checkFn := func(actual *CheckResponse, err error) {
		defer close(calledCh)
	}

	st := CheckInterval(params, 500*time.Millisecond, checkFn)
	defer st.Stop()

	select {
	case <-calledCh:
		t.Fatal("expected callback to not invoke")
	case <-time.After(time.Second):
	}
}

func TestCheckInterval_immediate(t *testing.T) {
	expected := &CheckResponse{
		CurrentVersion:      "1.0.0",
		CurrentReleaseDate:  1460459932, // 2016-04-12 11:18:52
		CurrentDownloadURL:  "https://test-app.used-for-testing",
		CurrentChangelogURL: "https://test-app.used-for-testing",
		ProjectWebsite:      "https://test-app.used-for-testing",
		Outdated:            false,
		Alerts:              nil,
	}

	params := &CheckParams{
		Product: "test-app",
		Version: "1.0.0",
	}

	calledCh := make(chan struct{})
	checkFn := func(actual *CheckResponse, err error) {
		defer close(calledCh)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("bad: %#v", actual)
		}
	}

	st := CheckInterval(params, 500*time.Second, checkFn)
	defer st.Stop()

	select {
	case <-calledCh:
	case <-time.After(time.Second):
		t.Fatalf("timeout")
	}
}

func TestRandomStagger(t *testing.T) {
	intv := 24 * time.Hour
	min := 18 * time.Hour
	max := 30 * time.Hour
	for i := 0; i < 1000; i++ {
		out := randomStagger(intv)
		if out < min || out > max {
			t.Fatalf("bad: %v", out)
		}
	}
}
