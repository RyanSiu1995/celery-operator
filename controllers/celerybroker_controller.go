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

func (r *CeleryBrokerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log.WithValues("celerybroker", req.NamespacedName)

	// your logic here
	instance := &celeryv4.CeleryBroker{}
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
	if instance.Spec.Type == celeryv4.ExternalBroker {
		instance.Status.BrokerAddress = instance.Spec.BrokerAddress
	} else {
		pod, service, addr := instance.Generate()
		if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		found := &corev1.Pod{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Broker pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			if err := r.Client.Create(context.TODO(), pod); err != nil {
				return ctrl.Result{}, err
			}

			reqLogger.Info("Creating a new Broker service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			if err := r.Client.Create(context.TODO(), service); err != nil {
				return ctrl.Result{}, err
			}
		}
		instance.Status.BrokerAddress = addr
	}
	err = r.Client.Status().Update(context.TODO(), instance)
	if err != nil {
		return ctrl.Result{}, sysError.New("Cannot update broker status")
	}

	return ctrl.Result{}, nil
}

func (r *CeleryBrokerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&celeryv4.CeleryBroker{}).
		Complete(r)
}
