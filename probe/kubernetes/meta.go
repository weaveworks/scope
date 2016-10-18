package kubernetes

import (
	"time"

	"k8s.io/kubernetes/pkg/api"

	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	Name        = "kubernetes_name"
	Namespace   = "kubernetes_namespace"
	Created     = "kubernetes_created"
	LabelPrefix = "kubernetes_labels_"
)

// Meta represents a metadata information about a Kubernetes object
type Meta interface {
	UID() string
	Name() string
	Namespace() string
	Created() string
	Labels() map[string]string
	MetaNode(id string) report.Node
}

type meta struct {
	ObjectMeta api.ObjectMeta
}

func (m meta) UID() string {
	return string(m.ObjectMeta.UID)
}

func (m meta) Name() string {
	return m.ObjectMeta.Name
}

func (m meta) Namespace() string {
	return m.ObjectMeta.Namespace
}

func (m meta) Created() string {
	return m.ObjectMeta.CreationTimestamp.Format(time.RFC3339Nano)
}

func (m meta) Labels() map[string]string {
	return m.ObjectMeta.Labels
}

// MetaNode gets the node metadata
func (m meta) MetaNode(id string) report.Node {
	return report.MakeNodeWith(id, map[string]string{
		Name:      m.Name(),
		Namespace: m.Namespace(),
		Created:   m.Created(),
	}).AddTable(LabelPrefix, m.Labels())
}
