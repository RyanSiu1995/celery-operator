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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

// CeleryReconciler reconciles a Celery object
type CeleryReconciler Reconciler

// +kubebuilder:rbac:groups=celery.celeryproject.org,resources=celeries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=celery.celeryproject.org,resources=celeries/status,verbs=get;update;patch

func (r *CeleryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	reqLogger := r.Log.WithValues("celery", req.NamespacedName)

	//
	// Fetch the Celery instance
	//
	instance := &celeryv4.Celery{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
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

	//
	// Handle Broker object
	//
	broker := instance.GenerateBroker()
	found := &celeryv4.CeleryBroker{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: broker.Name, Namespace: broker.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(instance, broker, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.Info("Creating a new CeleryBroker", "CeleryBroker.Namespace", broker.Namespace, "CeleryBroker.Name", broker.Name)
		if err := r.Client.Create(ctx, broker); err != nil {
			return ctrl.Result{}, err
		}
	}

	//
	// Handle Schedulers object
	//
	schedulers := instance.GenerateSchedulers()
	for _, scheduler := range schedulers {
		found := &celeryv4.CeleryScheduler{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: scheduler.Name, Namespace: scheduler.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			if err := controllerutil.SetControllerReference(instance, scheduler, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			reqLogger.Info("Creating a new CeleryScheduler", "CeleryScheduler.Namespace", scheduler.Namespace, "CeleryScheduler.Name", scheduler.Name)
			if err := r.Client.Create(ctx, scheduler); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	//
	// Handle workers
	//
	workers := instance.GenerateWorkers()
	for _, worker := range workers {
		found := &celeryv4.CeleryWorker{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: worker.Name, Namespace: worker.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			if err := controllerutil.SetControllerReference(instance, worker, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			reqLogger.Info("Creating a new CeleryWorker", "CeleryWorker.Namespace", worker.Namespace, "CeleryWorker.Name", worker.Name)
			if err := r.Client.Create(ctx, worker); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *CeleryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&celeryv4.Celery{}).
		Owns(&celeryv4.CeleryWorker{}).
		Owns(&celeryv4.CeleryScheduler{}).
		Owns(&celeryv4.CeleryBroker{}).
		Complete(r)
}
