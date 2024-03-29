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

var (
	private_subnets map[string]string = map[string]string{
		"alpha": "subnet-092aa15246e226be2,subnet-0c85b6c43be91a809",
		//"qa":    "subnet-e61597c9,subnet-8d9016d0,subnet-e240a3ed",
		"qa":   "",
		"prod": "subnet-61ba8d4e,subnet-48665215,subnet-4066c14f",
		//"prod": "",
	}
)

func LoadServiceYamlFiles(rootDir string) []*corev1.Service {
	services := []*corev1.Service{}
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
				case *corev1.Service:
					services = append(services, obj.(*corev1.Service))
				}
			}
		}
		return nil
	})

	if err != nil {
		klog.Errorf("Error while reading YAML files. Err was: %s", err)
	}

	return services
}

func SyncServices(kubeConfig *rest.Config, services []*corev1.Service, environ string) []*corev1.Service {
	klog.Infof("Syncing services from cluster: %s, namespace: %s\n", kubeConfig.Host, corev1.NamespaceDefault)
	service, err := k8s_resources.NewService(kubeConfig, corev1.NamespaceDefault)
	if err != nil {
		panic(err)
	}

	synced_services := []*corev1.Service{}
	for _, s := range services {
		src_service, err := service.GetService(s.Name)
		if err != nil {
			klog.Errorf("Failed to get service: %s. Err was: %s", s.Name, err)
			continue
		}

		if src_service != nil {
			annotations := s.GetAnnotations()

			if externalDns, ok := annotations["external-dns.alpha.kubernetes.io/hostname"]; ok {
				dnsTokens := strings.Split(externalDns, ".")
				annotations["external-dns.alpha.kubernetes.io/hostname"] = fmt.Sprintf(
					"%s.blue.%s", dnsTokens[0],
					strings.Join(dnsTokens[1:], "."))

				if lbInternal, ok := annotations["service.beta.kubernetes.io/aws-load-balancer-internal"]; ok {
					if lbInternal != "true" {
						annotations["service.beta.kubernetes.io/aws-load-balancer-internal"] = "true"
					}
					if private_subnets[environ] != "" {
						annotations["service.beta.kubernetes.io/aws-load-balancer-subnets"] = private_subnets[environ]
					}
				}
				s.SetAnnotations(annotations)
			}

			synced_services = append(synced_services, s)
		}
	}

	return synced_services
}

func PrintServices(services []*corev1.Service) {
	for _, s := range services {
		result, _ := yaml.Marshal(s)
		fmt.Printf("%s\n", string(result))
	}
}

func ApplyServices(kubeConfig *rest.Config, services []*corev1.Service) {
	service, err := k8s_resources.NewService(kubeConfig, corev1.NamespaceDefault)
	if err != nil {
		panic(err)
	}

	for _, s := range services {
		klog.Infof("Applying service: %s ...", s.Name)
		err := service.ApplyService(s)
		if err != nil {
			klog.Errorf("Failed to apply service. Err was: %s", err)
			continue
		}
		klog.Infoln("Done.")
	}
}
