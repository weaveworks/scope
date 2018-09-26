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
	CreateVolumeSnapshot = report.KubernetesCreateVolumeSnapshot
	GetLogs              = report.KubernetesGetLogs
	DeletePod            = report.KubernetesDeletePod
	ScaleUp              = report.KubernetesScaleUp
	ScaleDown            = report.KubernetesScaleDown
)

// GetLogs is the control to get the logs for a kubernetes pod
func (r *Reporter) GetLogs(req xfer.Request, namespaceID, podID string, containerNames []string) xfer.Response {
	readCloser, err := r.client.GetLogs(namespaceID, podID, containerNames)
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

func (r *Reporter) createVolumeSnapshot(req xfer.Request, namespaceID, persistentVolumeClaimID, capacity string) xfer.Response {
	err := r.client.CreateVolumeSnapshot(namespaceID, persistentVolumeClaimID, capacity)
	if err != nil {
		return xfer.ResponseError(err)
	}
	return xfer.Response{}
}

func (r *Reporter) deletePod(req xfer.Request, namespaceID, podID string, _ []string) xfer.Response {
	if err := r.client.DeletePod(namespaceID, podID); err != nil {
		return xfer.ResponseError(err)
	}
	return xfer.Response{
		RemovedNode: req.NodeID,
	}
}

// CapturePod is exported for testing
func (r *Reporter) CapturePod(f func(xfer.Request, string, string, []string) xfer.Response) func(xfer.Request) xfer.Response {
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
		return f(req, pod.Namespace(), pod.Name(), pod.ContainerNames())
	}
}

// CaptureDeployment is exported for testing
func (r *Reporter) CaptureDeployment(f func(xfer.Request, string, string) xfer.Response) func(xfer.Request) xfer.Response {
	return func(req xfer.Request) xfer.Response {
		uid, ok := report.ParseDeploymentNodeID(req.NodeID)
		if !ok {
			return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
		}
		var deployment Deployment
		r.client.WalkDeployments(func(d Deployment) error {
			if d.UID() == uid {
				deployment = d
			}
			return nil
		})
		if deployment == nil {
			return xfer.ResponseErrorf("Deployment not found: %s", uid)
		}
		return f(req, deployment.Namespace(), deployment.Name())
	}
}

// CapturePersistentVolumeClaim will return name, namespace and capacity of PVC
func (r *Reporter) CapturePersistentVolumeClaim(f func(xfer.Request, string, string, string) xfer.Response) func(xfer.Request) xfer.Response {
	return func(req xfer.Request) xfer.Response {
		uid, ok := report.ParsePersistentVolumeClaimNodeID(req.NodeID)
		if !ok {
			return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
		}
		// find persistentVolumeClaim by UID
		var persistentVolumeClaim PersistentVolumeClaim
		r.client.WalkPersistentVolumeClaims(func(p PersistentVolumeClaim) error {
			if p.UID() == uid {
				persistentVolumeClaim = p
			}
			return nil
		})
		if persistentVolumeClaim == nil {
			return xfer.ResponseErrorf("Persistent volume claim not found: %s", uid)
		}
		return f(req, persistentVolumeClaim.Namespace(), persistentVolumeClaim.Name(), persistentVolumeClaim.GetCapacity())
	}
}

// ScaleUp is the control to scale up a deployment
func (r *Reporter) ScaleUp(req xfer.Request, namespace, id string) xfer.Response {
	return xfer.ResponseError(r.client.ScaleUp(report.Deployment, namespace, id))
}

// ScaleDown is the control to scale up a deployment
func (r *Reporter) ScaleDown(req xfer.Request, namespace, id string) xfer.Response {
	return xfer.ResponseError(r.client.ScaleDown(report.Deployment, namespace, id))
}

func (r *Reporter) registerControls() {
	controls := map[string]xfer.ControlHandlerFunc{
		CreateVolumeSnapshot: r.CapturePersistentVolumeClaim(r.createVolumeSnapshot),
		GetLogs:              r.CapturePod(r.GetLogs),
		DeletePod:            r.CapturePod(r.deletePod),
		ScaleUp:              r.CaptureDeployment(r.ScaleUp),
		ScaleDown:            r.CaptureDeployment(r.ScaleDown),
	}
	r.handlerRegistry.Batch(nil, controls)
}

func (r *Reporter) deregisterControls() {
	controls := []string{
		CreateVolumeSnapshot,
		GetLogs,
		DeletePod,
		ScaleUp,
		ScaleDown,
	}
	r.handlerRegistry.Batch(controls, nil)
}
