package controllers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

var ctx = context.Background()

var _ = Describe("Celery Creation", func() {
	Describe("Celery Creation", func() {
		Context("Create a broker and worker", func() {
			It("should have a single broker and worker", func() {
				celerySpecInYaml, err := ioutil.ReadFile("../tests/fixtures/celery.yaml")
				Expect(err).NotTo(HaveOccurred())
				celeryObject := &celeryv4.Celery{}
				celerySpecInJSON, err := yaml.YAMLToJSON(celerySpecInYaml)
				Expect(err).NotTo(HaveOccurred())
				err = json.Unmarshal(celerySpecInJSON, celeryObject)
				Expect(err).NotTo(HaveOccurred())
				err = k8sClient.Create(ctx, celeryObject)
				Expect(err).NotTo(HaveOccurred())

				time.Sleep(2 * time.Second)
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      "celery-test-1",
				}, &celeryv4.Celery{})
				Expect(err).NotTo(HaveOccurred())

				time.Sleep(2 * time.Second)
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      "celery-test-1-broker",
				}, &celeryv4.CeleryBroker{})
				Expect(err).NotTo(HaveOccurred())
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      "celery-test-1-broker-broker",
				}, &corev1.Pod{})
				Expect(err).NotTo(HaveOccurred())
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      "celery-test-1-broker-broker-service",
				}, &corev1.Service{})
				Expect(err).NotTo(HaveOccurred())
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      "celery-test-1-worker-deployment-0",
				}, &appv1.Deployment{})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
