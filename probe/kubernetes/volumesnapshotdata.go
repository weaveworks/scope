package kubernetes

import (
	snapshotv1 "github.com/openebs/k8s-snapshot-client/snapshot/pkg/apis/volumesnapshot/v1"
	"github.com/weaveworks/scope/report"
)

// VolumeSnapshotData represent kubernetes VolumeSnapshotData interface
type VolumeSnapshotData interface {
	Meta
	GetNode(probeID string) report.Node
}

// volumeSnapshotData represents kubernetes volume snapshot data
type volumeSnapshotData struct {
	*snapshotv1.VolumeSnapshotData
	Meta
}

// NewVolumeSnapshotData returns new Volume Snapshot Data type
func NewVolumeSnapshotData(p *snapshotv1.VolumeSnapshotData) VolumeSnapshotData {
	return &volumeSnapshotData{VolumeSnapshotData: p, Meta: meta{p.ObjectMeta}}
}

// GetNode returns VolumeSnapshotData as Node
func (p *volumeSnapshotData) GetNode(probeID string) report.Node {
	return p.MetaNode(report.MakeVolumeSnapshotDataNodeID(p.UID())).WithLatests(map[string]string{
		report.ControlProbeID: probeID,
		NodeType:              "Volume Snapshot Data",
		VolumeName:            p.Spec.PersistentVolumeRef.Name,
		VolumeSnapshotName:    p.Spec.VolumeSnapshotRef.Name,
	})
}
