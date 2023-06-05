package k8s_resources

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func (cj *CronJob) ListCronJobs() (*batchv1.CronJobList, error) {
	list, err := cj.client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (cj *CronJob) GetCronJob(name string) (*batchv1.CronJob, error) {
	job, err := cj.client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (cj *CronJob) CreateCronJob(job *batchv1.CronJob) error {
	_, err := cj.client.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (cj *CronJob) UpdateCronJob(job *batchv1.CronJob) error {
	_, err := cj.client.Update(context.TODO(), job, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (cj *CronJob) ApplyCronJob(cronJob *batchv1.CronJob) error {
	result, _ := cj.GetCronJob(cronJob.Name)
	if result != nil {
		containerImageMap := map[string]string{}
		for _, c := range cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
			containerImageMap[c.Name] = c.Image
		}

		for i, c := range result.Spec.JobTemplate.Spec.Template.Spec.Containers {
			result.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Image = containerImageMap[c.Name]
		}

		result.Spec.Schedule = cronJob.Spec.Schedule

		err := cj.UpdateCronJob(result)
		if err != nil {
			return err
		}
	} else {
		err := cj.CreateCronJob(cronJob)
		if err != nil {
			return err
		}
	}

	return nil
}

/* Experimental
func (s *CronJob) ApplyCronJob(job *batchv1.CronJob) error {
	cronJobApplyConfig, err := v1.ExtractCronJob(job, "k8s_resource_sync")
	if err != nil {
		return err
	}

	_, err = s.client.Apply(context.TODO(), cronJobApplyConfig, metav1.ApplyOptions{FieldManager: "k8s_resource_sync"})
	if err != nil {
		return err
	}

	return nil
}
*/
