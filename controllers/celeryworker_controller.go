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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

// CeleryWorkerReconciler reconciles a CeleryWorker object
type CeleryWorkerReconciler Reconciler

// +kubebuilder:rbac:groups=celery.celeryproject.org,resources=celeryworkers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=celery.celeryproject.org,resources=celeryworkers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pod,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pod/status,verbs=get

func (r *CeleryWorkerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	reqLogger := r.Log.WithValues("celeryworker", req.NamespacedName)

	instance := &celeryv4.CeleryWorker{}
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
	// Handle the object creation
	existingPodList := &corev1.PodList{}
	err = r.Client.List(ctx, existingPodList, client.MatchingLabels{
		"celery-app": instance.Name,
		"type":       "worker",
	})

	// If there is an update compared to existing spec, recreate all pods
	if !instance.IsUpToDate(existingPodList.Items) {
		reqLogger.Info("The spec has been updated...Recreating all the pods...")
		podList := instance.Generate(instance.Spec.Replicas)
		for _, pod := range existingPodList.Items {
			reqLogger.Info("Deleteing the old Worker pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			if err := r.Client.Delete(ctx, &pod); err != nil {
				return ctrl.Result{}, err
			}
		}
		for _, pod := range podList {
			reqLogger.Info("Creating a new Worker pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Client.Create(ctx, pod); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	replicaDiff := instance.Spec.Replicas - len(existingPodList.Items)
	if replicaDiff >= 0 {
		// If the desired replicas is smaller than existing old, create pods
		podList := instance.Generate(replicaDiff)
		for _, pod := range podList {
			found := &corev1.Pod{}
			err = r.Client.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {
				reqLogger.Info("Creating a new Worker pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
				if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
					return ctrl.Result{}, err
				}
				if err := r.Client.Create(ctx, pod); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	} else {
		// If the desired replicas is larger than existing old, delete pods
		podList := existingPodList.Items[:(replicaDiff * -1)]
		for _, pod := range podList {
			found := &corev1.Pod{}
			err = r.Client.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
			if err != nil {
				reqLogger.Info("Creating a new Worker pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
				return ctrl.Result{}, err
			} else {
				if err = r.Client.Delete(ctx, found); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *CeleryWorkerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&celeryv4.CeleryWorker{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
