/*
   skeleton DS controller
*/

package controllers

import (
	"context"
	"time"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DirectoryServiceReconciler reconciles a DirectoryService object
type DirectoryServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Reconcile loop for DS controller
// Add in all the RBAC permissions that a DS controller needs. StatefulSets, etc.
// +kubebuilder:rbac:groups=directory.forgerock.com,resources=directoryservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=directory.forgerock.com,resources=directoryservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
func (r *DirectoryServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	// This adds the log data to every log line
	var log = r.Log.WithValues("directoryservice", req.NamespacedName)

	log.Info("Started")

	var ds directoryv1alpha1.DirectoryService

	// Load the DirectoryService
	if err := r.Get(ctx, req.NamespacedName, &ds); err != nil {
		log.Info("unable to fetch DirectorService. You can probably ignore this..")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// finalizer hooks..
	// This registers finalizers for deleting the object
	myFinalizerName := "directory.finalizers.forgerock.com"

	// examine DeletionTimestamp to determine if object is under deletion
	if ds.ObjectMeta.DeletionTimestamp.IsZero() {
		log.V(3).Info("Registering finalizer for Directory Service", "name", ds.Name)
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(ds.GetFinalizers(), myFinalizerName) {
			ds.SetFinalizers(append(ds.GetFinalizers(), myFinalizerName))
			if err := r.Update(context.Background(), &ds); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		log.Info("Deleting Directory Service", "name", ds.Name)
		// The object is being deleted
		if containsString(ds.GetFinalizers(), myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(&ds); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			ds.SetFinalizers(removeString(ds.GetFinalizers(), myFinalizerName))
			if err := r.Update(context.Background(), &ds); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	//// SECRETS ////
	if res, err := r.reconcileSecrets(ctx, &ds); err != nil {
		return res, err
	}

	//// StatefulSets ////
	if res, err := r.reconcileSTS(ctx, &ds); err != nil {
		return res, err
	}

	//// Services ////
	if res, err := r.reconcileService(ctx, &ds); err != nil {
		return res, err
	}

	// TODO: Remove this...
	// Just for testing
	if ds.CreationTimestamp.IsZero() {
		log.Info("ohh noes... should not happen")

	} else {
		log.Info("Update ds status")
		now := time.Now()
		ds.Status.LastUpdate.Seconds = now.Unix()
	}

	log.Info("DS object", "directoryserver", ds)

	// Update the status of our ds
	if err := r.Status().Update(ctx, &ds); err != nil {
		log.Error(err, "unable to update Directory status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 20}, nil
}

// SetupWithManager stuff
func (r *DirectoryServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&directoryv1alpha1.DirectoryService{}).
		Complete(r)
}

func (r *DirectoryServiceReconciler) deleteExternalResources(ds *directoryv1alpha1.DirectoryService) error {
	//
	// delete any external resources associated with the ds set
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple times for same object.
	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
