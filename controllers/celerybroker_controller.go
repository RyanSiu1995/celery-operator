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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

// CeleryBrokerReconciler reconciles a CeleryBroker object
type CeleryBrokerReconciler Reconciler

// +kubebuilder:rbac:groups=celery.celeryproject.org,resources=celerybrokers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=celery.celeryproject.org,resources=celerybrokers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pod,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pod/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=service,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=service/status,verbs=get

func (r *CeleryBrokerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	reqLogger := r.Log.WithValues("celerybroker", req.NamespacedName)

	// your logic here
	instance := &celeryv4.CeleryBroker{}
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
	reqLogger.Info("Getting the spec of broker", "Broker.Namespace", instance.Namespace, "Broker.Name", instance.Name, "Broker.Spec", instance.Spec)
	// Handle the object creation
	if instance.Spec.Type == celeryv4.ExternalBroker {
		instance.Status.BrokerAddress = instance.Spec.BrokerAddress
		pod, service, _ := instance.Generate()
		found := &corev1.Pod{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
		if err == nil {
			if err = r.Client.Delete(ctx, found); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Client.Delete(ctx, service); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		pod, service, addr := instance.Generate()
		found := &corev1.Pod{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			reqLogger.Info("Creating a new Broker pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			if err := r.Client.Create(ctx, pod); err != nil {
				return ctrl.Result{}, err
			}

			reqLogger.Info("Creating a new Broker service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			if err := r.Client.Create(ctx, service); err != nil {
				return ctrl.Result{}, err
			}
		}
		instance.Status.BrokerAddress = addr
	}
	err = r.Client.Status().Update(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CeleryBrokerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&celeryv4.CeleryBroker{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
