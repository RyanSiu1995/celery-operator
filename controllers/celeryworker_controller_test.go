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

var _ = Describe("CeleryWorker Creation", func() {
	Describe("CeleryWorker Creation", func() {
		It("should have a single worker pod", func() {
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
	})
})
