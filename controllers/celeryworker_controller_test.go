package controllers

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

var _ = Describe("CeleryWorker CRUD", func() {
	var template *celeryv4.CeleryWorker
	var uniqueName string
	var err error

	BeforeEach(func() {
		template = &celeryv4.CeleryWorker{}
		err = getTemplateConfig("../tests/fixtures/celery_workers.yaml", template)
		Expect(err).NotTo(HaveOccurred())
		uniqueName = template.Name + rand.String(5)
		template.Name = uniqueName
		err = k8sClient.Create(ctx, template)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName,
		}, template)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should have two worker pods", func() {
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(2))
	})

	It("should respawan worker pod after deletion", func() {
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(2))

		err = k8sClient.DeleteAllOf(ctx,
			&corev1.Pod{},
			client.InNamespace("default"),
			client.MatchingLabels{
				"celery-app": uniqueName,
				"type":       "worker",
			},
		)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(1 * time.Second)
		newPodList := &corev1.PodList{}
		err = k8sClient.List(ctx, newPodList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(newPodList.Items)).To(Equal(2))
		for i, _ := range newPodList.Items {
			Expect(newPodList.Items[i].Name).NotTo(Equal(podList.Items[i].Name))
		}
	})

	It("should update successfully", func() {
		template.Spec.TargetQueues = []string{"test1"}
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(2))
		for i, _ := range podList.Items {
			Expect(podList.Items[i].Spec.Containers[0].Command).To(Equal([]string{
				"celery",
				"worker",
				"-A",
				"appName",
				"-b",
				"redis://127.0.0.1/1",
				"--queues",
				"test1",
			}))
		}
	})

	It("should change the replica successfully", func() {
		template.Spec.Replicas = 4
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(4))

		time.Sleep(1 * time.Second)

		template.Spec.Replicas = 1
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(1))
	})
})
