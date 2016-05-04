package kubernetes

import (
	"github.com/weaveworks/scope/report"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
)

// These constants are keys used in node metadata
const (
	ServiceID          = "kubernetes_service_id"
	ServiceName        = "kubernetes_service_name"
	ServiceCreated     = "kubernetes_service_created"
	ServiceIP          = "kubernetes_service_ip"
	ServicePublicIP    = "kubernetes_service_public_ip"
	ServiceLabelPrefix = "kubernetes_service_label_"
)

// Service represents a Kubernetes service
type Service interface {
	Meta
	GetNode() report.Node
	Selector() labels.Selector
}

type service struct {
	*api.Service
	Meta
}

// NewService creates a new Service
func NewService(s *api.Service) Service {
	return &service{Service: s, Meta: meta{s.ObjectMeta}}
}

func (s *service) Selector() labels.Selector {
	if s.Spec.Selector == nil {
		return labels.Nothing()
	}
	return labels.SelectorFromSet(labels.Set(s.Spec.Selector))
}

func (s *service) GetNode() report.Node {
	latest := map[string]string{
		ServiceID:      s.ID(),
		ServiceName:    s.Name(),
		ServiceCreated: s.Created(),
		Namespace:      s.Namespace(),
		ServiceIP:      s.Spec.ClusterIP,
	}
	if s.Spec.LoadBalancerIP != "" {
		latest[ServicePublicIP] = s.Spec.LoadBalancerIP
	}
	return report.MakeNodeWith(
		report.MakeServiceNodeID(s.UID()),
		latest,
	).
		AddTable(ServiceLabelPrefix, s.ObjectMeta.Labels)
}
