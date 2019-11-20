package mailgunwebhook

import (
	"context"
	"reflect"

	"time"

	"github.com/mailgun/mailgun-go/v3"
	mailgunv1alpha1 "github.com/whyseco/mailgun-operator/pkg/apis/mailgun/v1alpha1"
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

var log = logf.Log.WithName("controller_mailgunwebhook")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MailgunWebhook Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMailgunWebhook{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("mailgunwebhook-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MailgunWebhook
	err = c.Watch(&source.Kind{Type: &mailgunv1alpha1.MailgunWebhook{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMailgunWebhook implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMailgunWebhook{}

// ReconcileMailgunWebhook reconciles a MailgunWebhook object
type ReconcileMailgunWebhook struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MailgunWebhook object and makes changes based on the state read
// and what is in the MailgunWebhook.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMailgunWebhook) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MailgunWebhook")

	// Fetch the MailgunWebhook instance
	instance := &mailgunv1alpha1.MailgunWebhook{}
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

	err = checkWebhook(ctx, mg, "opened", instance.Spec.Opened)

	return reconcile.Result{}, err
}

func checkWebhook(ctx context.Context, mg *mailgun.MailgunImpl, kind string, urls []string) error {
	currentUrls, err := mg.GetWebhook(ctx, kind)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(currentUrls, urls) {
		if err := mg.UpdateWebhook(ctx, kind, urls); err != nil {
			status := mailgun.GetStatusFromErr(err)

			if status == 404 {
				err = mg.CreateWebhook(ctx, kind, urls)
			}
			return err
		}
	}
	return nil
}
