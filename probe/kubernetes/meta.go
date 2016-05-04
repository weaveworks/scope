package kubernetes

import (
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
)

// Meta represents a metadata information about a Kubernetes object
type Meta interface {
	UID() string
	ID() string
	Name() string
	Namespace() string
	Created() string
	Labels() labels.Labels
}

type meta struct {
	ObjectMeta api.ObjectMeta
}

func (m meta) UID() string {
	return string(m.ObjectMeta.UID)
}

func (m meta) ID() string {
	return m.ObjectMeta.Namespace + "/" + m.ObjectMeta.Name
}

func (m meta) Name() string {
	return m.ObjectMeta.Name
}

func (m meta) Namespace() string {
	return m.ObjectMeta.Namespace
}

func (m meta) Created() string {
	return m.ObjectMeta.CreationTimestamp.Format(time.RFC822)
}

func (m meta) Labels() labels.Labels {
	return labels.Set(m.ObjectMeta.Labels)
}
