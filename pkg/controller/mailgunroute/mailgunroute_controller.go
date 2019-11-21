package mailgunroute

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/thoas/go-funk"
	mailgunv1alpha1 "github.com/whyseco/mailgun-operator/pkg/apis/mailgun/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_mailgunroute")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MailgunRoute Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMailgunRoute{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("mailgunroute-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MailgunRoute
	err = c.Watch(&source.Kind{Type: &mailgunv1alpha1.MailgunRoute{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner MailgunRoute
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mailgunv1alpha1.MailgunRoute{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMailgunRoute implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMailgunRoute{}

// ReconcileMailgunRoute reconciles a MailgunRoute object
type ReconcileMailgunRoute struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

const routeFinalizer = "finalizer.route.mailgun.com"

// Reconcile reads that state of the cluster for a MailgunRoute object and makes changes based on the state read
// and what is in the MailgunRoute.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMailgunRoute) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MailgunRoute")

	// Fetch the MailgunRoute instance
	instance := &mailgunv1alpha1.MailgunRoute{}
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	mg := mailgun.NewMailgun(instance.Spec.Domain, instance.Spec.ApiKey)
	// Check if the Route instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isRouteMarkedToBeDeleted := instance.GetDeletionTimestamp() != nil
	if isRouteMarkedToBeDeleted {
		if funk.Contains(instance.GetFinalizers(), routeFinalizer) {
			// Run finalization logic for routeFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeMailgunRoute(reqLogger, instance, mg, ctx); err != nil {
				return reconcile.Result{}, err
			}

			// Remove routeFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			finalizers := funk.FilterString(instance.GetFinalizers(), func(x string) bool { return x != routeFinalizer })
			instance.SetFinalizers(finalizers)
			err := r.client.Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// Add finalizer for this CR
	if !funk.Contains(instance.GetFinalizers(), routeFinalizer) {
		if err := r.addFinalizer(reqLogger, instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	route := mailgun.Route{
		Priority:    instance.Spec.Priority,
		Description: instance.Spec.Description,
		Expression:  instance.Spec.Expression,
		Actions:     instance.Spec.Actions,
	}

	if len(instance.Status.Id) == 0 {
		newRoute, err := mg.CreateRoute(ctx, route)
		if err == nil {
			// Update status after creation
			instance.Status.Id = newRoute.Id
			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				reqLogger.Error(err, "Failed to update MailgunRoute status")
				return reconcile.Result{}, err
			}
		}
	} else {
		_, err = mg.UpdateRoute(ctx, instance.Status.Id, route)
	}

	return reconcile.Result{}, err
}

func (r *ReconcileMailgunRoute) finalizeMailgunRoute(reqLogger logr.Logger, m *mailgunv1alpha1.MailgunRoute,
	mg *mailgun.MailgunImpl, ctx context.Context) error {

	if len(m.Status.Id) != 0 {
		err := mg.DeleteRoute(ctx, m.Status.Id)
		reqLogger.Info("Successfully finalized MailgunRoute")
		return err
	}
	return nil
}

func (r *ReconcileMailgunRoute) addFinalizer(reqLogger logr.Logger, m *mailgunv1alpha1.MailgunRoute) error {
	reqLogger.Info("Adding Finalizer for the MailgunRoute")
	m.SetFinalizers(append(m.GetFinalizers(), routeFinalizer))

	// Update CR
	err := r.client.Update(context.TODO(), m)
	if err != nil {
		reqLogger.Error(err, "Failed to update MailgunRoute with finalizer")
		return err
	}
	return nil
}
