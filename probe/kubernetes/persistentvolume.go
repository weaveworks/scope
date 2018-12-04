package kubernetes

import (
	"github.com/weaveworks/scope/report"
	apiv1 "k8s.io/api/core/v1"
)

// PersistentVolume represent kubernetes PersistentVolume interface
type PersistentVolume interface {
	Meta
	GetNode() report.Node
	GetAccessMode() string
	GetVolume() string
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

// GetAccessMode returns access mode associated with PV
func (p *persistentVolume) GetAccessMode() string {
	var accessMode string

	if len(p.Spec.AccessModes) > 0 {
		// A volume can only be mounted using one access mode at a time,
		//even if it supports many.
		accessMode = string(p.Spec.AccessModes[0])
	}
	return accessMode
}

// GetVolume returns volume name
func (p *persistentVolume) GetVolume() string {
	var volume string
	if p.Spec.ClaimRef != nil {
		volume = p.Spec.ClaimRef.Name
	}
	return volume
}

// GetNode returns Persistent Volume as Node
func (p *persistentVolume) GetNode() report.Node {
	return p.MetaNode(report.MakePersistentVolumeNodeID(p.UID())).WithLatests(map[string]string{
		NodeType:         "Persistent Volume",
		VolumeClaim:      p.GetVolume(),
		StorageClassName: p.Spec.StorageClassName,
		Status:           string(p.Status.Phase),
		AccessModes:      p.GetAccessMode(),
	})
}
