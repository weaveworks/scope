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
	GetLogs   = report.KubernetesGetLogs
	DeletePod = report.KubernetesDeletePod
	ScaleUp   = report.KubernetesScaleUp
	ScaleDown = report.KubernetesScaleDown
)

// GetLogs is the control to get the logs for a kubernetes pod
func (r *Reporter) GetLogs(req xfer.Request, namespaceID, podID string) xfer.Response {
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

func (r *Reporter) deletePod(req xfer.Request, namespaceID, podID string) xfer.Response {
	if err := r.client.DeletePod(namespaceID, podID); err != nil {
		return xfer.ResponseError(err)
	}
	return xfer.Response{
		RemovedNode: req.NodeID,
	}
}

// CapturePod is exported for testing
func (r *Reporter) CapturePod(f func(xfer.Request, string, string) xfer.Response) func(xfer.Request) xfer.Response {
	return func(req xfer.Request) xfer.Response {
		uid, ok := report.ParsePodNodeID(req.NodeID)
		if !ok {
			return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
		}
		// find pod by UID
		var pod Pod
		r.client.WalkPods(func(p Pod) error {
			if p.UID() == uid {
				pod = p
			}
			return nil
		})
		if pod == nil {
			return xfer.ResponseErrorf("Pod not found: %s", uid)
		}
		return f(req, pod.Namespace(), pod.Name())
	}
}

// CaptureResource is exported for testing
func (r *Reporter) CaptureResource(f func(xfer.Request, string, string, string) xfer.Response) func(xfer.Request) xfer.Response {
	return func(req xfer.Request) xfer.Response {
		var resource, uid string
		for _, parser := range []struct {
			res string
			f   func(string) (string, bool)
		}{
			{report.Deployment, report.ParseDeploymentNodeID},
			{report.ReplicaSet, report.ParseReplicaSetNodeID},
		} {
			if u, ok := parser.f(req.NodeID); ok {
				resource, uid = parser.res, u
				break
			}
		}
		if resource == "" {
			return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
		}

		switch resource {
		case report.Deployment:
			var deployment Deployment
			r.client.WalkDeployments(func(d Deployment) error {
				if d.UID() == uid {
					deployment = d
				}
				return nil
			})
			if deployment != nil {
				return f(req, "deployment", deployment.Namespace(), deployment.Name())
			}
		case report.ReplicaSet:
			var replicaSet ReplicaSet
			var res string
			r.client.WalkReplicaSets(func(r ReplicaSet) error {
				if r.UID() == uid {
					replicaSet = r
					res = "replicaset"
				}
				return nil
			})
			if replicaSet == nil {
				r.client.WalkReplicationControllers(func(r ReplicationController) error {
					if r.UID() == uid {
						replicaSet = ReplicaSet(r)
						res = "replicationcontroller"
					}
					return nil
				})
			}
			if replicaSet != nil {
				return f(req, res, replicaSet.Namespace(), replicaSet.Name())
			}
		}
		return xfer.ResponseErrorf("%s not found: %s", resource, uid)
	}
}

// ScaleUp is the control to scale up a deployment
func (r *Reporter) ScaleUp(req xfer.Request, resource, namespace, id string) xfer.Response {
	return xfer.ResponseError(r.client.ScaleUp(resource, namespace, id))
}

// ScaleDown is the control to scale up a deployment
func (r *Reporter) ScaleDown(req xfer.Request, resource, namespace, id string) xfer.Response {
	return xfer.ResponseError(r.client.ScaleDown(resource, namespace, id))
}

func (r *Reporter) registerControls() {
	controls := map[string]xfer.ControlHandlerFunc{
		GetLogs:   r.CapturePod(r.GetLogs),
		DeletePod: r.CapturePod(r.deletePod),
		ScaleUp:   r.CaptureResource(r.ScaleUp),
		ScaleDown: r.CaptureResource(r.ScaleDown),
	}
	r.handlerRegistry.Batch(nil, controls)
}

func (r *Reporter) deregisterControls() {
	controls := []string{
		GetLogs,
		DeletePod,
		ScaleUp,
		ScaleDown,
	}
	r.handlerRegistry.Batch(controls, nil)
}
