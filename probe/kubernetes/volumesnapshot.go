package kubernetes

import (
	snapshotv1 "github.com/openebs/k8s-snapshot-client/snapshot/pkg/apis/volumesnapshot/v1"
	"github.com/weaveworks/scope/report"
)

// SnapshotPVName is the label key which provides PV name
const (
	SnapshotPVName = "SnapshotMetadata-PVName"
)

// Capacity is the annotation key which provides the storage size
const (
	Capacity = "capacity"
)

// VolumeSnapshot represent kubernetes VolumeSnapshot interface
type VolumeSnapshot interface {
	Meta
	GetNode(probeID string) report.Node
	GetVolumeName() string
	GetCapacity() string
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

// GetVolumeName returns the PVC name for volume snapshot
func (p *volumeSnapshot) GetVolumeName() string {
	return p.Spec.PersistentVolumeClaimName
}

// GetCapacity returns the capacity of the source PVC stored in annotation
func (p *volumeSnapshot) GetCapacity() string {
	capacity := p.GetAnnotations()[Capacity]
	if capacity != "" {
		return capacity
	}
	return ""
}

// GetNode returns VolumeSnapshot as Node
func (p *volumeSnapshot) GetNode(probeID string) report.Node {
	return p.MetaNode(report.MakeVolumeSnapshotNodeID(p.UID())).WithLatests(map[string]string{
		report.ControlProbeID: probeID,
		NodeType:              "Volume Snapshot",
		VolumeClaim:           p.GetVolumeName(),
		SnapshotData:          p.Spec.SnapshotDataName,
		VolumeName:            p.GetLabels()[SnapshotPVName],
	}).WithLatestActiveControls(CloneVolumeSnapshot, DeleteVolumeSnapshot, Describe)
}
