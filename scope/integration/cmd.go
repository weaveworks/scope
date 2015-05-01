package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// cmdline is e.g. `fixprobe -publish.interval=10ms fixture.json`
func start(t *testing.T, cmdline string) *exec.Cmd {
	toks := strings.Split(cmdline, " ")
	if len(toks) <= 0 {
		t.Fatalf("bad cmdline %q", cmdline)
	}

	component, args := toks[0], toks[1:]

	relpath := fmt.Sprintf("../%s/%s", component, component)

	if _, err := os.Stat(relpath); err != nil {
		t.Fatalf("%s: %s", component, err)
	}

	cmd := &exec.Cmd{
		Dir:    filepath.Dir(relpath),
		Path:   filepath.Base(relpath),
		Args:   append([]string{filepath.Base(relpath)}, args...),
		Stdout: testWriter{t, component},
		Stderr: testWriter{t, component},
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("%s: Start: %s", component, err)
	}

	return cmd
}

func stop(t *testing.T, c *exec.Cmd) {
	done := make(chan struct{})

	go func() {
		defer close(done)

		if err := c.Process.Kill(); err != nil {
			t.Fatalf("%s: Kill: %s", filepath.Base(c.Path), err)
		}

		if _, err := c.Process.Wait(); err != nil {
			t.Fatalf("%s: Wait: %s", filepath.Base(c.Path), err)
		}
	}()

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatalf("timeout when trying to stop %s", filepath.Base(c.Path))
	}
}

type testWriter struct {
	*testing.T
	component string
}

func (w testWriter) Write(p []byte) (int, error) {
	w.T.Logf("<%10s> %s", w.component, p)
	return len(p), nil
}

func cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd
}
