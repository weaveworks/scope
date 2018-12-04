package kubernetes

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	Name            = report.KubernetesName
	Namespace       = report.KubernetesNamespace
	Created         = report.KubernetesCreated
	LabelPrefix     = "kubernetes_labels_"
	VolumeClaimName = report.KubernetesVolumeClaim
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
	ObjectMeta metav1.ObjectMeta
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
	}).AddPrefixPropertyList(LabelPrefix, m.Labels())
}

type namespaceMeta struct {
	ObjectMeta metav1.ObjectMeta
}

func (m namespaceMeta) UID() string {
	return string(m.ObjectMeta.UID)
}

func (m namespaceMeta) Name() string {
	return m.ObjectMeta.Name
}

func (m namespaceMeta) Namespace() string {
	return m.ObjectMeta.Namespace
}

func (m namespaceMeta) Created() string {
	return m.ObjectMeta.CreationTimestamp.Format(time.RFC3339Nano)
}

func (m namespaceMeta) Labels() map[string]string {
	return m.ObjectMeta.Labels
}

// MetaNode gets the node metadata
// For namespaces, ObjectMeta.Namespace is not set
func (m namespaceMeta) MetaNode(id string) report.Node {
	return report.MakeNodeWith(id, map[string]string{
		Name:    m.Name(),
		Created: m.Created(),
	}).AddPrefixPropertyList(LabelPrefix, m.Labels())
}
