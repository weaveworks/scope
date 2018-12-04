package render

import (
	"context"

	"github.com/weaveworks/scope/report"
)

// TopologySelector selects a single topology from a report.
// NB it is also a Renderer!
type TopologySelector string

// Render implements Renderer
func (t TopologySelector) Render(ctx context.Context, r report.Report) Nodes {
	topology, _ := r.Topology(string(t))
	return Nodes{Nodes: topology.Nodes}
}

// The topology selectors implement a Renderer which fetch the nodes from the
// various report topologies.
var (
	SelectEndpoint              = TopologySelector(report.Endpoint)
	SelectProcess               = TopologySelector(report.Process)
	SelectContainer             = TopologySelector(report.Container)
	SelectContainerImage        = TopologySelector(report.ContainerImage)
	SelectHost                  = TopologySelector(report.Host)
	SelectPod                   = TopologySelector(report.Pod)
	SelectService               = TopologySelector(report.Service)
	SelectDeployment            = TopologySelector(report.Deployment)
	SelectDaemonSet             = TopologySelector(report.DaemonSet)
	SelectStatefulSet           = TopologySelector(report.StatefulSet)
	SelectCronJob               = TopologySelector(report.CronJob)
	SelectECSTask               = TopologySelector(report.ECSTask)
	SelectECSService            = TopologySelector(report.ECSService)
	SelectSwarmService          = TopologySelector(report.SwarmService)
	SelectOverlay               = TopologySelector(report.Overlay)
	SelectPersistentVolume      = TopologySelector(report.PersistentVolume)
	SelectPersistentVolumeClaim = TopologySelector(report.PersistentVolumeClaim)
	SelectStorageClass          = TopologySelector(report.StorageClass)
	SelectVolumeSnapshot        = TopologySelector(report.VolumeSnapshot)
	SelectVolumeSnapshotData    = TopologySelector(report.VolumeSnapshotData)
)
