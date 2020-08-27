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

var _ = Describe("CeleryBroker Creation", func() {
	Describe("CeleryBroker Creation", func() {
		Context("Create a broker and worker", func() {
			It("should have a single broker pod and service", func() {
				celerybrokerSpecInYaml, err := ioutil.ReadFile("../tests/fixtures/celery_broker.yaml")
				Expect(err).NotTo(HaveOccurred())
				celerybrokerObject := &celeryv4.CeleryBroker{}
				celerybrokerSpecInJSON, err := yaml.YAMLToJSON(celerybrokerSpecInYaml)
				Expect(err).NotTo(HaveOccurred())
				err = json.Unmarshal(celerybrokerSpecInJSON, celerybrokerObject)
				Expect(err).NotTo(HaveOccurred())
				err = k8sClient.Create(ctx, celerybrokerObject)
				Expect(err).NotTo(HaveOccurred())

				time.Sleep(2 * time.Second)
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      "celery-broker-test-1",
				}, &celeryv4.CeleryBroker{})
				Expect(err).NotTo(HaveOccurred())

				time.Sleep(2 * time.Second)
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      "celery-broker-test-1-broker",
				}, &corev1.Pod{})
				Expect(err).NotTo(HaveOccurred())
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      "celery-broker-test-1-broker-service",
				}, &corev1.Service{})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
