package controllers

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

var _ = Describe("CeleryWorker CRUD", func() {
	It("should have two worker pods", func() {
		celeryworkerSpecInYaml, err := ioutil.ReadFile("../tests/fixtures/celery_workers.yaml")
		Expect(err).NotTo(HaveOccurred())
		celeryworkerObject := &celeryv4.CeleryWorker{}
		celeryworkerSpecInJSON, err := yaml.YAMLToJSON(celeryworkerSpecInYaml)
		Expect(err).NotTo(HaveOccurred())
		err = json.Unmarshal(celeryworkerSpecInJSON, celeryworkerObject)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(ctx, celeryworkerObject)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(2 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-worker-test-1",
		}, &celeryv4.CeleryWorker{})
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(2 * time.Second)
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": "celery-worker-test-1",
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(2))
	})

	It("should respawan worker pod after deletion", func() {
		celeryworkerSpecInYaml, err := ioutil.ReadFile("../tests/fixtures/celery_workers_2.yaml")
		Expect(err).NotTo(HaveOccurred())
		celeryworkerObject := &celeryv4.CeleryWorker{}
		celeryworkerSpecInJSON, err := yaml.YAMLToJSON(celeryworkerSpecInYaml)
		Expect(err).NotTo(HaveOccurred())
		err = json.Unmarshal(celeryworkerSpecInJSON, celeryworkerObject)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(ctx, celeryworkerObject)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-worker-test-2",
		}, &celeryv4.CeleryWorker{})
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(1 * time.Second)
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": "celery-worker-test-2",
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(1))

		err = k8sClient.DeleteAllOf(ctx, &corev1.Pod{}, client.InNamespace("default"), client.MatchingLabels{"celery-app": "celery-worker-test-2", "type": "worker"})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		newPodList := &corev1.PodList{}
		err = k8sClient.List(ctx, newPodList, client.MatchingLabels{
			"celery-app": "celery-worker-test-2",
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(newPodList.Items)).To(Equal(1))
		Expect(newPodList.Items[0].Name).NotTo(Equal(podList.Items[0].Name))
	})

	It("should update successfully", func() {
		celeryworkerSpecInYaml, err := ioutil.ReadFile("../tests/fixtures/celery_workers_3.yaml")
		Expect(err).NotTo(HaveOccurred())
		celeryworkerObject := &celeryv4.CeleryWorker{}
		celeryworkerSpecInJSON, err := yaml.YAMLToJSON(celeryworkerSpecInYaml)
		Expect(err).NotTo(HaveOccurred())
		err = json.Unmarshal(celeryworkerSpecInJSON, celeryworkerObject)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(ctx, celeryworkerObject)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(1 * time.Second)
		old := &celeryv4.CeleryWorker{}
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-worker-test-3",
		}, old)
		Expect(err).NotTo(HaveOccurred())
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": "celery-worker-test-3",
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(1))
		Expect(podList.Items[0].Spec.Containers[0].Command).To(Equal([]string{
			"celery",
			"worker",
			"-A",
			"appName",
			"-b",
			"redis://127.0.0.1/1",
			"--queues",
			"test1,test2",
		}))

		old.Spec.TargetQueues = []string{"test1"}
		err = k8sClient.Update(ctx, old)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": "celery-worker-test-3",
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(1))
		Expect(podList.Items[0].Spec.Containers[0].Command).To(Equal([]string{
			"celery",
			"worker",
			"-A",
			"appName",
			"-b",
			"redis://127.0.0.1/1",
			"--queues",
			"test1",
		}))
	})

	It("should change the replica successfully", func() {
		celeryworkerSpecInYaml, err := ioutil.ReadFile("../tests/fixtures/celery_workers_4.yaml")
		Expect(err).NotTo(HaveOccurred())
		celeryworkerObject := &celeryv4.CeleryWorker{}
		celeryworkerSpecInJSON, err := yaml.YAMLToJSON(celeryworkerSpecInYaml)
		Expect(err).NotTo(HaveOccurred())
		err = json.Unmarshal(celeryworkerSpecInJSON, celeryworkerObject)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(ctx, celeryworkerObject)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(1 * time.Second)
		old := &celeryv4.CeleryWorker{}
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-worker-test-4",
		}, old)
		Expect(err).NotTo(HaveOccurred())
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": "celery-worker-test-4",
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(1))

		old.Spec.Replicas = 4
		err = k8sClient.Update(ctx, old)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": "celery-worker-test-4",
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(4))

		time.Sleep(1 * time.Second)

		old.Spec.Replicas = 1
		err = k8sClient.Update(ctx, old)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.List(ctx, podList, client.MatchingLabels{
			"celery-app": "celery-worker-test-4",
			"type":       "worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(1))
	})
})
