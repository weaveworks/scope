package kubernetes

import (
	"io"
	"io/ioutil"
	"strconv"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/report"
)

// Control IDs used by the kubernetes integration.
const (
	GetLogs = "kubernetes_get_logs"
)

func (r *Reporter) getLogs(req xfer.Request) xfer.Response {
	namespaceID, podID, ok := report.ParsePodNodeID(req.NodeID)
	if !ok {
		return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
	}

	k8sReq := r.client.RESTClient().Get().
		Namespace(namespaceID).
		Name(podID).
		Resource("pods").
		SubResource("log").
		Param("follow", strconv.FormatBool(true)).
		Param("previous", strconv.FormatBool(false)).
		Param("timestamps", strconv.FormatBool(true))

	readCloser, err := k8sReq.Stream()
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
	controls.Register(GetLogs, r.getLogs)
}

func (r *Reporter) deregisterControls() {
	controls.Rm(GetLogs)
}
