package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"
)

func TestCgroupMapper(t *testing.T) {
	tmp := setupTmpFS(t, map[string]string{
		"/systemd/tasks":              "1\n2\n4911\n1000\n25156\n",
		"/systemd/notify_on_release":  "0\n",
		"/netscape/tasks":             "666\n4242\n",
		"/netscape/notify_on_release": "0\n",
		"/weirdfile":                  "",
	})
	defer os.RemoveAll(tmp)

	m := newCgroupMapper(tmp, 1*time.Second)
	for pid, want := range map[uint]string{
		111:   "",
		999:   "",
		4911:  "systemd",
		1:     "systemd", // first one in the file
		25156: "systemd", // last one in the tasks file
		4242:  "netscape",
	} {
		if have, _ := m.Map(pid); want != have {
			t.Errorf("%d: want %q, have %q", pid, want, have)
		}
	}
}

func setupTmpFS(t *testing.T, fs map[string]string) string {
	tmp, err := ioutil.TempDir(os.TempDir(), "scope-probe-test-cgroup-mapper")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("using TempDir %s", tmp)

	for file, content := range fs {
		dir := path.Dir(file)
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0777); err != nil {
			os.RemoveAll(tmp)
			t.Fatalf("MkdirAll: %v", err)
		}

		if err := ioutil.WriteFile(filepath.Join(tmp, file), []byte(content), 0655); err != nil {
			os.RemoveAll(tmp)
			t.Fatalf("WriteFile: %v", err)
		}
	}
	return tmp
}
