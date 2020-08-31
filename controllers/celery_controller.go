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
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	existingBroker := &celeryv4.CeleryBroker{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: broker.Name, Namespace: broker.Namespace}, existingBroker)
	if err != nil && errors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(instance, broker, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.Info("Creating a new CeleryBroker", "CeleryBroker.Namespace", broker.Namespace, "CeleryBroker.Name", broker.Name)
		if err := r.Client.Create(ctx, broker); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if !existingBroker.Equal(broker) {
			reqLogger.Info("Updating CeleryBroker",
				"OldCeleryBroker.Namespace", existingBroker.Namespace,
				"OldCeleryBroker.Name", existingBroker.Name,
				"OldCeleryBroker.Spec", existingBroker.Spec,
				"NewCeleryBroker.Namespace", broker.Namespace,
				"NewCeleryBroker.Name", broker.Name,
				"NewCeleryBroker.Spec", broker.Spec)
			existingBroker.Spec = broker.Spec
			if err := r.Client.Update(ctx, existingBroker); err != nil {
				reqLogger.Error(err, "Error in patching the broker")
				return ctrl.Result{}, err
			}
		}
	}

	//
	// Handle Schedulers object
	//
	schedulers := instance.GenerateSchedulers()
	existingSchedulers := &celeryv4.CelerySchedulerList{}
	err = r.Client.List(ctx, existingSchedulers, client.MatchingLabels{
		"celery-app": instance.Name,
		"type":       "scheduler",
	})
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: REQUEUE_TIMEOUT}, err
	}
	existing := len(existingSchedulers.Items)
	reqLogger.Info("Checking the difference in schedulers", "existing", existing, "target", len(schedulers))
	if existing > len(schedulers) {
		schedulersToBeDeleted := existingSchedulers.Items[:existing-len(schedulers)]
		for _, s := range schedulersToBeDeleted {
			reqLogger.Info("Deleteing the scheduler", "CeleryScheduler.Namespace", s.Namespace, "CeleryScheduler.Name", s.Name)
			err = r.Client.Delete(ctx, &s)
			if err != nil {
				return ctrl.Result{Requeue: true, RequeueAfter: REQUEUE_TIMEOUT}, err
			}
		}
	}
	err = r.Client.List(ctx, existingSchedulers, client.MatchingLabels{
		"celery-app": instance.Name,
		"type":       "scheduler",
	})
	for i, scheduler := range schedulers {
		found := &celeryv4.CeleryScheduler{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: scheduler.Name, Namespace: scheduler.Namespace}, found)
		if i < existing {
			if err != nil && errors.IsNotFound(err) {
				if err := controllerutil.SetControllerReference(instance, scheduler, r.Scheme); err != nil {
					return ctrl.Result{}, err
				}
				reqLogger.Info("Creating a new CeleryScheduler", "CeleryScheduler.Namespace", scheduler.Namespace, "CeleryScheduler.Name", scheduler.Name)
				if err := r.Client.Create(ctx, scheduler); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				reqLogger.Info("Going to patch with name spec", "CeleryScheduler.Namespace", scheduler.Namespace, "CeleryScheduler.Name", scheduler.Name, "CeleryScheduler.Spec", scheduler.Spec)
				if err := r.Client.Patch(ctx, &existingSchedulers.Items[i], client.MergeFrom(scheduler)); err != nil {
					return ctrl.Result{Requeue: true}, err
				}
			}
		} else {
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
	}

	//
	// Handle workers
	//
	existingWorkers := &celeryv4.CeleryWorkerList{}
	err = r.Client.List(ctx, existingWorkers, client.MatchingLabels{
		"celery-app": instance.Name,
		"type":       "worker",
	})
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: REQUEUE_TIMEOUT}, err
	}
	existing = len(existingWorkers.Items)
	workers := instance.GenerateWorkers()
	reqLogger.Info("Checking the difference in workers", "existing", existing, "target", len(workers))
	if existing > len(workers) {
		workersToBeDeleted := existingWorkers.Items[:existing-len(workers)]
		for _, s := range workersToBeDeleted {
			reqLogger.Info("Deleteing the worker", "CeleryWorker.Namespace", s.Namespace, "CeleryWorker.Name", s.Name)
			err = r.Client.Delete(ctx, &s)
			if err != nil {
				return ctrl.Result{Requeue: true, RequeueAfter: REQUEUE_TIMEOUT}, err
			}
		}
	}
	err = r.Client.List(ctx, existingWorkers, client.MatchingLabels{
		"celery-app": instance.Name,
		"type":       "worker",
	})
	for i, worker := range workers {
		found := &celeryv4.CeleryWorker{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: worker.Name, Namespace: worker.Namespace}, found)
		if i < existing {
			if err != nil && errors.IsNotFound(err) {
				if err := controllerutil.SetControllerReference(instance, worker, r.Scheme); err != nil {
					return ctrl.Result{}, err
				}
				reqLogger.Info("Creating a new CeleryWorker", "CeleryWorker.Namespace", worker.Namespace, "CeleryWorker.Name", worker.Name)
				if err := r.Client.Create(ctx, worker); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				reqLogger.Info("Going to patch with name spec", "CeleryWorker.Namespace", worker.Namespace, "CeleryWorker.Name", worker.Name, "CeleryWorker.Spec", worker.Spec)
				if err := r.Client.Patch(ctx, &existingWorkers.Items[i], client.MergeFrom(worker)); err != nil {
					return ctrl.Result{Requeue: true}, err
				}
			}
		} else {
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
