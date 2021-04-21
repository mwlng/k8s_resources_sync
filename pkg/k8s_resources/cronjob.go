package k8s_resources

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/client-go/applyconfigurations/batch/v1"
	typedv1 "k8s.io/client-go/kubernetes/typed/batch/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CronJob struct {
	client typedv1.CronJobInterface
}

func NewCronJob(config *rest.Config, namespace string) (*CronJob, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &CronJob{
		client: clientset.BatchV1().CronJobs(namespace),
	}, nil
}

func (d *CronJob) ListCronJobs() (*batchv1.CronJobList, error) {
	list, err := d.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (d *CronJob) GetCronJob(name string) (*batchv1.CronJob, error) {
	job, err := d.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (d *CronJob) CreateCronJob(job *batchv1.CronJob) error {
	_, err := d.client.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (d *CronJob) ApplyCronJob(job *batchv1.CronJob) error {
	cronJobApplyConfig, err := v1.ExtractCronJob(job, "k8s_resource_sync")
	if err != nil {
		return err
	}

	_, err = d.client.Apply(context.TODO(), cronJobApplyConfig, metav1.ApplyOptions{})
	if err != nil {
		return err
	}

	return nil
}
