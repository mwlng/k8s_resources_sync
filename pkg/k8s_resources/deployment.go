package k8s_resources

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Deployment struct {
	client v1.DeploymentInterface
}

func NewDeployment(config *rest.Config, namespace string) (*Deployment, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Deployment{
		client: clientset.AppsV1().Deployments(namespace),
	}, nil
}

func (d *Deployment) ListDeloyments() (*appsv1.DeploymentList, error) {
	list, err := d.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}
