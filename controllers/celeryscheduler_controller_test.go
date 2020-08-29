package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

var _ = Describe("CeleryScheduler CRUD", func() {
	// Global Test Objects
	var template *celeryv4.CeleryScheduler
	var uniqueName string
	var err error

	// Utility functions
	var ensureNumberOfSchedulersToBe = func(target int) *corev1.PodList {
		podList := &corev1.PodList{}
		Eventually(func() int {
			podList := &corev1.PodList{}
			Eventually(func() error {
				return k8sClient.List(ctx, podList, client.MatchingLabels{
					"celery-app": uniqueName,
					"type":       "scheduler",
				})
			}).Should(BeNil())
			return len(podList.Items)
		}, 2, 0.01).Should(BeNumerically("==", target))
		return podList
	}

	BeforeEach(func() {
		template = &celeryv4.CeleryScheduler{}
		err = getTemplateConfig("../tests/fixtures/celery_schedulers.yaml", template)
		Expect(err).NotTo(HaveOccurred())
		uniqueName = template.Name + rand.String(5)
		template.Name = uniqueName
		err = k8sClient.Create(ctx, template)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      uniqueName,
			}, template)
		}).Should(BeNil())
	})

	AfterEach(func() {
		// Clean up the environment to save the computating resources
		_ = k8sClient.Delete(ctx, template)
	})

	It("should have two scheduler pods", func() {
		ensureNumberOfSchedulersToBe(2)
	})

	It("should respawn the scheduler pod after deletion", func() {
		// Get the old pod for comparison
		podList := ensureNumberOfSchedulersToBe(2)

		// Respawn the pods by deleting the old one
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
		newPodList := ensureNumberOfSchedulersToBe(2)

		for i, _ := range newPodList.Items {
			Expect(newPodList.Items[i].Name).NotTo(Equal(podList.Items[i].Name))
		}
	})

	It("should schedule the pods correctly", func() {
		template.Spec.Replicas = 3
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		ensureNumberOfSchedulersToBe(3)

		// Wait for the stablization of the resources
		Consistently(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      uniqueName,
			}, template)
		}).Should(BeNil())

		template.Spec.Replicas = 1
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		ensureNumberOfSchedulersToBe(1)
	})
})
