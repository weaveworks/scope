package kubernetes

import (
	"time"

	"github.com/weaveworks/scope/report"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
)

// These constants are keys used in node metadata
const (
	ServiceID      = "kubernetes_service_id"
	ServiceName    = "kubernetes_service_name"
	ServiceCreated = "kubernetes_service_created"
)

// Service represents a Kubernetes service
type Service interface {
	ID() string
	Name() string
	Namespace() string
	GetNode() report.Node
	Selector() labels.Selector
}

type service struct {
	*api.Service
}

// NewService creates a new Service
func NewService(s *api.Service) Service {
	return &service{Service: s}
}

func (s *service) ID() string {
	return s.ObjectMeta.Namespace + "/" + s.ObjectMeta.Name
}

func (s *service) Name() string {
	return s.ObjectMeta.Name
}

func (s *service) Namespace() string {
	return s.ObjectMeta.Namespace
}

func (s *service) Selector() labels.Selector {
	return labels.SelectorFromSet(labels.Set(s.Spec.Selector))
}

func (s *service) GetNode() report.Node {
	return report.MakeNodeWith(map[string]string{
		ServiceID:      s.ID(),
		ServiceName:    s.Name(),
		ServiceCreated: s.ObjectMeta.CreationTimestamp.Format(time.RFC822),
		Namespace:      s.Namespace(),
	})
}
