package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

var ctx = context.TODO()

var _ = Describe("Celery CRUD", func() {
	var template *celeryv4.Celery
	var uniqueName string
	var err error

	BeforeEach(func() {
		template = &celeryv4.Celery{}
		err = getTemplateConfig("../tests/fixtures/celery.yaml", template)
		Expect(err).NotTo(HaveOccurred())
		uniqueName = template.Name + rand.String(5)
		template.Name = uniqueName
		err = k8sClient.Create(ctx, template)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		// Clean up the environment to save the computating resources
		_ = k8sClient.Delete(ctx, template)
	})

	It("should have a single broker and worker", func() {
		// Have the celery object created
		time.Sleep(2 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName,
		}, &celeryv4.Celery{})
		Expect(err).NotTo(HaveOccurred())

		// Have the broker created
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-broker",
		}, &celeryv4.CeleryBroker{})
		Expect(err).NotTo(HaveOccurred())

		// Have two schedulers created
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-scheduler-1",
		}, &celeryv4.CeleryScheduler{})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-scheduler-2",
		}, &celeryv4.CeleryScheduler{})
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-worker-1",
		}, &celeryv4.CeleryWorker{})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-worker-2",
		}, &celeryv4.CeleryWorker{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should recreate the CRDs", func() {
		// Have the celery object created
		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName,
		}, &celeryv4.Celery{})
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.DeleteAllOf(
			ctx,
			&celeryv4.CeleryBroker{},
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
		}, &celeryv4.CeleryBroker{})
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.DeleteAllOf(ctx,
			&celeryv4.CeleryScheduler{},
			client.InNamespace("default"),
			client.MatchingLabels{
				"celery-app": uniqueName,
				"type":       "scheduler",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-scheduler-1",
		}, &celeryv4.CeleryScheduler{})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-scheduler-2",
		}, &celeryv4.CeleryScheduler{})
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.DeleteAllOf(ctx,
			&celeryv4.CeleryWorker{},
			client.InNamespace("default"),
			client.MatchingLabels{
				"celery-app": uniqueName,
				"type":       "worker",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-worker-1",
		}, &celeryv4.CeleryWorker{})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      uniqueName + "-worker-2",
		}, &celeryv4.CeleryWorker{})
		Expect(err).NotTo(HaveOccurred())
	})
})
