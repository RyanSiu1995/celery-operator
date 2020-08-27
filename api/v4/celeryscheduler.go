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

func (csr *CeleryScheduler) Replace(i int) *corev1.Pod {
	labels := map[string]string{
		"celery-app": csr.Name,
		"type":       "scheduler",
	}
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
	csr.Status.PodList[i] = pod
	return pod
}

// Generate will create the pod spec of the broker.
func (csr *CeleryScheduler) Generate() []*corev1.Pod {
	// If podList exists, then returns the podList instead
	if len(csr.Status.PodList) != 0 {
		return csr.Status.PodList
	}

	labels := map[string]string{
		"celery-app": csr.Name,
		"type":       "scheduler",
	}

	podList := make([]*corev1.Pod, 0)
	for i := 0; i < csr.Spec.Replicas; i++ {
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
	csr.Status.PodList = podList
	return podList
}
