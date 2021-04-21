package k8s_resources

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/client-go/applyconfigurations/core/v1"
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

func (d *Service) ListServices() (*corev1.ServiceList, error) {
	list, err := d.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (d *Service) GetService(name string) (*corev1.Service, error) {
	service, err := d.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (d *Service) CreateService(service *corev1.Service) error {
	_, err := d.client.Create(context.TODO(), service, metav1.CreateOptions{FieldManager: "k8k8s_resource_sync"})
	if err != nil {
		return err
	}

	return nil
}

func (d *Service) ApplyService(service *corev1.Service) error {
	serviceApplyConfig, err := v1.ExtractService(service, "k8s_resource_sync")
	if err != nil {
		return err
	}

	_, err = d.client.Apply(context.TODO(), serviceApplyConfig, metav1.ApplyOptions{})
	if err != nil {
		return err
	}

	return nil
}
