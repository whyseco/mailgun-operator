package mailgundomain

import (
	"context"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/thoas/go-funk"
	mailgunv1alpha1 "github.com/whyseco/mailgun-operator/pkg/apis/mailgun/v1alpha1"
	"github.com/whyseco/mailgun-operator/pkg/helpers"
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

var log = logf.Log.WithName("controller_mailgundomain")

const domainFinalizer = "finalizer.domain.mailgun.com"

// Add creates a new MailgunDomain Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMailgunDomain{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("mailgundomain-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MailgunDomain
	err = c.Watch(&source.Kind{Type: &mailgunv1alpha1.MailgunDomain{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner MailgunDomain
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mailgunv1alpha1.MailgunDomain{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMailgunDomain implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMailgunDomain{}

// ReconcileMailgunDomain reconciles a MailgunDomain object
type ReconcileMailgunDomain struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MailgunDomain object and makes changes based on the state read
// and what is in the MailgunDomain.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMailgunDomain) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MailgunDomain")

	// Fetch the MailgunDomain instance
	instance := &mailgunv1alpha1.MailgunDomain{}
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

	apiKey := ""
	if apiKey, err = helpers.GetApiKey(ctx, reqLogger, r.client, instance.Spec.SecretName, request.Namespace); err != nil {
		return reconcile.Result{}, err
	}

	mg := mailgun.NewMailgun(instance.Spec.Domain, apiKey)

	// Check if the Domain instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isDomainMarkedToBeDeleted := instance.GetDeletionTimestamp() != nil
	if isDomainMarkedToBeDeleted {
		if funk.Contains(instance.GetFinalizers(), domainFinalizer) {
			// Run finalization logic for domainFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeMailgunDomain(reqLogger, instance, mg, ctx); err != nil {
				return reconcile.Result{}, err
			}

			// Remove domainFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			finalizers := funk.FilterString(instance.GetFinalizers(), func(x string) bool { return x != domainFinalizer })
			instance.SetFinalizers(finalizers)
			err := r.client.Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// Add finalizer for this CR
	if !funk.Contains(instance.GetFinalizers(), domainFinalizer) {
		if err := r.addFinalizer(reqLogger, instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	domain, err := mg.GetDomain(ctx, instance.Spec.Domain)
	if err != nil {
		if status := mailgun.GetStatusFromErr(err); status != http.StatusNotFound {
			return reconcile.Result{}, err
		}
		domain, err = mg.CreateDomain(ctx, instance.Spec.Domain,
			&mailgun.CreateDomainOptions{
				SpamAction:         mailgun.SpamAction(instance.Spec.SpamAction),
				Password:           instance.Spec.Password,
				Wildcard:           instance.Spec.Wildcard,
				ForceDKIMAuthority: instance.Spec.ForceDKIMAuthority,
				DKIMKeySize:        instance.Spec.DKIMKeySize,
				IPS:                instance.Spec.IPS,
			})
	}

	if err == nil {
		// Update status after creation
		instance.Status.DomainState = domain.Domain.State
		instance.Status.SendingDnsRecord = mapDNSRecords(domain.SendingDNSRecords)
		instance.Status.ReceivingDnsRecord = mapDNSRecords(domain.ReceivingDNSRecords)
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update MailgunDomain status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, err
}

func (r *ReconcileMailgunDomain) finalizeMailgunDomain(reqLogger logr.Logger, m *mailgunv1alpha1.MailgunDomain,
	mg *mailgun.MailgunImpl, ctx context.Context) error {

	err := mg.DeleteDomain(ctx, m.Spec.Domain)

	if status := mailgun.GetStatusFromErr(err); status == http.StatusNotFound {
		return nil
	}

	reqLogger.Info("Successfully finalized MailgunDomain")
	return err
}

func (r *ReconcileMailgunDomain) addFinalizer(reqLogger logr.Logger, m *mailgunv1alpha1.MailgunDomain) error {
	reqLogger.Info("Adding Finalizer for the MailgunDomain")
	m.SetFinalizers(append(m.GetFinalizers(), domainFinalizer))

	// Update CR
	err := r.client.Update(context.TODO(), m)
	if err != nil {
		reqLogger.Error(err, "Failed to update MailgunDomain with finalizer")
		return err
	}
	return nil
}

func mapDNSRecords(records []mailgun.DNSRecord) []mailgunv1alpha1.MailgunDomainDnsRecord {
	newRecords := make([]mailgunv1alpha1.MailgunDomainDnsRecord, len(records))
	for i, _ := range records {
		newRecords[i] = mailgunv1alpha1.MailgunDomainDnsRecord{
			RecordType: records[i].RecordType,
			Priority:   records[i].Priority,
			Valid:      records[i].Valid,
			Name:       records[i].Name,
			Value:      records[i].Value,
		}
	}
	return newRecords
}
