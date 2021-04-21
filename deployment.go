package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/mwlng/k8s_resources_sync/pkg/dyna_client"
	"github.com/mwlng/k8s_resources_sync/pkg/k8s_resources"
)

func LoadDeploymentYamlFiles(rootDir string) []*appsv1.Deployment {
	deployments := []*appsv1.Deployment{}
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
				case *appsv1.Deployment:
					deployments = append(deployments, obj.(*appsv1.Deployment))
				}
			}
		}
		return nil
	})

	if err != nil {
		klog.Errorf("Error while reading YAML files. Err was: %s", err)
	}

	return deployments
}

func SyncDeployments(kubeConfig *rest.Config, deployments []*appsv1.Deployment) {
	klog.Infof("Syncing deployments from cluster: %s, namespace: %s\n", kubeConfig.Host, corev1.NamespaceDefault)
	deployment, err := k8s_resources.NewDeployment(kubeConfig, corev1.NamespaceDefault)
	if err != nil {
		panic(err)
	}

	for _, d := range deployments {
		src_deployment, err := deployment.GetDeployment(d.Name)
		if err != nil {
			klog.Errorf("Failed to get deployment: %s. Err was: %s", d.Name, err)
			continue
		}

		containerImageMap := map[string]string{}
		for _, c := range src_deployment.Spec.Template.Spec.Containers {
			containerImageMap[c.Name] = c.Image
		}

		for i, c := range d.Spec.Template.Spec.Containers {
			d.Spec.Template.Spec.Containers[i].Image = containerImageMap[c.Name]
		}

		d.Spec.Replicas = src_deployment.Spec.Replicas
	}
}

func PrintDeployments(deployments []*appsv1.Deployment) {
	for _, d := range deployments {
		result, _ := yaml.Marshal(d)
		fmt.Printf("%s\n", string(result))
	}
}

func ApplyDeployments(kubeConfig *rest.Config, deployments []*appsv1.Deployment) {
	dyna_client, err := dyna_client.NewDynaClient(kubeConfig)
	if err != nil {
		panic(err)
	}

	for _, d := range deployments {
		klog.Infof("Applying deployment %s ...", d.Name)
		data, _ := yaml.Marshal(d)
		err := dyna_client.Apply(data, "k8s_reources_sync")
		if err != nil {
			klog.Errorf("Failed to apply deployment. Err was: %s", err)
			continue
		}
		klog.Infoln("Done.")
	}
}
