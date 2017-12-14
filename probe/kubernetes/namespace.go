package kubernetes

import (
	"github.com/weaveworks/scope/report"

	apiv1 "k8s.io/client-go/pkg/api/v1"
)

// NamespaceResource represents a Kubernetes namespace
// `Namespace` is already taken in meta.go
type NamespaceResource interface {
	Meta
	GetNode() report.Node
}

type namespace struct {
	ns *apiv1.Namespace
	Meta
}

// NewNamespace creates a new Namespace
func NewNamespace(ns *apiv1.Namespace) NamespaceResource {
	return &namespace{ns: ns, Meta: namespaceMeta{ns.ObjectMeta}}
}

func (ns *namespace) GetNode() report.Node {
	return ns.MetaNode(report.MakeNamespaceNodeID(ns.UID()))
}
