package k8s_resources

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/client-go/applyconfigurations/apps/v1"
	typedv1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Deployment struct {
	client typedv1.DeploymentInterface
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

func (d *Deployment) GetDeployment(name string) (*appsv1.Deployment, error) {
	deployment, err := d.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func (d *Deployment) CreateDeployment(deployment *appsv1.Deployment) error {
	_, err := d.client.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (d *Deployment) ApplyDeployment(deployment *appsv1.Deployment) error {
	deploymentApplyConfig, err := v1.ExtractDeployment(deployment, "k8s_resource_sync")
	if err != nil {
		return err
	}

	_, err = d.client.Apply(context.TODO(), deploymentApplyConfig, metav1.ApplyOptions{FieldManager: "k8_resources_sync"})
	if err != nil {
		return err
	}

	return nil
}
