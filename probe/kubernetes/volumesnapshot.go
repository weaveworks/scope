package kubernetes

import (
	snapshotv1 "github.com/openebs/k8s-snapshot-client/snapshot/pkg/apis/volumesnapshot/v1"
	"github.com/weaveworks/scope/report"
)

// SnapshotPVName is the label key which provides PV name
const (
	SnapshotPVName = "SnapshotMetadata-PVName"
)

// VolumeSnapshot represent kubernetes VolumeSnapshot interface
type VolumeSnapshot interface {
	Meta
	GetNode(probeID string) report.Node
}

// volumeSnapshot represents kubernetes volume snapshots
type volumeSnapshot struct {
	*snapshotv1.VolumeSnapshot
	Meta
}

// NewVolumeSnapshot returns new Volume Snapshot type
func NewVolumeSnapshot(p *snapshotv1.VolumeSnapshot) VolumeSnapshot {
	return &volumeSnapshot{VolumeSnapshot: p, Meta: meta{p.ObjectMeta}}
}

// GetNode returns VolumeSnapshot as Node
func (p *volumeSnapshot) GetNode(probeID string) report.Node {
	return p.MetaNode(report.MakeVolumeSnapshotNodeID(p.UID())).WithLatests(map[string]string{
		report.ControlProbeID: probeID,
		NodeType:              "Volume Snapshot",
		VolumeClaim:           p.Spec.PersistentVolumeClaimName,
		SnapshotData:          p.Spec.SnapshotDataName,
		VolumeName:            p.GetLabels()[SnapshotPVName],
	})
}
