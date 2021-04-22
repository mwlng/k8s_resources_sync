package k8s_resources

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	typedv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ClusterRoleBinding struct {
	client typedv1.ClusterRoleBindingInterface
}

func NewClusterRoleBinding(config *rest.Config) (*ClusterRoleBinding, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &ClusterRoleBinding{
		client: clientset.RbacV1().ClusterRoleBindings(),
	}, nil
}

func (crb *ClusterRoleBinding) ListClusterRoleBindings() (*rbacv1.ClusterRoleBindingList, error) {
	list, err := crb.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (crb *ClusterRoleBinding) GetClusterRoleBinding(name string) (*rbacv1.ClusterRoleBinding, error) {
	roleBinding, err := crb.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return roleBinding, nil
}

func (crb *ClusterRoleBinding) CreateClusterRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding) error {
	_, err := crb.client.Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (crb *ClusterRoleBinding) UpdateClusterRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding) error {
	_, err := crb.client.Update(context.TODO(), clusterRoleBinding, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (crb *ClusterRoleBinding) ApplyClusterRoleBinding(clusterRoleBinding *rbacv1.ClusterRoleBinding) error {
	result, _ := crb.GetClusterRoleBinding(clusterRoleBinding.Name)
	if result != nil {
		result.Subjects = clusterRoleBinding.Subjects
		result.RoleRef = clusterRoleBinding.RoleRef
		err := crb.UpdateClusterRoleBinding(result)
		if err != nil {
			return err
		}
	} else {
		err := crb.CreateClusterRoleBinding(clusterRoleBinding)
		if err != nil {
			return err
		}
	}

	return nil
}
