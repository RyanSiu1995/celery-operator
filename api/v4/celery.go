package v4

import (
	"errors"
	"fmt"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

// GetBroker returns the broker information
func (cr *Celery) GetBroker() (*appv1.Deployment, *corev1.Service, error) {
	if cr.Spec.Broker.Type == ExternalBroker {
		brokerAddress := cr.Spec.Broker.BrokerAddress
		if &brokerAddress == nil {
			return nil, nil, errors.New("No Broker Address is given in External Broker Mode")
		}
		cr.Status.BrokerAddress = brokerAddress
		return nil, nil, nil
	}
	deployment, service, brokerAddress := cr.generateBroker()
	cr.Status.BrokerAddress = brokerAddress
	return deployment, service, nil
}

func (cr *Celery) generateBroker() (*appv1.Deployment, *corev1.Service, string) {
	labels := map[string]string{
		"celery-app": cr.Name,
		"type":       "broker",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetName() + "-broker-service",
			Namespace: cr.GetNamespace(),
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     "ClusterIP",
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "redis-port",
					Port:       6379,
					TargetPort: intstr.FromInt(6379),
				},
			},
		},
	}

	replicaNumber := int32(1)
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetName() + "-broker-deployment",
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
							Name:  "redis",
							Image: "redis:3.0.5",
							Ports: []corev1.ContainerPort{
								{
									Name:          "redis",
									ContainerPort: 6379,
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment, service, fmt.Sprintf("redis://%s.%s", cr.Name+"-broker-service", cr.Namespace)
}
