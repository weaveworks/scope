package kubernetes

import (
	apiv1 "k8s.io/api/core/v1"

	"github.com/weaveworks/scope/report"
)

// PersistentVolume represent kubernetes PersistentVolume interface
type PersistentVolume interface {
	Meta
	GetNode() report.Node
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

// GetNode returns Persistent Volume Claim as Node
func (p *persistentVolume) GetNode() report.Node {
	var parents report.Sets
	latests := map[string]string{
		NodeType:         "Persistent Volume",
		Status:           string(p.Status.Phase),
		StorageClassName: p.Spec.StorageClassName,
		AccessModes:      string(p.Spec.AccessModes[0]),
	}
	if p.Spec.ClaimRef != nil && p.Spec.ClaimRef.Kind == "PersistentVolumeClaim" {
		latests[VolumeClaim] = p.Spec.ClaimRef.Name
		id := report.MakePersistentVolumeClaimNodeID(string(p.Spec.ClaimRef.UID))
		parents = parents.Add(report.PersistentVolumeClaim, report.MakeStringSet(id))
	}
	return p.MetaNode(report.MakePersistentVolumeNodeID(p.UID())).WithLatests(latests).WithParents(parents)
}
