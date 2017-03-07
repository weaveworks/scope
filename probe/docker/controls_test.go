package docker_test

import (
	"io"
	"reflect"
	"testing"
	"time"

	commonTest "github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestControls(t *testing.T) {
	mdc := newMockClient()
	setupStubs(mdc, func() {
		hr := controls.NewDefaultHandlerRegistry()
		registry, _ := docker.NewRegistry(10*time.Second, nil, false, "", hr, "")
		defer registry.Stop()

		for _, tc := range []struct{ command, result string }{
			{docker.StopContainer, "stopped"},
			{docker.StartContainer, "started"},
			{docker.RestartContainer, "restarted"},
			{docker.PauseContainer, "paused"},
			{docker.UnpauseContainer, "unpaused"},
		} {
			result := hr.HandleControlRequest(xfer.Request{
				Control: tc.command,
				NodeID:  report.MakeContainerNodeID("a1b2c3d4e5"),
			})
			if !reflect.DeepEqual(result, xfer.Response{
				Error: tc.result,
			}) {
				t.Error(result)
			}
		}
	})
}

type mockPipe struct{}

func (mockPipe) Ends() (io.ReadWriter, io.ReadWriter)                { return nil, nil }
func (mockPipe) CopyToWebsocket(io.ReadWriter, xfer.Websocket) error { return nil }
func (mockPipe) Close() error                                        { return nil }
func (mockPipe) Closed() bool                                        { return false }
func (mockPipe) OnClose(func())                                      {}

func TestPipes(t *testing.T) {
	oldNewPipe := controls.NewPipe
	defer func() { controls.NewPipe = oldNewPipe }()
	controls.NewPipe = func(_ controls.PipeClient, _ string) (string, xfer.Pipe, error) {
		return "pipeid", mockPipe{}, nil
	}

	mdc := newMockClient()
	setupStubs(mdc, func() {
		hr := controls.NewDefaultHandlerRegistry()
		registry, _ := docker.NewRegistry(10*time.Second, nil, false, "", hr, "")
		defer registry.Stop()

		test.Poll(t, 100*time.Millisecond, true, func() interface{} {
			_, ok := registry.GetContainer("ping")
			return ok
		})

		for _, want := range []struct {
			control  string
			response xfer.Response
		}{
			{
				control: docker.AttachContainer,
				response: xfer.Response{
					Pipe:   "pipeid",
					RawTTY: true,
				},
			},

			{
				control: docker.ExecContainer,
				response: xfer.Response{
					Pipe:             "pipeid",
					RawTTY:           true,
					ResizeTTYControl: docker.ResizeExecTTY,
				},
			},
		} {
			result := hr.HandleControlRequest(xfer.Request{
				Control: want.control,
				NodeID:  report.MakeContainerNodeID("ping"),
			})
			if !reflect.DeepEqual(result, want.response) {
				t.Errorf("diff %s: %s", want.control, commonTest.Diff(want, result))
			}
		}
	})
}
