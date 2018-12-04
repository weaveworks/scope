package render

import (
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

// KubernetesVolumesRenderer is a Renderer which combines all Kubernetes
// volumes components such as stateful Pods, Persistent Volume, Persistent Volume Claim, Storage Class.
var KubernetesVolumesRenderer = MakeReduce(
	VolumesRenderer,
	PodToVolumeRenderer,
	PVCToStorageClassRenderer,
)

// VolumesRenderer is a Renderer which produces a renderable kubernetes PV & PVC
// graph by merging the pods graph and the Persistent Volume topology.
var VolumesRenderer = volumesRenderer{}

// volumesRenderer is a Renderer to render PV & PVC nodes.
type volumesRenderer struct{}

// Render renders PV & PVC nodes along with adjacency
func (v volumesRenderer) Render(rpt report.Report) Nodes {
	nodes := make(report.Nodes)
	for id, n := range rpt.PersistentVolumeClaim.Nodes {
		volume, _ := n.Latest.Lookup(kubernetes.VolumeName)
		for pvNodeID, p := range rpt.PersistentVolume.Nodes {
			volumeName, _ := p.Latest.Lookup(kubernetes.Name)
			if volume == volumeName {
				n.Adjacency = n.Adjacency.Add(p.ID)
				n.Children = n.Children.Add(p)
			}
			nodes[pvNodeID] = p
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
func (v podToVolumesRenderer) Render(rpt report.Report) Nodes {
	nodes := make(report.Nodes)
	for podID, podNode := range rpt.Pod.Nodes {
		ClaimName, _ := podNode.Latest.Lookup(kubernetes.VolumeClaim)
		for _, pvcNode := range rpt.PersistentVolumeClaim.Nodes {
			pvcName, _ := pvcNode.Latest.Lookup(kubernetes.Name)
			if pvcName == ClaimName {
				podNode.Adjacency = podNode.Adjacency.Add(pvcNode.ID)
				podNode.Children = podNode.Children.Add(pvcNode)
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
func (v pvcToStorageClassRenderer) Render(rpt report.Report) Nodes {
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
