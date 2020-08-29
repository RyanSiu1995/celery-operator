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

var _ = Describe("CeleryScheduler CRUD", func() {
	var template *celeryv4.CeleryScheduler
	var uniqueName string
	var err error

	BeforeEach(func() {
		template = &celeryv4.CeleryScheduler{}
		err = getTemplateConfig("../tests/fixtures/celery_schedulers.yaml", template)
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

	AfterEach(func() {
		// Clean up the environment to save the computating resources
		_ = k8sClient.Delete(ctx, template)
	})

	It("should have two scheduler pods", func() {
		time.Sleep(1 * time.Second)
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "scheduler",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(2))
	})

	It("should respawn the scheduler pod after deletion", func() {
		time.Sleep(1 * time.Second)
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "scheduler",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(2))

		err = k8sClient.DeleteAllOf(
			ctx,
			&corev1.Pod{},
			client.InNamespace("default"),
			client.MatchingLabels{
				"celery-app": uniqueName,
				"type":       "scheduler",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		newPodList := &corev1.PodList{}
		err = k8sClient.List(ctx, newPodList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "scheduler",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(newPodList.Items)).To(Equal(2))
		for i, _ := range newPodList.Items {
			Expect(newPodList.Items[i].Name).NotTo(Equal(podList.Items[i].Name))
		}
	})

	It("should schedule the pods correctly", func() {
		time.Sleep(1 * time.Second)
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "scheduler",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(2))

		template.Spec.Replicas = 3
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "scheduler",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(3))

		template.Spec.Replicas = 1
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": uniqueName,
			"type":       "scheduler",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(1))
	})
})
