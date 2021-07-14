/*

Copyright 2020 ForgeRock AS.

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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
)

// DirectoryBackupReconciler reconciles a DirectoryBackup object
type DirectoryBackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=directory.forgerock.io,resources=directorybackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=directory.forgerock.io,resources=directorybackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=directory.forgerock.io,resources=directorybackups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DirectoryBackup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DirectoryBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var logger = log.FromContext(ctx)

	// Reconcile the backup target PVC that holds the backup

	logger.Info("Reconciling directorybackup")

	// fetch the DirectoryBackup object
	var db directoryv1alpha1.DirectoryBackup

	// Load the DirectoryService
	if err := r.Get(ctx, req.NamespacedName, &db); err != nil {
		logger.Info("Unable to fetch directoryservice - it is in the process of being deleted. This is OK")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	/// Recon the backup target PVC that holds the backup

	var pvc v1.PersistentVolumeClaim
	pvc.Name = db.Spec.BackupPVC.Name
	pvc.Namespace = db.GetNamespace()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		logger.Info("CreateorUpdate backup pvc", "pvc", pvc)

		var err error
		// does the sts not exist yet?
		if pvc.CreationTimestamp.IsZero() {
			logger.Info("Creating Backup PVC", "backupPVC", pvc.Name)
			var x *v1.PersistentVolumeClaim = &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: db.GetName(),
					// todo: Labels, etc.
					Namespace: db.GetNamespace(),
					Labels:    createLabels(db.GetName(), nil),
					Annotations: map[string]string{
						"pv.beta.kubernetes.io/gid": "0",
					},
				},
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceName(v1.ResourceStorage): resource.MustParse(db.Spec.BackupPVC.Size),
						},
					},
					StorageClassName: &db.Spec.BackupPVC.StorageClassName,
				},
			}

			x.DeepCopyInto(&pvc)

			_ = controllerutil.SetControllerReference(&db, &pvc, r.Scheme)
			//
		} else {
			// If the sts exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			logger.Info("Updating backup pvc", "pvc", pvc)
		}

		logger.V(8).Info("sts after update/create", "pvc", pvc)
		return err
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	//  Snapshot the target PVC
	

	logger.Info("Done")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DirectoryBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&directoryv1alpha1.DirectoryBackup{}).
		Complete(r)
}
