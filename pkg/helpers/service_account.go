package helpers

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/mwlng/k8s_resources_sync/pkg/k8s_resources"
)

func LoadServiceAccountYamlFiles(rootDir string) []*corev1.ServiceAccount {
	accounts := []*corev1.ServiceAccount{}
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
				case *corev1.ServiceAccount:
					accounts = append(accounts, obj.(*corev1.ServiceAccount))
				}
			}
		}
		return nil
	})

	if err != nil {
		klog.Errorf("Error while reading YAML files. Err was: %s", err)
	}

	return accounts
}

func SyncServiceAccounts(kubeConfig *rest.Config, serviceAccounts []*corev1.ServiceAccount) []*corev1.ServiceAccount {
	klog.Infof("Syncing service accounts from cluster: %s, namespace: %s\n", kubeConfig.Host, corev1.NamespaceDefault)
	serviceAccount, err := k8s_resources.NewServiceAccount(kubeConfig, corev1.NamespaceDefault)
	if err != nil {
		panic(err)
	}

	synced_serviceAccounts := []*corev1.ServiceAccount{}
	for _, account := range serviceAccounts {
		src_serviceAccount, err := serviceAccount.GetServiceAccount(account.Name)
		if err != nil {
			klog.Errorf("Failed to get service account: %s. Err was: %s", account.Name, err)
			continue
		}

		if src_serviceAccount != nil {
			synced_serviceAccounts = append(synced_serviceAccounts, src_serviceAccount)

		}
	}

	return synced_serviceAccounts
}

func PrintServiceAccounts(serviceAccounts []*corev1.Service) {
	for _, s := range serviceAccounts {
		result, _ := yaml.Marshal(s)
		fmt.Printf("%s\n", string(result))
	}
}

func ApplyServiceAccounts(kubeConfig *rest.Config, serviceAccounts []*corev1.ServiceAccount) {
	serviceAccount, err := k8s_resources.NewServiceAccount(kubeConfig, corev1.NamespaceDefault)
	if err != nil {
		panic(err)
	}

	for _, account := range serviceAccounts {
		klog.Infof("Applying service: %s ...", account.Name)
		err := serviceAccount.ApplyServiceAccount(account)
		if err != nil {
			klog.Errorf("Failed to apply service. Err was: %s", err)
			continue
		}
		klog.Infoln("Done.")
	}
}
