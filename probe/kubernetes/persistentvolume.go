package kubernetes

import (
	"reflect"

	"github.com/weaveworks/scope/report"
	apiv1 "k8s.io/api/core/v1"
)

// PersistentVolume represent kubernetes PersistentVolume interface
type PersistentVolume interface {
	Meta
	GetNode(probeID string) report.Node
	GetAccessMode() string
	GetVolume() string
	GetStorageDriver() string
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

// GetStorageDriver returns the backing driver of Persistent Volume
func (p *persistentVolume) GetStorageDriver() string {
	persistentVolumeSource := reflect.ValueOf(p.Spec.PersistentVolumeSource)

	// persistentVolumeSource will have exactly one field which won't be nil,
	// depending on the type of backing driver used to create the persistent volume
	// Iterate over the fields and return the non-nil field name
	for i := 0; i < persistentVolumeSource.NumField(); i++ {
		if !reflect.ValueOf(persistentVolumeSource.Field(i).Interface()).IsNil() {
			return persistentVolumeSource.Type().Field(i).Name
		}
	}
	return ""
}

// GetNode returns Persistent Volume as Node
func (p *persistentVolume) GetNode(probeID string) report.Node {
	latests := map[string]string{
		NodeType:              "Persistent Volume",
		VolumeClaim:           p.GetVolume(),
		StorageClassName:      p.Spec.StorageClassName,
		Status:                string(p.Status.Phase),
		AccessModes:           p.GetAccessMode(),
		report.ControlProbeID: probeID,
	}

	if p.GetStorageDriver() != "" {
		latests[StorageDriver] = p.GetStorageDriver()
	}

	return p.MetaNode(report.MakePersistentVolumeNodeID(p.UID())).
		WithLatests(latests).
		WithLatestActiveControls(Describe)
}
