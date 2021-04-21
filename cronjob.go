package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/mwlng/k8s_resources_sync/pkg/k8s_resources"
)

func LoadCronJobYamlFiles(rootDir string) []*batchv1.CronJob {
	cronJobs := []*batchv1.CronJob{}
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
				case *batchv1.CronJob:
					cronJobs = append(cronJobs, obj.(*batchv1.CronJob))
				}
			}
		}
		return nil
	})

	if err != nil {
		klog.Errorf("Error while reading YAML files. Err was: %s", err)
	}

	return cronJobs
}

func SyncCronJobs(kubeConfig *rest.Config, cronJobs []*batchv1.CronJob) []*batchv1.CronJob {
	klog.Infof("Syncing cron jobs from cluster: %s, namespace %s\n", kubeConfig.Host, corev1.NamespaceDefault)
	cronJob, err := k8s_resources.NewCronJob(kubeConfig, corev1.NamespaceDefault)
	if err != nil {
		panic(err)
	}

	synced_cronJobs := []*batchv1.CronJob{}
	for _, job := range cronJobs {
		src_cronJob, err := cronJob.GetCronJob(job.Name)
		if err != nil {
			klog.Errorf("Failed to get cron job: %s. Err was: %s", job.Name, err)
			continue
		}

		if src_cronJob != nil {
			synced_cronJobs = append(synced_cronJobs, job)
		}
	}

	return synced_cronJobs
}

func PrintCronJobs(cronJobs []*batchv1.CronJob) {
	for _, job := range cronJobs {
		result, _ := yaml.Marshal(job)
		fmt.Printf("%s\n", string(result))
	}
}

func ApplyCronJobs(kubeConfig *rest.Config, cronJobs []*batchv1.CronJob) {
	cronJob, err := k8s_resources.NewCronJob(kubeConfig, corev1.NamespaceDefault)
	if err != nil {
		panic(err)
	}

	for _, job := range cronJobs {
		klog.Infof("Applying cron job: %s ...", job.Name)
		err := cronJob.ApplyCronJob(job)
		if err != nil {
			klog.Errorf("Failed to apply cron job. Err was: %s", err)
			continue
		}
		klog.Infoln("Done.")
	}
}
