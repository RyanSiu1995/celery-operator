package controllers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

var ctx = context.TODO()

var _ = Describe("Celery CRUD", func() {
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

		// Have the celery object created
		time.Sleep(2 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-1",
		}, &celeryv4.Celery{})
		Expect(err).NotTo(HaveOccurred())

		// Have the broker created
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-1-broker",
		}, &celeryv4.CeleryBroker{})
		Expect(err).NotTo(HaveOccurred())

		// Have two schedulers created
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-1-scheduler-1",
		}, &celeryv4.CeleryScheduler{})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-1-scheduler-2",
		}, &celeryv4.CeleryScheduler{})
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-1-worker-1",
		}, &celeryv4.CeleryWorker{})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-1-worker-2",
		}, &celeryv4.CeleryWorker{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should recreate the CRDs", func() {
		celerySpecInYaml, err := ioutil.ReadFile("../tests/fixtures/celery_2.yaml")
		Expect(err).NotTo(HaveOccurred())
		celeryObject := &celeryv4.Celery{}
		celerySpecInJSON, err := yaml.YAMLToJSON(celerySpecInYaml)
		Expect(err).NotTo(HaveOccurred())
		err = json.Unmarshal(celerySpecInJSON, celeryObject)
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Create(ctx, celeryObject)
		Expect(err).NotTo(HaveOccurred())

		// Have the celery object created
		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-2",
		}, &celeryv4.Celery{})
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.DeleteAllOf(ctx, &celeryv4.CeleryBroker{}, client.InNamespace("default"), client.MatchingLabels{"celery-app": "celery-test-2", "type": "broker"})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-2-broker",
		}, &celeryv4.CeleryBroker{})
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.DeleteAllOf(ctx, &celeryv4.CeleryScheduler{}, client.InNamespace("default"), client.MatchingLabels{"celery-app": "celery-test-2", "type": "scheduler"})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-2-scheduler-1",
		}, &celeryv4.CeleryScheduler{})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-2-scheduler-2",
		}, &celeryv4.CeleryScheduler{})
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.DeleteAllOf(ctx, &celeryv4.CeleryWorker{}, client.InNamespace("default"), client.MatchingLabels{"celery-app": "celery-test-2", "type": "worker"})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-2-worker-1",
		}, &celeryv4.CeleryWorker{})
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.Get(ctx, client.ObjectKey{
			Namespace: "default",
			Name:      "celery-test-2-worker-2",
		}, &celeryv4.CeleryWorker{})
		Expect(err).NotTo(HaveOccurred())
	})
})
