/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package remotecommand

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/kubelet/server/remotecommand"
	"k8s.io/kubernetes/pkg/types"
	"k8s.io/kubernetes/pkg/util/httpstream"
)

type fakeExecutor struct {
	t             *testing.T
	testName      string
	errorData     string
	stdoutData    string
	stderrData    string
	expectStdin   bool
	stdinReceived bytes.Buffer
	tty           bool
	messageCount  int
	command       []string
	exec          bool
}

func (ex *fakeExecutor) ExecInContainer(name string, uid types.UID, container string, cmd []string, in io.Reader, out, err io.WriteCloser, tty bool) error {
	return ex.run(name, uid, container, cmd, in, out, err, tty)
}

func (ex *fakeExecutor) AttachContainer(name string, uid types.UID, container string, in io.Reader, out, err io.WriteCloser, tty bool) error {
	return ex.run(name, uid, container, nil, in, out, err, tty)
}

func (ex *fakeExecutor) run(name string, uid types.UID, container string, cmd []string, in io.Reader, out, err io.WriteCloser, tty bool) error {
	ex.command = cmd
	ex.tty = tty

	if e, a := "pod", name; e != a {
		ex.t.Errorf("%s: pod: expected %q, got %q", ex.testName, e, a)
	}
	if e, a := "uid", uid; e != string(a) {
		ex.t.Errorf("%s: uid: expected %q, got %q", ex.testName, e, a)
	}
	if ex.exec {
		if e, a := "ls /", strings.Join(ex.command, " "); e != a {
			ex.t.Errorf("%s: command: expected %q, got %q", ex.testName, e, a)
		}
	} else {
		if len(ex.command) > 0 {
			ex.t.Errorf("%s: command: expected nothing, got %v", ex.testName, ex.command)
		}
	}

	if len(ex.errorData) > 0 {
		return errors.New(ex.errorData)
	}

	if len(ex.stdoutData) > 0 {
		for i := 0; i < ex.messageCount; i++ {
			fmt.Fprint(out, ex.stdoutData)
		}
	}

	if len(ex.stderrData) > 0 {
		for i := 0; i < ex.messageCount; i++ {
			fmt.Fprint(err, ex.stderrData)
		}
	}

	if ex.expectStdin {
		io.Copy(&ex.stdinReceived, in)
	}

	return nil
}

func fakeServer(t *testing.T, testName string, exec bool, stdinData, stdoutData, stderrData, errorData string, tty bool, messageCount int, serverProtocols []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		executor := &fakeExecutor{
			t:            t,
			testName:     testName,
			errorData:    errorData,
			stdoutData:   stdoutData,
			stderrData:   stderrData,
			expectStdin:  len(stdinData) > 0,
			tty:          tty,
			messageCount: messageCount,
			exec:         exec,
		}

		if exec {
			remotecommand.ServeExec(w, req, executor, "pod", "uid", "container", 0, 10*time.Second, serverProtocols)
		} else {
			remotecommand.ServeAttach(w, req, executor, "pod", "uid", "container", 0, 10*time.Second, serverProtocols)
		}

		if e, a := strings.Repeat(stdinData, messageCount), executor.stdinReceived.String(); e != a {
			t.Errorf("%s: stdin: expected %q, got %q", testName, e, a)
		}
	})
}

