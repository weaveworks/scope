package render

import (
	"context"
	"strings"

	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

// KubernetesVolumesRenderer is a Renderer which combines all Kubernetes
// volumes components such as stateful Pods, Persistent Volume, Persistent Volume Claim, Storage Class.
var KubernetesVolumesRenderer = MakeReduce(
	VolumesRenderer,
	PodToVolumeRenderer,
	PVCToStorageClassRenderer,
	PVToSnapshotRenderer,
	VolumeSnapshotRenderer,
)

// VolumesRenderer is a Renderer which produces a renderable kubernetes PV & PVC
// graph by merging the pods graph and the Persistent Volume topology.
var VolumesRenderer = volumesRenderer{}

// volumesRenderer is a Renderer to render PV & PVC nodes.
type volumesRenderer struct{}

// Render renders PV & PVC nodes along with adjacency
func (v volumesRenderer) Render(ctx context.Context, rpt report.Report) Nodes {
	nodes := make(report.Nodes)
	for id, n := range rpt.PersistentVolumeClaim.Nodes {
		volume, _ := n.Latest.Lookup(kubernetes.VolumeName)
		for _, p := range rpt.PersistentVolume.Nodes {
			volumeName, _ := p.Latest.Lookup(kubernetes.Name)
			if volume == volumeName {
				n.Adjacency = n.Adjacency.Add(p.ID)
				n.Children = n.Children.Add(p)
			}
		}
		nodes[id] = n
	}
	return Nodes{Nodes: nodes}
}

// PodToVolumeRenderer is a Renderer which produces a renderable kubernetes Pod
// graph by merging the pods graph and the Persistent Volume Claim topology.
// Pods having persistent volumes are rendered.
var PodToVolumeRenderer = podToVolumesRenderer{}

// VolumesRenderer is a Renderer to render Pods & PVCs.
type podToVolumesRenderer struct{}

// Render renders the Pod nodes having volumes adjacency.
func (v podToVolumesRenderer) Render(ctx context.Context, rpt report.Report) Nodes {
	nodes := make(report.Nodes)
	for podID, podNode := range rpt.Pod.Nodes {
		claimNames, found := podNode.Latest.Lookup(kubernetes.VolumeClaim)
		if !found {
			continue
		}
		podNamespace, _ := podNode.Latest.Lookup(kubernetes.Namespace)
		claimNameList := strings.Split(claimNames, report.ScopeDelim)
		for _, ClaimName := range claimNameList {
			for _, pvcNode := range rpt.PersistentVolumeClaim.Nodes {
				pvcName, _ := pvcNode.Latest.Lookup(kubernetes.Name)
				pvcNamespace, _ := pvcNode.Latest.Lookup(kubernetes.Namespace)
				if (pvcName == ClaimName) && (podNamespace == pvcNamespace) {
					podNode.Adjacency = podNode.Adjacency.Add(pvcNode.ID)
					podNode.Children = podNode.Children.Add(pvcNode)
					break
				}
			}
		}
		nodes[podID] = podNode
	}
	return Nodes{Nodes: nodes}
}

// PVCToStorageClassRenderer is a Renderer which produces a renderable kubernetes PVC
// & Storage class graph.
var PVCToStorageClassRenderer = pvcToStorageClassRenderer{}

// pvcToStorageClassRenderer is a Renderer to render PVC & StorageClass.
type pvcToStorageClassRenderer struct{}

// Render renders the PVC & Storage Class nodes with adjacency.
func (v pvcToStorageClassRenderer) Render(ctx context.Context, rpt report.Report) Nodes {
	nodes := make(report.Nodes)
	for scID, scNode := range rpt.StorageClass.Nodes {
		storageClass, _ := scNode.Latest.Lookup(kubernetes.Name)
		for _, pvcNode := range rpt.PersistentVolumeClaim.Nodes {
			storageClassName, _ := pvcNode.Latest.Lookup(kubernetes.StorageClassName)
			if storageClassName == storageClass {
				scNode.Adjacency = scNode.Adjacency.Add(pvcNode.ID)
				scNode.Children = scNode.Children.Add(pvcNode)
			}
		}
		nodes[scID] = scNode
	}
	return Nodes{Nodes: nodes}
}

//PVToSnapshotRenderer is a Renderer which produces a renderable kubernetes PV
var PVToSnapshotRenderer = pvToSnapshotRenderer{}

//pvToSnapshotRenderer is a Renderer to render PV & Snapshot.
type pvToSnapshotRenderer struct{}

//Render renders the PV & Snapshot nodes with adjacency.
func (v pvToSnapshotRenderer) Render(ctx context.Context, rpt report.Report) Nodes {
	nodes := make(report.Nodes)
	for pvNodeID, p := range rpt.PersistentVolume.Nodes {
		volumeName, _ := p.Latest.Lookup(kubernetes.Name)
		for _, volumeSnapshotNode := range rpt.VolumeSnapshot.Nodes {
			snapshotPVName, _ := volumeSnapshotNode.Latest.Lookup(kubernetes.VolumeName)
			if volumeName == snapshotPVName {
				p.Adjacency = p.Adjacency.Add(volumeSnapshotNode.ID)
				p.Children = p.Children.Add(volumeSnapshotNode)
			}
		}
		nodes[pvNodeID] = p
	}
	return Nodes{Nodes: nodes}
}

// VolumeSnapshotRenderer is a renderer which produces a renderable Kubernetes Volume Snapshot and Volume Snapshot Data
var VolumeSnapshotRenderer = volumeSnapshotRenderer{}

// volumeSnapshotRenderer is a render to volume snapshot & volume snapshot data
type volumeSnapshotRenderer struct{}

// Render renders the volumeSnapshots & volumeSnapshotData with adjacency
// It checks for the volumeSnapshotData name in volumeSnapshot, adjacency is created by matching the volumeSnapshotData name.
func (v volumeSnapshotRenderer) Render(ctx context.Context, rpt report.Report) Nodes {
	nodes := make(report.Nodes)
	for volumeSnapshotID, volumeSnapshotNode := range rpt.VolumeSnapshot.Nodes {
		snapshotData, _ := volumeSnapshotNode.Latest.Lookup(kubernetes.SnapshotData)
		for volumeSnapshotDataID, volumeSnapshotDataNode := range rpt.VolumeSnapshotData.Nodes {
			snapshotDataName, _ := volumeSnapshotDataNode.Latest.Lookup(kubernetes.Name)
			if snapshotDataName == snapshotData {
				volumeSnapshotNode.Adjacency = volumeSnapshotNode.Adjacency.Add(volumeSnapshotDataNode.ID)
				volumeSnapshotNode.Children = volumeSnapshotNode.Children.Add(volumeSnapshotDataNode)
			}
			nodes[volumeSnapshotDataID] = volumeSnapshotDataNode
		}
		nodes[volumeSnapshotID] = volumeSnapshotNode
	}
	return Nodes{Nodes: nodes}
}
