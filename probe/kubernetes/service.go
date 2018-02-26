package kubernetes

import (
	"fmt"

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

// human-readable version of a Kubernetes ServicePort
func servicePortString(p apiv1.ServicePort) string {
	if p.NodePort == 0 {
		return fmt.Sprintf("%d/%s", p.Port, p.Protocol)
	}
	return fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, p.Protocol)
}

func (s *service) GetNode() report.Node {
	latest := map[string]string{
		IP:   s.Spec.ClusterIP,
		Type: string(s.Spec.Type),
	}
	if s.Spec.LoadBalancerIP != "" {
		latest[PublicIP] = s.Spec.LoadBalancerIP
	}
	if len(s.Spec.Ports) != 0 {
		portStr := ""
		for _, p := range s.Spec.Ports {
			portStr = portStr + servicePortString(p) + ","
		}
		latest[Ports] = portStr[:len(portStr)-1]
	}
	return s.MetaNode(report.MakeServiceNodeID(s.UID())).WithLatests(latest)
}

func (s *service) ClusterIP() string {
	return s.Spec.ClusterIP
}
