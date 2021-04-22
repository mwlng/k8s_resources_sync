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

func LoadClusterRoleYamlFiles(rootDir string) []*rbacv1.ClusterRole {
	roles := []*rbacv1.ClusterRole{}
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
				case *rbacv1.ClusterRole:
					roles = append(roles, obj.(*rbacv1.ClusterRole))
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

func SyncClusterRoles(kubeConfig *rest.Config, clusterRoles []*rbacv1.ClusterRole) []*rbacv1.ClusterRole {
	klog.Infof("Syncing cluster roles from cluster: %s\n", kubeConfig.Host)
	clusterRole, err := k8s_resources.NewClusterRole(kubeConfig)
	if err != nil {
		panic(err)
	}

	synced_clusterRoles := []*rbacv1.ClusterRole{}
	for _, role := range clusterRoles {
		src_clusterRole, err := clusterRole.GetClusterRole(role.Name)
		if err != nil {
			klog.Errorf("Failed to get service: %s. Err was: %s", role.Name, err)
			continue
		}

		if src_clusterRole != nil {
			role.Rules = src_clusterRole.Rules

			synced_clusterRoles = append(synced_clusterRoles, role)
		}
	}

	return synced_clusterRoles
}

func PrintClusterRoles(clusterRoles []*rbacv1.ClusterRole) {
	for _, role := range clusterRoles {
		result, _ := yaml.Marshal(role)
		fmt.Printf("%s\n", string(result))
	}
}

func ApplyClusterRoles(kubeConfig *rest.Config, clusterRoles []*rbacv1.ClusterRole) {
	clusterRole, err := k8s_resources.NewClusterRole(kubeConfig)
	if err != nil {
		panic(err)
	}

	for _, role := range clusterRoles {
		klog.Infof("Applying cluster role: %s ...", role.Name)
		err := clusterRole.ApplyClusterRole(role)
		if err != nil {
			klog.Errorf("Failed to apply cluster role. Err was: %s", err)
			continue
		}
		klog.Infoln("Done.")
	}
}
