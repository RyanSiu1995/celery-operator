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

var _ = Describe("CeleryBroker CRUD", func() {
	var template *celeryv4.CeleryBroker
	var uniqueName string
	var err error

	BeforeEach(func() {
		template = &celeryv4.CeleryBroker{}
		err = getTemplateConfig("../tests/fixtures/celery_broker.yaml", template)
		Expect(err).NotTo(HaveOccurred())
		uniqueName = template.Name + rand.String(5)
		template.Name = uniqueName
		err = k8sClient.Create(ctx, template)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName,
		}, &celeryv4.CeleryBroker{})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		// Clean up the environment to save the computating resources
		_ = k8sClient.Delete(ctx, template)
	})

	It("should have a single broker pod and service", func() {
		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-broker",
		}, &corev1.Pod{})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-broker-service",
		}, &corev1.Service{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should recreate the service and pod after deleting them", func() {
		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-broker",
		}, &corev1.Pod{})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-broker-service",
		}, &corev1.Service{})
		Expect(err).NotTo(HaveOccurred())

		// Not able to delete the service
		err = k8sClient.DeleteAllOf(ctx,
			&corev1.Service{},
			client.InNamespace("default"),
			client.MatchingLabels{
				"celery-app": uniqueName,
				"type":       "broker",
			},
		)
		Expect(err).To(HaveOccurred())
		time.Sleep(1 * time.Second)

		// Delete pod
		err = k8sClient.DeleteAllOf(ctx,
			&corev1.Pod{},
			client.InNamespace("default"),
			client.MatchingLabels{
				"celery-app": uniqueName,
				"type":       "broker",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-broker",
		}, &corev1.Pod{})
		Expect(err).NotTo(HaveOccurred())
	})
})
