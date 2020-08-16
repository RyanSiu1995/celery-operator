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
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	celeryprojectv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

// CeleryReconciler reconciles a Celery object
type CeleryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=celeryproject.org,resources=celeries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=celeryproject.org,resources=celeries/status,verbs=get;update;patch

func (r *CeleryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log.WithValues("celery", req.NamespacedName)

	//
	// Fetch the Celery instance
	//
	instance := &celeryprojectv4.Celery{}
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

	//
	// Handle Broker object
	//
	brokerDeployment, brokerService, err := instance.GetBroker()
	if err != nil {
		return ctrl.Result{}, err
	}

	// Handle the Deployment and Service Creation for the Broker
	if brokerDeployment != nil && brokerService != nil {
		// Set Celery instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, brokerDeployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := controllerutil.SetControllerReference(instance, brokerService, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		found := &appv1.Deployment{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: brokerDeployment.Name, Namespace: brokerDeployment.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Broker deployment", "Deployment.Namespace", brokerDeployment.Namespace, "Deployment.Name", brokerDeployment.Name)
			if err := r.Client.Create(context.TODO(), brokerDeployment); err != nil {
				return ctrl.Result{}, err
			}

			reqLogger.Info("Creating a new Broker service", "Service.Namespace", brokerService.Namespace, "Service.Name", brokerService.Name)
			if err := r.Client.Create(context.TODO(), brokerService); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		return ctrl.Result{}, sysError.New("Cannot generate the service and deployment successfully...")
	}

	// Update Broker Inforamtion after setting up
	err = r.Client.Status().Update(context.TODO(), instance)
	if err != nil {
		return ctrl.Result{}, sysError.New("Cannot update broker status")
	}

	//
	// Handle workers
	//
	// workerDeployments, err := instance.GetWorkers()
	// for _, workerDeployment := range workerDeployments {
	// 	reqLogger.Info("Handling the workerDeployment", workerDeployment)
	// 	if err := controllerutil.SetControllerReference(instance, workerDeployment, r.Scheme); err != nil {
	// 		return ctrl.Result{}, err
	// 	}
	// 	found := &appv1.Deployment{}
	// 	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: workerDeployment.Name, Namespace: workerDeployment.Namespace}, found)
	// 	if err != nil && errors.IsNotFound(err) {
	// 		reqLogger.Info("Creating a new worker deployment", "Deployment.Namespace", workerDeployment.Namespace, "Deployment.Name", workerDeployment.Name)
	// 		if err := r.Client.Create(context.TODO(), brokerDeployment); err != nil {
	// 			return ctrl.Result{}, err
	// 		}
	// 	}
	// }
	return ctrl.Result{}, nil
}

func (r *CeleryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&celeryprojectv4.Celery{}).
		Complete(r)
}