func TestStream(t *testing.T) {
	testCases := []struct {
		TestName        string
		Stdin           string
		Stdout          string
		Stderr          string
		Error           string
		Tty             bool
		MessageCount    int
		ClientProtocols []string
		ServerProtocols []string
	}{
		{
			TestName:        "error",
			Error:           "bail",
			Stdout:          "a",
			ClientProtocols: []string{remotecommand.StreamProtocolV2Name},
			ServerProtocols: []string{remotecommand.StreamProtocolV2Name},
		},
		{
			TestName:        "in/out/err",
			Stdin:           "a",
			Stdout:          "b",
			Stderr:          "c",
			MessageCount:    100,
			ClientProtocols: []string{remotecommand.StreamProtocolV2Name},
			ServerProtocols: []string{remotecommand.StreamProtocolV2Name},
		},
		{
			TestName:        "in/out/tty",
			Stdin:           "a",
			Stdout:          "b",
			Tty:             true,
			MessageCount:    100,
			ClientProtocols: []string{remotecommand.StreamProtocolV2Name},
			ServerProtocols: []string{remotecommand.StreamProtocolV2Name},
		},
		{
			// 1.0 kubectl, 1.0 kubelet
			TestName:        "unversioned client, unversioned server",
			Stdout:          "b",
			Stderr:          "c",
			MessageCount:    1,
			ClientProtocols: []string{},
			ServerProtocols: []string{},
		},
		{
			// 1.0 kubectl, 1.1+ kubelet
			TestName:        "unversioned client, versioned server",
			Stdout:          "b",
			Stderr:          "c",
			MessageCount:    1,
			ClientProtocols: []string{},
			ServerProtocols: []string{remotecommand.StreamProtocolV2Name, remotecommand.StreamProtocolV1Name},
		},
		{
			// 1.1+ kubectl, 1.0 kubelet
			TestName:        "versioned client, unversioned server",
			Stdout:          "b",
			Stderr:          "c",
			MessageCount:    1,
			ClientProtocols: []string{remotecommand.StreamProtocolV2Name, remotecommand.StreamProtocolV1Name},
			ServerProtocols: []string{},
		},
	}

	for _, testCase := range testCases {
		for _, exec := range []bool{true, false} {
			var name string
			if exec {
				name = testCase.TestName + " (exec)"
			} else {
				name = testCase.TestName + " (attach)"
			}
			var (
				streamIn             io.Reader
				streamOut, streamErr io.Writer
			)
			localOut := &bytes.Buffer{}
			localErr := &bytes.Buffer{}

			server := httptest.NewServer(fakeServer(t, name, exec, testCase.Stdin, testCase.Stdout, testCase.Stderr, testCase.Error, testCase.Tty, testCase.MessageCount, testCase.ServerProtocols))

			url, _ := url.ParseRequestURI(server.URL)
			c := restclient.NewRESTClient(url, "", restclient.ContentConfig{GroupVersion: &unversioned.GroupVersion{Group: "x"}}, -1, -1, nil)
			req := c.Post().Resource("testing")

			if exec {
				req.Param("command", "ls")
				req.Param("command", "/")
			}

			if len(testCase.Stdin) > 0 {
				req.Param(api.ExecStdinParam, "1")
				streamIn = strings.NewReader(strings.Repeat(testCase.Stdin, testCase.MessageCount))
			}

			if len(testCase.Stdout) > 0 {
				req.Param(api.ExecStdoutParam, "1")
				streamOut = localOut
			}

			if testCase.Tty {
				req.Param(api.ExecTTYParam, "1")
			} else if len(testCase.Stderr) > 0 {
				req.Param(api.ExecStderrParam, "1")
				streamErr = localErr
			}

			conf := &restclient.Config{
				Host: server.URL,
			}
			e, err := NewExecutor(conf, "POST", req.URL())
			if err != nil {
				t.Errorf("%s: unexpected error: %v", name, err)
				continue
			}
			err = e.Stream(testCase.ClientProtocols, streamIn, streamOut, streamErr, testCase.Tty)
			hasErr := err != nil

			if len(testCase.Error) > 0 {
				if !hasErr {
					t.Errorf("%s: expected an error", name)
				} else {
					if e, a := testCase.Error, err.Error(); !strings.Contains(a, e) {
						t.Errorf("%s: expected error stream read %q, got %q", name, e, a)
					}
				}

				// TODO: Uncomment when fix #19254
				// server.Close()
				continue
			}

			if hasErr {
				t.Errorf("%s: unexpected error: %v", name, err)
				// TODO: Uncomment when fix #19254
				// server.Close()
				continue
			}

			if len(testCase.Stdout) > 0 {
				if e, a := strings.Repeat(testCase.Stdout, testCase.MessageCount), localOut; e != a.String() {
					t.Errorf("%s: expected stdout data '%s', got '%s'", name, e, a)
				}
			}

			if testCase.Stderr != "" {
				if e, a := strings.Repeat(testCase.Stderr, testCase.MessageCount), localErr; e != a.String() {
					t.Errorf("%s: expected stderr data '%s', got '%s'", name, e, a)
				}
			}

			// TODO: Uncomment when fix #19254
			// server.Close()
		}
	}
}

type fakeUpgrader struct {
	req           *http.Request
	resp          *http.Response
	conn          httpstream.Connection
	err, connErr  error
	checkResponse bool

	t *testing.T
}

func (u *fakeUpgrader) RoundTrip(req *http.Request) (*http.Response, error) {
	u.req = req
	return u.resp, u.err
}

func (u *fakeUpgrader) NewConnection(resp *http.Response) (httpstream.Connection, error) {
	if u.checkResponse && u.resp != resp {
		u.t.Errorf("response objects passed did not match: %#v", resp)
	}
	return u.conn, u.connErr
}

type fakeConnection struct {
	httpstream.Connection
}

// Dial is the common functionality between any stream based upgrader, regardless of protocol.
// This method ensures that someone can use a generic stream executor without being dependent
// on the core Kube client config behavior.
func TestDial(t *testing.T) {
	upgrader := &fakeUpgrader{
		t:             t,
		checkResponse: true,
		conn:          &fakeConnection{},
		resp: &http.Response{
			StatusCode: http.StatusSwitchingProtocols,
			Body:       ioutil.NopCloser(&bytes.Buffer{}),
		},
	}
	var called bool
	testFn := func(rt http.RoundTripper) http.RoundTripper {
		if rt != upgrader {
			t.Fatalf("unexpected round tripper: %#v", rt)
		}
		called = true
		return rt
	}
	exec, err := NewStreamExecutor(upgrader, testFn, "POST", &url.URL{Host: "something.com", Scheme: "https"})
	if err != nil {
		t.Fatal(err)
	}
	conn, protocol, err := exec.Dial("protocol1")
	if err != nil {
		t.Fatal(err)
	}
	if conn != upgrader.conn {
		t.Errorf("unexpected connection: %#v", conn)
	}
	if !called {
		t.Errorf("wrapper not called")
	}
	_ = protocol
}
