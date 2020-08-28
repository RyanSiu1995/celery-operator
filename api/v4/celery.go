package v4

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Generate Broker defines the way to create one broker based on config
func (cr *Celery) GenerateBroker() *CeleryBroker {
	labels := map[string]string{
		"celery-app": cr.Name,
		"type":       "broker",
	}

	return &CeleryBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetName() + "-broker",
			Namespace: cr.GetNamespace(),
			Labels:    labels,
		},
		Spec: cr.Spec.Broker,
	}
}

// GenerateSchedulers defines the way to create schedulers based on config
func (cr *Celery) GenerateSchedulers() []*CeleryScheduler {
	labels := map[string]string{
		"celery-app": cr.Name,
		"type":       "scheduler",
	}
	defaultImage := cr.Spec.Image
	brokerAddr := cr.Status.BrokerAddress
	schedulers := make([]*CeleryScheduler, 0)
	for i, schedulerSpec := range cr.Spec.Schedulers {
		if schedulerSpec.Image == "" {
			schedulerSpec.Image = defaultImage
		}
		schedulerSpec.BrokerAddress = brokerAddr
		scheduler := &CeleryScheduler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-scheduler-%d", cr.GetName(), i+1),
				Namespace: cr.GetNamespace(),
				Labels:    labels,
			},
			Spec: schedulerSpec,
		}
		schedulers = append(schedulers, scheduler)
	}
	return schedulers
}

// GenerateWorkers defines the way to create workers based on config
func (cr *Celery) GenerateWorkers() []*CeleryWorker {
	labels := map[string]string{
		"celery-app": cr.Name,
		"type":       "worker",
	}
	defaultImage := cr.Spec.Image
	brokerAddr := cr.Status.BrokerAddress
	workers := make([]*CeleryWorker, 0)
	for i, workerSpec := range cr.Spec.Workers {
		if workerSpec.Image == "" {
			workerSpec.Image = defaultImage
		}
		workerSpec.BrokerAddress = brokerAddr
		worker := &CeleryWorker{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-worker-%d", cr.GetName(), i+1),
				Namespace: cr.GetNamespace(),
				Labels:    labels,
			},
			Spec: workerSpec,
		}
		workers = append(workers, worker)
	}
	return workers
}
