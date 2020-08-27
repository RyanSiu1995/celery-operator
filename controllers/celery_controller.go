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

	celeryv4 "github.com/RyanSiu1995/celery-operator/api/v4"
)

// CeleryReconciler reconciles a Celery object
type CeleryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=celery.celeryproject.org,resources=celeries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=celery.celeryproject.org,resources=celeries/status,verbs=get;update;patch

func (r *CeleryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Log.WithValues("celery", req.NamespacedName)

	//
	// Fetch the Celery instance
	//
	instance := &celeryv4.Celery{}
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
	broker := instance.GenerateBroker()
	if err := controllerutil.SetControllerReference(instance, broker, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	found := &celeryv4.CeleryBroker{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: broker.Name, Namespace: broker.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Broker", "CeleryBroker.Namespace", broker.Namespace, "CeleryBroker.Name", broker.Name)
		if err := r.Client.Create(context.TODO(), broker); err != nil {
			return ctrl.Result{}, err
		}
	}

	//
	// Handle workers
	//
	workerDeployments, err := instance.GetWorkers()
	if err != nil {
		return ctrl.Result{}, sysError.New("Cannot create the worker deployment")
	}
	for _, workerDeployment := range workerDeployments {
		if err := controllerutil.SetControllerReference(instance, workerDeployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		found := &appv1.Deployment{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: workerDeployment.Name, Namespace: workerDeployment.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new worker deployment", "Deployment.Namespace", workerDeployment.Namespace, "Deployment.Name", workerDeployment.Name)
			if err := r.Client.Create(context.TODO(), workerDeployment); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *CeleryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&celeryv4.Celery{}).
		Complete(r)
}
