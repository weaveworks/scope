package plugins

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
	"testing"
	"time"

	exec_hook "github.com/weaveworks/scope/common/exec"
	fs_hook "github.com/weaveworks/scope/common/fs"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/exec"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/fs"
)

const pluginDir = "/etc/weave/scope/plugins"

var mockFS = fs.Dir("",
	fs.Dir("etc",
		fs.Dir("weave",
			fs.Dir("scope",
				fs.Dir("plugins",
					fs.File{
						FName: "notaplugin",
					},
					fs.File{
						FName: "plugin1",
						FMode: 0700,
					},
					fs.File{
						FName: "plugin2",
						FMode: 0777,
					},
				),
			),
		),
	),
)

func TestPluginDiscovery(t *testing.T) {
	fs_hook.Mock(mockFS)
	defer fs_hook.Restore()

	registry := NewPluginRegistry(pluginDir)
	defer registry.Stop()

	test.Poll(t, 100*time.Millisecond, []string{
		path.Join(pluginDir, "/plugin1"),
		path.Join(pluginDir, "/plugin2"),
	},
		func() interface{} {
			registry.Lock()
			defer registry.Unlock()
			result := []string{}
			for _, p := range registry.plugins {
				result = append(result, p.path)
			}
			return result
		})
}

func TestPluginRPC(t *testing.T) {
	stdoutr, stdoutw := io.Pipe()
	stdinr, stdinw := io.Pipe()
	exec_hook.Mock(func(name string, args ...string) exec_hook.Cmd {
		if name != "/foo/bar" {
			t.Fatal(name)
		}
		return exec.NewMockCmd(stdoutr, stdinw)
	})
	defer exec_hook.Restore()
	go func() {
		req, err := http.ReadRequest(bufio.NewReader(stdinr))
		if err != nil {
			t.Fatal(err)
		}
		if req.URL.Path != "/report" {
			t.Fatal(req.URL.Path)
		}
		buf := bytes.Buffer{}
		if err := json.NewEncoder(&buf).Encode(fixture.Report); err != nil {
			t.Fatal(err)
		}
		resp := http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(&buf),
		}
		resp.Write(stdoutw)
	}()

	plugin := plugin{
		path: "/foo/bar",
	}
	report, err := plugin.report()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(fixture.Report, report) {
		test.Diff(fixture.Report, report)
	}
}
