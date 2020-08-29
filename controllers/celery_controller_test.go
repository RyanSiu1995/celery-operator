package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

var ctx = context.TODO()

var _ = Describe("Celery CRUD", func() {
	// Global Test Objects
	var template *celeryv4.Celery
	var uniqueName string
	var err error

	// Utility functions
	var ensureObjectCreatedGenerator = func(targetObject runtime.Object, targetName string, count ...int) func() {
		timeout := 2
		pollInterval := 0.01
		if len(count) == 0 {
			return func() {
				Eventually(func() error {
					return k8sClient.Get(ctx, client.ObjectKey{
						Namespace: "default",
						Name:      fmt.Sprintf("%s-%s", uniqueName, targetName),
					}, targetObject)
				}, timeout, pollInterval).Should(BeNil())
			}
		} else {
			return func() {
				Eventually(func() bool {
					for i := 0; i < count[0]; i++ {
						err = k8sClient.Get(ctx, client.ObjectKey{
							Namespace: "default",
							Name:      fmt.Sprintf("%s-%s-%d", uniqueName, targetName, i+1),
						}, targetObject)
						if err != nil {
							return false
						}
					}
					return true
				}, timeout, pollInterval).Should(BeTrue())
			}
		}
	}

	var ensureBrokerCreated = ensureObjectCreatedGenerator(&celeryv4.CeleryBroker{}, "broker")
	var ensureWorkersCreated = ensureObjectCreatedGenerator(&celeryv4.CeleryWorker{}, "worker", 2)
	var ensureSchedulersCreated = ensureObjectCreatedGenerator(&celeryv4.CeleryScheduler{}, "scheduler", 2)

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
