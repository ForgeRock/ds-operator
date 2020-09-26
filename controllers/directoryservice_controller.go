/*
   skeleton DS controller
*/

package controllers

import (
	"context"

	"github.com/pkg/errors"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
		log.Info("Registering finalizer for Directory Service", "name", ds.Name)
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

	// From https://engineering.pivotal.io/post/gp4k-kubebuilder-lessons/
	// In your mutate callback, you should surgically modify individual fields of the object. Don’t overwrite
	// large chunks of the object, or the whole object, as we tried to do initially.

	// This creates a stub sts with only the name/namespace set.
	// The CreateOrUpdate Method will then take this and fill it in the actual values (if the sts exists already)
	var sts apps.StatefulSet
	sts.Name = ds.Name
	sts.Namespace = ds.Namespace

	_, err := ctrl.CreateOrUpdate(ctx, r, &sts, func() error {
		// todo:
		// Fill in STS fields. If the object already exists this should only update fields.
		// ModifyStatefulSet(ds,&sts)
		log.Info("CreateorUpdate statefulset", "sts", sts)

		var err error
		// does the sts not exist yet? Is this the right check?
		if sts.CreationTimestamp.IsZero() {
			err = createDSStatefulSet(&ds, &sts)
			_ = controllerutil.SetControllerReference(&ds, &sts, r.Scheme)
			log.Info("Created New sts from template", "sts", sts)
		} else {
			// If the sts exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			err = updateDSStatefulSet(&ds, &sts)
		}

		log.Info("sts after update/create", "sts", sts)
		return err

	})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to CreateOrUpdate StateFulSet")
	}

	// create or update the service
	var svc v1.Service
	svc.Name = ds.Name
	svc.Namespace = ds.Namespace

	_, err = ctrl.CreateOrUpdate(ctx, r, &svc, func() error {
		log.Info("CreateorUpdate service", "svc", svc)

		var err error
		// does the service not exist yet?
		if svc.CreationTimestamp.IsZero() {
			err = createService(&ds, &svc)
			log.Info("Setting ownerref for service", "svc", svc.Name)
			_ = controllerutil.SetControllerReference(&ds, &svc, r.Scheme)
			log.Info("Created New sts from template", "sts", sts)
		} else {
			// If the sts exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			//err = updateService(&ds, &sts)
			log.Info("TODO: Handle update of ds service")
		}

		log.Info("svc after update/create", "svc", svc)
		return err
	})

	// Create secrets or references to secrets for passwords
	if len(ds.Spec.SecretReferencePasswords) > 0 {
		log.Info("Looking for secret reference", "secretReferencePasswords", ds.Spec.SecretReferencePasswords)

	}

	//// SECRETS ////

	// TODO: THe webhook can set Defaults, but for testing do we also want to set defaults here...
	// Defaults the name of the secret that contains the DS passwords.
	if len(ds.Spec.SecretReferencePasswords) == 0 {
		ds.Spec.SecretReferencePasswords = ds.Name + "-passwords-test"
	}
	var adminSecret v1.Secret
	adminSecret.Name = ds.Spec.SecretReferencePasswords
	adminSecret.Namespace = ds.Namespace

	_, err = ctrl.CreateOrUpdate(ctx, r, &adminSecret, func() error {
		if adminSecret.CreationTimestamp.IsZero() {
			createAdminSecret(&ds, &adminSecret)
			_ = controllerutil.SetControllerReference(&ds, &adminSecret, r.Scheme)

		} else {
			// todo: Do we allow changing the secret in any way?
			log.Info("TODO- update admin secret")
		}
		log.Info("Updated admin secret", "adminSecret", adminSecret)
		return nil
	})

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to CreateOrUpdate Service")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager stuff
func (r *DirectoryServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&directoryv1alpha1.DirectoryService{}).
		Complete(r)
}

// This function updates an existing statefulset to match settings in the custom resource
// TODO: What kinds of things should we update?
func updateDSStatefulSet(ds *directoryv1alpha1.DirectoryService, sts *apps.StatefulSet) error {

	sts.Spec.Replicas = ds.Spec.Replicas
	return nil
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
