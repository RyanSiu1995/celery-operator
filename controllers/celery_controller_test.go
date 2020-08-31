package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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
	var ensureObjectCreatedGenerator = func(targetObject runtime.Object, targetName string, defaultCount ...int) func(...int) {
		timeout := 2
		pollInterval := 0.01
		if len(defaultCount) == 0 {
			return func(_ ...int) {
				Eventually(func() error {
					return k8sClient.Get(ctx, client.ObjectKey{
						Namespace: "default",
						Name:      fmt.Sprintf("%s-%s", uniqueName, targetName),
					}, targetObject)
				}, timeout, pollInterval).Should(BeNil())
			}
		} else {
			return func(count ...int) {
				def := defaultCount[0]
				if len(count) > 0 {
					def = count[0]
				}
				Eventually(func() bool {
					for i := 0; i < def; i++ {
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

	It("should update the broker properly", func() {
		ensureBrokerCreated()
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      fmt.Sprintf("%s-broker-broker", uniqueName),
			}, &corev1.Pod{})
		}).Should(BeNil())
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      fmt.Sprintf("%s-broker-broker-service", uniqueName),
			}, &corev1.Service{})
		}).Should(BeNil())

		template.Spec.Broker.Type = celeryv4.ExternalBroker
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() celeryv4.BrokerType {
			broker := &celeryv4.CeleryBroker{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-broker", uniqueName),
				}, broker)
			}).Should(BeNil())
			return broker.Spec.Type
		}, 2, 0.1).Should(Equal(celeryv4.ExternalBroker))
	})

	It("should increase and decrease the scheduler properly", func() {
		// Delete all schedulers and wait for respawning
		ensureSchedulersCreated()
		template.Spec.Schedulers = append(template.Spec.Schedulers, celeryv4.CelerySchedulerSpec{
			SchedulerClass: "a.b.c",
			AppName:        "appName2",
			Replicas:       1,
		})
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		ensureSchedulersCreated(3)
		Eventually(func() int {
			list := &celeryv4.CelerySchedulerList{}
			Eventually(func() error {
				return k8sClient.List(ctx, list, client.MatchingLabels{
					"celery-app": template.Name,
					"type":       "scheduler",
				})
			}).Should(BeNil())
			return len(list.Items)
		}).Should(BeNumerically("==", 3))

		// Delete all schedulers and wait for respawning
		ensureSchedulersCreated()
		template.Spec.Schedulers = template.Spec.Schedulers[:1]
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		ensureSchedulersCreated(1)
		Eventually(func() int {
			list := &celeryv4.CelerySchedulerList{}
			Eventually(func() error {
				return k8sClient.List(ctx, list, client.MatchingLabels{
					"celery-app": template.Name,
					"type":       "scheduler",
				})
			}).Should(BeNil())
			return len(list.Items)
		}, 2, 0.1).Should(BeNumerically("==", 1))
	})

	It("should update the scheduler correctly", func() {
		Skip("scheduler logic hasn't been implemented yet")
		// Delete all schedulers and wait for respawning
		ensureSchedulersCreated()
		template.Spec.Schedulers[0].AppName = "updatedAppName"
		template.Spec.Schedulers = append(template.Spec.Schedulers, celeryv4.CelerySchedulerSpec{
			SchedulerClass: "a.b.c",
			AppName:        "appName2",
			Replicas:       1,
		})
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		ensureSchedulersCreated(3)
		Eventually(func() string {
			scheduler := &celeryv4.CeleryScheduler{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-scheduler-1", uniqueName),
				}, scheduler)
			}).Should(BeNil())
			return scheduler.Spec.AppName
		}, 2, 0.1).Should(Equal("updateAppName"))
	})

	It("should update worker correctly", func() {
		ensureWorkersCreated()
		podList := &corev1.PodList{}
		Eventually(func() []string {
			Eventually(func() int {
				Eventually(func() error {
					return k8sClient.List(ctx, podList, client.MatchingLabels{
						"celery-app": fmt.Sprintf("%s-worker-1", uniqueName),
						"type":       "worker",
					})
				}).Should(BeNil())
				return len(podList.Items)
			}).Should(BeNumerically("==", 1))
			return podList.Items[0].Spec.Containers[0].Command
		}).Should(Equal([]string{
			"celery",
			"worker",
			"-A",
			"test1",
			"-b",
			"", // FIXME The broker is not set
		}))
		template.Spec.Workers[0].AppName = "newAppName"
		err = k8sClient.Update(ctx, template)
		Eventually(func() []string {
			Eventually(func() int {
				Eventually(func() error {
					return k8sClient.List(ctx, podList, client.MatchingLabels{
						"celery-app": fmt.Sprintf("%s-worker-1", uniqueName),
						"type":       "worker",
					})
				}).Should(BeNil())
				return len(podList.Items)
			}).Should(BeNumerically("==", 1)) // FIXME This is not the correct replicas
			return podList.Items[0].Spec.Containers[0].Command
		}).Should(Equal([]string{
			"celery",
			"worker",
			"-A",
			"newAppName",
			"-b",
			"", // FIXME The broker is not set
		}))
	})

	It("should increase and decrease the worker properly", func() {
		ensureWorkersCreated()
		template.Spec.Workers = append(template.Spec.Workers, celeryv4.CeleryWorkerSpec{
			AppName:  "appName2",
			Replicas: 1,
		})
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		ensureWorkersCreated(3)
		Eventually(func() int {
			list := &celeryv4.CeleryWorkerList{}
			Eventually(func() error {
				return k8sClient.List(ctx, list, client.MatchingLabels{
					"celery-app": template.Name,
					"type":       "worker",
				})
			}).Should(BeNil())
			return len(list.Items)
		}, 2, 0.1).Should(BeNumerically("==", 3))

		template.Spec.Workers = template.Spec.Workers[:1]
		err = k8sClient.Update(ctx, template)
		Expect(err).NotTo(HaveOccurred())
		ensureWorkersCreated(1)
		Eventually(func() int {
			list := &celeryv4.CeleryWorkerList{}
			Eventually(func() error {
				return k8sClient.List(ctx, list, client.MatchingLabels{
					"celery-app": template.Name,
					"type":       "worker",
				})
			}).Should(BeNil())
			return len(list.Items)
		}, 2, 0.1).Should(BeNumerically("==", 1))
	})
})
