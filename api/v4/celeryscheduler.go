package v4

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func (csr *CeleryScheduler) getCommand() []string {
	command := []string{"celery", "beat", "-A", csr.Spec.AppName, "-b", csr.Spec.BrokerAddress}
	if csr.Spec.SchedulerClass != "" {
		command = append(command, []string{"--scheduler", csr.Spec.SchedulerClass}...)
	}
	return command
}

// Generate will create the pod spec of the broker.
func (csr *CeleryScheduler) Generate(count ...int) []*corev1.Pod {
	var targetNumber int
	if len(count) == 0 {
		targetNumber = csr.Spec.Replicas
	} else {
		targetNumber = count[0]
	}

	labels := map[string]string{
		"celery-app": csr.Name,
		"type":       "scheduler",
	}

	podList := make([]*corev1.Pod, 0)
	for i := 0; i < targetNumber; i++ {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      csr.GetName() + "-" + rand.String(5),
				Namespace: csr.GetNamespace(),
				Labels:    labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:      "celery-scheduler",
						Image:     csr.Spec.Image,
						Resources: csr.Spec.Resources,
						Command:   csr.getCommand(),
					},
				},
			},
		}
		podList = append(podList, pod)
	}
	return podList
}
