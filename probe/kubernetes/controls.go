package kubernetes

import (
	"io"
	"io/ioutil"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/report"
)

// Control IDs used by the kubernetes integration.
const (
	GetLogs = "kubernetes_get_logs"
)

// GetLogs is the control to get the logs for a kubernetes pod
func (r *Reporter) GetLogs(req xfer.Request) xfer.Response {
	namespaceID, podID, ok := report.ParsePodNodeID(req.NodeID)
	if !ok {
		return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
	}

	readCloser, err := r.client.GetLogs(namespaceID, podID)
	if err != nil {
		return xfer.ResponseError(err)
	}

	readWriter := struct {
		io.Reader
		io.Writer
	}{
		readCloser,
		ioutil.Discard,
	}
	id, pipe, err := controls.NewPipeFromEnds(nil, readWriter, r.pipes, req.AppID)
	if err != nil {
		return xfer.ResponseError(err)
	}
	pipe.OnClose(func() {
		readCloser.Close()
	})
	return xfer.Response{
		Pipe: id,
	}
}

func (r *Reporter) registerControls() {
	controls.Register(GetLogs, r.GetLogs)
}

func (r *Reporter) deregisterControls() {
	controls.Rm(GetLogs)
}
