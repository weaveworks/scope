package docker_test

import (
	"io"
	"reflect"
	"testing"
	"time"

	"$GITHUB_URI/common/xfer"
	"$GITHUB_URI/probe/controls"
	"$GITHUB_URI/probe/docker"
	"$GITHUB_URI/report"
	"$GITHUB_URI/test"
)

func TestControls(t *testing.T) {
	mdc := newMockClient()
	setupStubs(mdc, func() {
		registry, _ := docker.NewRegistry(10*time.Second, nil, false, "")
		defer registry.Stop()

		for _, tc := range []struct{ command, result string }{
			{docker.StopContainer, "stopped"},
			{docker.StartContainer, "started"},
			{docker.RestartContainer, "restarted"},
			{docker.PauseContainer, "paused"},
			{docker.UnpauseContainer, "unpaused"},
		} {
			result := controls.HandleControlRequest(xfer.Request{
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
		registry, _ := docker.NewRegistry(10*time.Second, nil, false, "")
		defer registry.Stop()

		test.Poll(t, 100*time.Millisecond, true, func() interface{} {
			_, ok := registry.GetContainer("ping")
			return ok
		})

		for _, tc := range []string{
			docker.AttachContainer,
			docker.ExecContainer,
		} {
			result := controls.HandleControlRequest(xfer.Request{
				Control: tc,
				NodeID:  report.MakeContainerNodeID("ping"),
			})
			want := xfer.Response{
				Pipe:   "pipeid",
				RawTTY: true,
			}
			if !reflect.DeepEqual(result, want) {
				t.Errorf("diff %s: %s", tc, test.Diff(want, result))
			}
		}
	})
}
