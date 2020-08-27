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

var _ = Describe("CeleryScheduler Creation", func() {
	Describe("CeleryScheulder Creation", func() {
		Context("Create a scheduler", func() {
			It("should have a single scheduler pod", func() {
				celeryschedulerSpecInYaml, err := ioutil.ReadFile("../tests/fixtures/celery_scheduler.yaml")
				Expect(err).NotTo(HaveOccurred())
				celeryschedulerObject := &celeryv4.CeleryScheduler{}
				celeryschedulerSpecInJSON, err := yaml.YAMLToJSON(celeryschedulerSpecInYaml)
				Expect(err).NotTo(HaveOccurred())
				err = json.Unmarshal(celeryschedulerSpecInJSON, celeryschedulerObject)
				Expect(err).NotTo(HaveOccurred())
				err = k8sClient.Create(ctx, celeryschedulerObject)
				Expect(err).NotTo(HaveOccurred())

				time.Sleep(2 * time.Second)
				err = k8sClient.Get(ctx, client.ObjectKey{
					Namespace: "default",
					Name:      "celery-scheduler-test-1",
				}, &celeryv4.CeleryScheduler{})
				Expect(err).NotTo(HaveOccurred())

				time.Sleep(2 * time.Second)
				podList := &corev1.PodList{}
				err = k8sClient.List(ctx, podList, client.MatchingLabels{
					"celery-app": "celery-scheduler-test-1",
					"type":       "scheduler",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(len(podList.Items)).To(Equal(2))
			})
		})
	})
})
