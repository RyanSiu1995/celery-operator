/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	sysError "errors"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

// CelerySchedulerReconciler reconciles a CeleryScheduler object
type CelerySchedulerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=celery.celeryproject.org,resources=celeryschedulers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=celery.celeryproject.org,resources=celeryschedulers/status,verbs=get;update;patch

func (r *CelerySchedulerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log.WithValues("celeryscheduler", req.NamespacedName)

	// your logic here
	instance := &celeryv4.CeleryScheduler{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	// Handle the object creation
	existingPodList := &corev1.PodList{}
	err = r.Client.List(context.TODO(), existingPodList, client.MatchingLabels{
		"celery-app": instance.Name,
		"type":       "scheduler",
	})
	podList := instance.Generate(instance.Spec.Replicas - len(existingPodList.Items))
	for _, pod := range podList {
		found := &corev1.Pod{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			reqLogger.Info("Creating a new Scheduler pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			if err := r.Client.Create(context.TODO(), pod); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	err = r.Client.Status().Update(context.TODO(), instance)
	if err != nil {
		return ctrl.Result{}, sysError.New("Cannot update Scheduler status")
	}

	return ctrl.Result{}, nil
}

func (r *CelerySchedulerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&celeryv4.CeleryScheduler{}).
		Complete(r)
}
