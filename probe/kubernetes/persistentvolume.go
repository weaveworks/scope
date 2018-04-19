package kubernetes

import (
	apiv1 "k8s.io/api/core/v1"
)

// PersistentVolume represent kubernetes PersistentVolume interface
type PersistentVolume interface {
	Meta
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
