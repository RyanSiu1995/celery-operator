package celery

import (
	"context"
	sysError "errors"

	celeryprojectv4 "github.com/RyanSiu1995/celery-operator/pkg/apis/celeryproject/v4"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_celery")

// Add creates a new Celery Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCelery{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("celery-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Celery
	err = c.Watch(&source.Kind{Type: &celeryprojectv4.Celery{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Celery
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &celeryprojectv4.Celery{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCelery implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCelery{}

// ReconcileCelery reconciles a Celery object
type ReconcileCelery struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Celery object and makes changes based on the state read
// and what is in the Celery.Spec
func (r *ReconcileCelery) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Celery")

	// Fetch the Celery instance
	instance := &celeryprojectv4.Celery{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Broker object
	var brokerAddress string
	if instance.Spec.Broker.Type == celeryprojectv4.ExternalBroker {
		if &instance.Spec.Broker.BrokerAddress == nil {
			return reconcile.Result{}, sysError.New("Broker address hasn't been set")
		}
		brokerAddress = instance.Spec.Broker.BrokerAddress
	} else {
		brokerDeployment, brokerService, address := generateBroker(instance)
		// Set Celery instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, brokerDeployment, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		if err := controllerutil.SetControllerReference(instance, brokerService, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		found := &appv1.Deployment{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: brokerDeployment.Name, Namespace: brokerDeployment.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Broker deployment", "Deployment.Namespace", brokerDeployment.Namespace, "Deployment.Name", brokerDeployment.Name)
			if err := r.client.Create(context.TODO(), brokerDeployment); err != nil {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating a new Broker service", "Service.Namespace", brokerDeployment.Namespace, "Pod.Name", brokerDeployment.Name)
			if err := r.client.Create(context.TODO(), brokerService); err != nil {
				return reconcile.Result{}, err
			}
		}
		// TODO Check if the service has been created
		brokerAddress = address
	}
	instance.Status.BrokerAddress = brokerAddress
	err = r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, sysError.New("Cannot update broker status")
	}
	reqLogger.Info("Broker information has been collected")

	// Define a new Scheduler object
	schedulerDeployment := generateScheduler(instance, brokerAddress)
	if err := controllerutil.SetControllerReference(instance, schedulerDeployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	found := &appv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: schedulerDeployment.Name, Namespace: schedulerDeployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Scheduler deployment", "Deployment.Namespace", schedulerDeployment.Namespace, "Deployment.Name", schedulerDeployment.Name)
		if err := r.client.Create(context.TODO(), schedulerDeployment); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
