package k8s_resources

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Service struct {
	client typedv1.ServiceInterface
}

func NewService(config *rest.Config, namespace string) (*Service, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Service{
		client: clientset.CoreV1().Services(namespace),
	}, nil
}

func (s *Service) ListServices() (*corev1.ServiceList, error) {
	list, err := s.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (s *Service) GetService(name string) (*corev1.Service, error) {
	service, err := s.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (s *Service) CreateService(service *corev1.Service) error {
	_, err := s.client.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateService(service *corev1.Service) error {
	_, err := s.client.Update(context.TODO(), service, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ApplyService(service *corev1.Service) error {
	result, _ := s.GetService(service.Name)
	if result != nil {
		//version, _ := strconv.ParseInt(result.GetResourceVersion(), 10, 32)
		//service.SetResourceVersion(fmt.Sprintf("%d", (version + 1)))
		//service.Spec.ClusterIP = result.Spec.ClusterIP
		result.SetAnnotations(service.GetAnnotations())
		err := s.UpdateService(result)
		if err != nil {
			return err
		}
	} else {
		err := s.CreateService(service)
		if err != nil {
			return err
		}
	}

	return nil
}

/* Experimental
func (s *Service) ApplyService(service *corev1.Service) error {
	serviceApplyConfig, err := v1.ExtractService(service, "k8s_resource_sync")
	if err != nil {
		return err
	}

	_, err = s.client.Apply(context.TODO(), serviceApplyConfig, metav1.ApplyOptions{})
	if err != nil {
		return err
	}

	return nil
}
*/
