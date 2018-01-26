package kubernetes

import (
	"github.com/weaveworks/scope/report"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// These constants are keys used in node metadata
const (
	PublicIP = report.KubernetesPublicIP
)

// Service represents a Kubernetes service
type Service interface {
	Meta
	GetNode() report.Node
	Selector() labels.Selector
	ClusterIP() string
}

type service struct {
	*apiv1.Service
	Meta
}

// NewService creates a new Service
func NewService(s *apiv1.Service) Service {
	return &service{Service: s, Meta: meta{s.ObjectMeta}}
}

func (s *service) Selector() labels.Selector {
	if s.Spec.Selector == nil {
		return labels.Nothing()
	}
	return labels.SelectorFromSet(labels.Set(s.Spec.Selector))
}

func (s *service) GetNode() report.Node {
	latest := map[string]string{IP: s.Spec.ClusterIP}
	if s.Spec.LoadBalancerIP != "" {
		latest[PublicIP] = s.Spec.LoadBalancerIP
	}
	return s.MetaNode(report.MakeServiceNodeID(s.UID())).WithLatests(latest)
}

func (s *service) ClusterIP() string {
	return s.Spec.ClusterIP
}
