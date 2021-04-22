package helpers

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/mwlng/k8s_resources_sync/pkg/k8s_resources"
)

func LoadClusterRoleBindingYamlFiles(rootDir string) []*rbacv1.ClusterRoleBinding {
	roles := []*rbacv1.ClusterRoleBinding{}
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".yml" || ext == ".yaml" {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					klog.Errorf("Error while reading YAML file. Err was: %s", err)
					return err
				}

				decode := scheme.Codecs.UniversalDeserializer().Decode
				obj, _, err := decode([]byte(data), nil, nil)

				if err != nil {
					klog.Errorf("Error while decoding YAML file: %s. Err was: %s", path, err)
					return nil
				}

				switch obj.(type) {
				case *rbacv1.ClusterRoleBinding:
					roles = append(roles, obj.(*rbacv1.ClusterRoleBinding))
				}
			}
		}
		return nil
	})

	if err != nil {
		klog.Errorf("Error while reading YAML files. Err was: %s", err)
	}

	return roles
}

func SyncClusterRoleBindings(kubeConfig *rest.Config, clusterRoleBindings []*rbacv1.ClusterRoleBinding) []*rbacv1.ClusterRoleBinding {
	klog.Infof("Syncing cluster role bindings from cluster: %s\n", kubeConfig.Host)
	clusterRoleBinding, err := k8s_resources.NewClusterRoleBinding(kubeConfig)
	if err != nil {
		panic(err)
	}

	synced_clusterRoleBindings := []*rbacv1.ClusterRoleBinding{}
	for _, roleBinding := range clusterRoleBindings {
		src_clusterRoleBinding, err := clusterRoleBinding.GetClusterRoleBinding(roleBinding.Name)
		if err != nil {
			klog.Errorf("Failed to get service: %s. Err was: %s", roleBinding.Name, err)
			continue
		}

		if src_clusterRoleBinding != nil {
			roleBinding.Subjects = src_clusterRoleBinding.Subjects
			roleBinding.RoleRef = src_clusterRoleBinding.RoleRef
			synced_clusterRoleBindings = append(synced_clusterRoleBindings, roleBinding)
		}
	}

	return synced_clusterRoleBindings
}

func PrintClusterRoleBindings(clusterRoleBindings []*rbacv1.ClusterRoleBinding) {
	for _, roleBinding := range clusterRoleBindings {
		result, _ := yaml.Marshal(roleBinding)
		fmt.Printf("%s\n", string(result))
	}
}

func ApplyClusterRoleBindings(kubeConfig *rest.Config, clusterRoleBindings []*rbacv1.ClusterRoleBinding) {
	clusterRoleBinding, err := k8s_resources.NewClusterRoleBinding(kubeConfig)
	if err != nil {
		panic(err)
	}

	for _, roleBinding := range clusterRoleBindings {
		klog.Infof("Applying cluster role binding: %s ...", roleBinding.Name)
		err := clusterRoleBinding.ApplyClusterRoleBinding(roleBinding)
		if err != nil {
			klog.Errorf("Failed to apply cluster role binding. Err was: %s", err)
			continue
		}
		klog.Infoln("Done.")
	}
}
