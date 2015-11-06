package docker_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

func TestControls(t *testing.T) {
	mdc := newMockClient()
	setupStubs(mdc, func() {
		registry, _ := docker.NewRegistry(10 * time.Second)
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
				NodeID:  report.MakeContainerNodeID("", "a1b2c3d4e5"),
			})
			if !reflect.DeepEqual(result, xfer.Response{
				Error: tc.result,
			}) {
				t.Error(result)
			}
		}
	})
}
