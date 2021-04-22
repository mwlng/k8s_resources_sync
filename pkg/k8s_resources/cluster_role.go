package k8s_resources

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	typedv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ClusterRole struct {
	client typedv1.ClusterRoleInterface
}

func NewClusterRole(config *rest.Config) (*ClusterRole, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &ClusterRole{
		client: clientset.RbacV1().ClusterRoles(),
	}, nil
}

func (cr *ClusterRole) ListClusterRoles() (*rbacv1.ClusterRoleList, error) {
	list, err := cr.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (cr *ClusterRole) GetClusterRole(name string) (*rbacv1.ClusterRole, error) {
	role, err := cr.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (cr *ClusterRole) CreateClusterRole(clusterRole *rbacv1.ClusterRole) error {
	_, err := cr.client.Create(context.TODO(), clusterRole, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (cr *ClusterRole) UpdateClusterRole(clusterRole *rbacv1.ClusterRole) error {
	_, err := cr.client.Update(context.TODO(), clusterRole, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (cr *ClusterRole) ApplyClusterRole(clusterRole *rbacv1.ClusterRole) error {
	result, _ := cr.GetClusterRole(clusterRole.Name)
	if result != nil {
		result.Rules = clusterRole.Rules
		err := cr.UpdateClusterRole(result)
		if err != nil {
			return err
		}
	} else {
		err := cr.CreateClusterRole(clusterRole)
		if err != nil {
			return err
		}
	}

	return nil
}
