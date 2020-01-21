package celery

import (
	celeryprojectv4 "github.com/RyanSiu1995/celery-operator/pkg/apis/celeryproject/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func generateScheduler(cr *celeryprojectv4.Celery, brokerString string) *corev1.Pod {
	labels := map[string]string{
		"celery-app": cr.Name,
		"type":       "scheduler",
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-scheulder",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "",
					Image:   "",
					Command: []string{"celery"},
					Args: []string{
						"beat",
						"-A",
						cr.Name,
						"-b",
						brokerString,
					},
				},
			},
		},
	}
}

func generateBroker(cr *celeryprojectv4.Celery) (*corev1.Pod, *corev1.Service) {
	if cr.Spec.Broker.Type == celeryprojectv4.ExternalBroker {
		return nil, nil
	}

	labels := map[string]string{
		"celery-app": cr.Name,
		"type":       "broker",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-service",
			Namespace: cr.Namespace,
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
			Name:      cr.Name + "-service",
			Namespace: cr.Namespace,
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

	return pod, service
}
