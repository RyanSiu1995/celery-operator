package v4

import (
	"errors"
	"fmt"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetWorkers function returns the worker deployment configuration accroding to the specification
func (cr *Celery) GetWorkers() ([]*appv1.Deployment, error) {
	if &cr.Status.BrokerAddress == nil {
		return nil, errors.New("no broker is available")
	}
	var workers []*appv1.Deployment
	broker := cr.Status.BrokerAddress
	for i, workerSpec := range cr.Spec.Workers {
		labels := map[string]string{
			"celery-app":    cr.Name,
			"type":          "worker",
			"worker-number": fmt.Sprintf("worker-%d", i),
		}
		replicaNumber := int32(workerSpec.Replicas)
		appName := workerSpec.AppName
		command := []string{"celery", "worker", "-A", appName, "-b", broker}
		deployment := &appv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-worker-deployment-%d", cr.GetName(), i),
				Namespace: cr.GetNamespace(),
				Labels:    labels,
			},
			Spec: appv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Replicas: &replicaNumber,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:      fmt.Sprintf("worker-%d", i),
								Image:     cr.Spec.Image,
								Resources: workerSpec.Resources,
								Command:   command,
							},
						},
					},
				},
			},
		}
		workers = append(workers, deployment)
	}
	return workers, nil
}

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
