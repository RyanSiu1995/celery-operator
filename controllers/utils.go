package controllers

import (
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	celeryprojectv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

func generateScheduler(cr *celeryprojectv4.Celery, brokerString string) *appv1.Deployment {
	labels := map[string]string{
		"celery-app": cr.Name,
		"type":       "scheduler",
	}

	replicaNumber := int32(1)
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-scheduler-deployment",
			Namespace: cr.Namespace,
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
							Name:    "celery-scheduler",
							Image:   "celery:4",
							Command: []string{"celery"},
							Args: []string{
								"beat",
								"-b",
								brokerString,
							},
						},
					},
				},
			},
		},
	}
}
