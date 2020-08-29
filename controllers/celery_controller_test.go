package controllers

import (
	"context"
	"fmt"

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

	var ensureBrokerCreated = func() {
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      uniqueName + "-broker",
			}, &celeryv4.CeleryBroker{})
		}, 2, 0.01).Should(BeNil())
	}

	var ensureWorkersCreated = func() {
		Eventually(func() bool {
			for i := 0; i < 2; i++ {
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-worker-%d", uniqueName, i+1),
				}, &celeryv4.CeleryWorker{})
				if err != nil {
					return false
				}
			}
			return true
		}, 2, 0.01).Should(BeTrue())
	}

	var ensureSchedulersCreated = func() {
		Eventually(func() bool {
			for i := 0; i < 2; i++ {
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-scheduler-%d", uniqueName, i+1),
				}, &celeryv4.CeleryScheduler{})
				if err != nil {
					return false
				}
			}
			return true
		}, 2, 0.01).Should(BeTrue())
	}

	BeforeEach(func() {
		template = &celeryv4.Celery{}
		err = getTemplateConfig("../tests/fixtures/celery.yaml", template)
		Expect(err).NotTo(HaveOccurred())
		uniqueName = template.Name + rand.String(5)
		template.Name = uniqueName

		// Create the test object
		err = k8sClient.Create(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      uniqueName,
			}, &celeryv4.Celery{})
		}).Should(BeNil())
	})

	AfterEach(func() {
		// Clean up the environment to save the computating resources
		_ = k8sClient.Delete(ctx, template)
	})

	It("should have a single broker and worker", func() {
		ensureBrokerCreated()
		ensureWorkersCreated()
		ensureSchedulersCreated()
	})

	It("should recreate the CRDs", func() {
		// Delete all brokers and wait for respawning
		ensureBrokerCreated()
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
		ensureBrokerCreated()

		// Delete all schedulers and wait for respawning
		ensureSchedulersCreated()
		err = k8sClient.DeleteAllOf(ctx,
			&celeryv4.CeleryScheduler{},
			client.InNamespace("default"),
			client.MatchingLabels{
				"celery-app": uniqueName,
				"type":       "scheduler",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		ensureSchedulersCreated()

		// Delete all workers and wait for respawning
		ensureWorkersCreated()
		err = k8sClient.DeleteAllOf(ctx,
			&celeryv4.CeleryWorker{},
			client.InNamespace("default"),
			client.MatchingLabels{
				"celery-app": uniqueName,
				"type":       "worker",
			},
		)
		Expect(err).NotTo(HaveOccurred())
		ensureWorkersCreated()
	})
})
