/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned/fake"
)

type fakeRemoteExecutor struct {
	method  string
	url     *url.URL
	execErr error
}

func (f *fakeRemoteExecutor) Execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	f.method = method
	f.url = url
	return f.execErr
}

func TestPodAndContainer(t *testing.T) {
	tests := []struct {
		args              []string
		argsLenAtDash     int
		p                 *ExecOptions
		name              string
		expectError       bool
		expectedPod       string
		expectedContainer string
		expectedArgs      []string
	}{
		{
			p:             &ExecOptions{},
			argsLenAtDash: -1,
			expectError:   true,
			name:          "empty",
		},
		{
			p:             &ExecOptions{PodName: "foo"},
			argsLenAtDash: -1,
			expectError:   true,
			name:          "no cmd",
		},
		{
			p:             &ExecOptions{PodName: "foo", ContainerName: "bar"},
			argsLenAtDash: -1,
			expectError:   true,
			name:          "no cmd, w/ container",
		},
		{
			p:             &ExecOptions{PodName: "foo"},
			args:          []string{"cmd"},
			argsLenAtDash: -1,
			expectedPod:   "foo",
			expectedArgs:  []string{"cmd"},
			name:          "pod in flags",
		},
		{
			p:             &ExecOptions{},
			args:          []string{"foo", "cmd"},
			argsLenAtDash: 0,
			expectError:   true,
			name:          "no pod, pod name is behind dash",
		},
		{
			p:             &ExecOptions{},
			args:          []string{"foo"},
			argsLenAtDash: -1,
			expectError:   true,
			name:          "no cmd, w/o flags",
		},
		{
			p:             &ExecOptions{},
			args:          []string{"foo", "cmd"},
			argsLenAtDash: -1,
			expectedPod:   "foo",
			expectedArgs:  []string{"cmd"},
			name:          "cmd, w/o flags",
		},
		{
			p:             &ExecOptions{},
			args:          []string{"foo", "cmd"},
			argsLenAtDash: 1,
			expectedPod:   "foo",
			expectedArgs:  []string{"cmd"},
			name:          "cmd, cmd is behind dash",
		},
		{
			p:                 &ExecOptions{ContainerName: "bar"},
			args:              []string{"foo", "cmd"},
			argsLenAtDash:     -1,
			expectedPod:       "foo",
			expectedContainer: "bar",
			expectedArgs:      []string{"cmd"},
			name:              "cmd, container in flag",
		},
	}
	for _, test := range tests {
		f, tf, codec := NewAPIFactory()
		tf.Client = &fake.RESTClient{
			Codec:  codec,
			Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) { return nil, nil }),
		}
		tf.Namespace = "test"
		tf.ClientConfig = &restclient.Config{}

		cmd := &cobra.Command{}
		options := test.p
		err := options.Complete(f, cmd, test.args, test.argsLenAtDash)
		if test.expectError && err == nil {
			t.Errorf("unexpected non-error (%s)", test.name)
		}
		if !test.expectError && err != nil {
			t.Errorf("unexpected error: %v (%s)", err, test.name)
		}
		if err != nil {
			continue
		}
		if options.PodName != test.expectedPod {
			t.Errorf("expected: %s, got: %s (%s)", test.expectedPod, options.PodName, test.name)
		}
		if options.ContainerName != test.expectedContainer {
			t.Errorf("expected: %s, got: %s (%s)", test.expectedContainer, options.ContainerName, test.name)
		}
		if !reflect.DeepEqual(test.expectedArgs, options.Command) {
			t.Errorf("expected: %v, got %v (%s)", test.expectedArgs, options.Command, test.name)
		}
	}
}

func TestExec(t *testing.T) {
	version := testapi.Default.GroupVersion().Version
	tests := []struct {
		name, version, podPath, execPath, container string
		pod                                         *api.Pod
		execErr                                     bool
	}{
		{
			name:     "pod exec",
			version:  version,
			podPath:  "/api/" + version + "/namespaces/test/pods/foo",
			execPath: "/api/" + version + "/namespaces/test/pods/foo/exec",
			pod:      execPod(),
		},
		{
			name:     "pod exec error",
			version:  version,
			podPath:  "/api/" + version + "/namespaces/test/pods/foo",
			execPath: "/api/" + version + "/namespaces/test/pods/foo/exec",
			pod:      execPod(),
			execErr:  true,
		},
	}
	for _, test := range tests {
		f, tf, codec := NewAPIFactory()
		tf.Client = &fake.RESTClient{
			Codec: codec,
			Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
				switch p, m := req.URL.Path, req.Method; {
				case p == test.podPath && m == "GET":
					body := objBody(codec, test.pod)
					return &http.Response{StatusCode: 200, Body: body}, nil
				default:
					// Ensures no GET is performed when deleting by name
					t.Errorf("%s: unexpected request: %s %#v\n%#v", test.name, req.Method, req.URL, req)
					return nil, fmt.Errorf("unexpected request")
				}
			}),
		}
		tf.Namespace = "test"
		tf.ClientConfig = &restclient.Config{ContentConfig: restclient.ContentConfig{GroupVersion: &unversioned.GroupVersion{Version: test.version}}}
		bufOut := bytes.NewBuffer([]byte{})
		bufErr := bytes.NewBuffer([]byte{})
		bufIn := bytes.NewBuffer([]byte{})
		ex := &fakeRemoteExecutor{}
		if test.execErr {
			ex.execErr = fmt.Errorf("exec error")
		}
		params := &ExecOptions{
			PodName:       "foo",
			ContainerName: "bar",
			In:            bufIn,
			Out:           bufOut,
			Err:           bufErr,
			Executor:      ex,
		}
		cmd := &cobra.Command{}
		args := []string{"test", "command"}
		if err := params.Complete(f, cmd, args, -1); err != nil {
			t.Fatal(err)
		}
		err := params.Run()
		if test.execErr && err != ex.execErr {
			t.Errorf("%s: Unexpected exec error: %v", test.name, err)
			continue
		}
		if !test.execErr && err != nil {
			t.Errorf("%s: Unexpected error: %v", test.name, err)
			continue
		}
		if test.execErr {
			continue
		}
		if ex.url.Path != test.execPath {
			t.Errorf("%s: Did not get expected path for exec request", test.name)
			continue
		}
		if ex.method != "POST" {
			t.Errorf("%s: Did not get method for exec request: %s", test.name, ex.method)
		}
	}
}

func execPod() *api.Pod {
	return &api.Pod{
		ObjectMeta: api.ObjectMeta{Name: "foo", Namespace: "test", ResourceVersion: "10"},
		Spec: api.PodSpec{
			RestartPolicy: api.RestartPolicyAlways,
			DNSPolicy:     api.DNSClusterFirst,
			Containers: []api.Container{
				{
					Name: "bar",
				},
			},
		},
		Status: api.PodStatus{
			Phase: api.PodRunning,
		},
	}
}
