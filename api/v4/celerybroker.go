package v4

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Generate will create the pod spec of the broker.
func (cbr *CeleryBroker) Generate() (*corev1.Pod, *corev1.Service, string) {
	labels := map[string]string{
		"celery-app": cbr.Name,
		"type":       "broker",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cbr.GetName() + "-broker-service",
			Namespace: cbr.GetNamespace(),
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
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cbr.GetName() + "-broker",
			Namespace: cbr.GetNamespace(),
			Labels:    labels,
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
	}
	return pod, service, fmt.Sprintf("redis://%s.%s", cbr.Name+"-broker-service", cbr.Namespace)
}
