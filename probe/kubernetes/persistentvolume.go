package kubernetes

import (
	"github.com/weaveworks/scope/report"
	apiv1 "k8s.io/api/core/v1"
)

// PersistentVolume represent kubernetes PersistentVolume interface
type PersistentVolume interface {
	Meta
	GetNode(probeID string) report.Node
}

// persistentVolume represents kubernetes persistent volume
type persistentVolume struct {
	*apiv1.PersistentVolume
	Meta
}

// NewPersistentVolume returns new persistentVolume type
func NewPersistentVolume(p *apiv1.PersistentVolume) PersistentVolume {
	return &persistentVolume{PersistentVolume: p, Meta: meta{p.ObjectMeta}}
}

// GetNode returns Persistent Volume as Node
func (p *persistentVolume) GetNode(probeID string) report.Node {
	return p.MetaNode(report.MakePersistentVolumeNodeID(p.UID())).WithLatests(map[string]string{
		report.ControlProbeID: probeID,
		NodeType:              "Persistent Volume",
		VolumeClaim:           p.Spec.ClaimRef.Name,
		StorageClassName:      p.Spec.StorageClassName,
		Status:                string(p.Status.Phase),
		AccessModes:           string(p.Spec.AccessModes[0]),
		ReclaimPolicy:         string(p.Spec.PersistentVolumeReclaimPolicy),
		Message:               p.Status.Message,
	})
}
