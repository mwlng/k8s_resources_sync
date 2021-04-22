package k8s_resources

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ServiceAccount struct {
	client typedv1.ServiceAccountInterface
}

func NewServiceAccount(config *rest.Config, namespace string) (*ServiceAccount, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &ServiceAccount{
		client: clientset.CoreV1().ServiceAccounts(namespace),
	}, nil
}

func (s *ServiceAccount) ListServiceAccounts() (*corev1.ServiceAccountList, error) {
	list, err := s.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (s *ServiceAccount) GetServiceAccount(name string) (*corev1.ServiceAccount, error) {
	account, err := s.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *ServiceAccount) CreateServiceAccount(serviceAccount *corev1.ServiceAccount) error {
	_, err := s.client.Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceAccount) UpdateServiceAccount(serviceAccount *corev1.ServiceAccount) error {
	_, err := s.client.Update(context.TODO(), serviceAccount, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceAccount) ApplyServiceAccount(serviceAccount *corev1.ServiceAccount) error {
	result, _ := s.GetServiceAccount(serviceAccount.Name)
	if result != nil {
		serviceAccount.ObjectMeta.UID = ""
		err := s.UpdateServiceAccount(serviceAccount)
		if err != nil {
			return err
		}
	} else {
		serviceAccount.ResourceVersion = ""
		err := s.CreateServiceAccount(serviceAccount)
		if err != nil {
			return err
		}
	}

	return nil
}
