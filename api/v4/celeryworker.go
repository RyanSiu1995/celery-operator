package v4

import (
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func (cwr *CeleryWorker) getCommand() []string {
	command := []string{"celery", "worker", "-A", cwr.Spec.AppName, "-b", cwr.Spec.BrokerAddress}
	if len(cwr.Spec.TargetQueues) > 0 {
		command = append(command, []string{
			"--queues",
			strings.Join(cwr.Spec.TargetQueues, ","),
		}...)
	}
	return command
}

func (cwr *CeleryWorker) IsUpToDate(podList []corev1.Pod) bool {
	for _, pod := range podList {
		if len(pod.Spec.Containers) != 1 ||
			pod.Spec.Containers[0].Image != cwr.Spec.Image ||
			strings.Join(pod.Spec.Containers[0].Command, "") != strings.Join(cwr.getCommand(), "") ||
			!reflect.DeepEqual(pod.Spec.Containers[0].Resources, cwr.Spec.Resources) {
			return false
		}
	}
	return true
}

// Generate will create the pod spec of the worker.
func (cwr *CeleryWorker) Generate(count ...int) []*corev1.Pod {
	var targetNumber int
	if len(count) == 0 {
		targetNumber = cwr.Spec.Replicas
	} else {
		targetNumber = count[0]
	}

	labels := map[string]string{
		"celery-app": cwr.Name,
		"type":       "worker",
	}

	podList := make([]*corev1.Pod, 0)
	for i := 0; i < targetNumber; i++ {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cwr.GetName() + "-" + rand.String(5),
				Namespace: cwr.GetNamespace(),
				Labels:    labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:      "celery-worker",
						Image:     cwr.Spec.Image,
						Resources: cwr.Spec.Resources,
						Command:   cwr.getCommand(),
					},
				},
			},
		}
		podList = append(podList, pod)
	}
	return podList
}
