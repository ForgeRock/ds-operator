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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	snapshot "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
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

	/// Create/update the backup target PVC that holds the backup
	pvc, err := createPVC(ctx, r.Client, db.Spec.BackupPVC.Name, db.GetNamespace(), db.Spec.BackupPVC.Size, db.Spec.BackupPVC.StorageClassName, "")

	if err != nil {
		logger.Error(err, "PVC claim creation failed", "pvcName", db.Spec.BackupPVC.Name)
		return ctrl.Result{}, err
	}

	_ = controllerutil.SetControllerReference(&db, &pvc, r.Scheme)

	//  Create a Snapshot of the target PVC
	var snap snapshot.VolumeSnapshot
	snap.Name = "db-" + db.Spec.ClaimToBackup
	snap.Namespace = db.GetNamespace()

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &snap, func() error {
		logger.V(8).Info("CreateorUpdate snapshot", "name", snap.GetName())

		// does the snap not exist yet?
		if snap.CreationTimestamp.IsZero() {
			snap.ObjectMeta.Labels = createLabels(snap.GetName(), nil)
			//snap.Annotations = map[string]string{"directory.forgerock.io/lastSnapshotTime": strconv.Itoa(int(now))}
			snap.Spec = snapshot.VolumeSnapshotSpec{
				VolumeSnapshotClassName: &db.Spec.VolumeSnapshotClassName,
				Source:                  snapshot.VolumeSnapshotSource{PersistentVolumeClaimName: &db.Spec.ClaimToBackup}}

			_ = controllerutil.SetControllerReference(&db, &snap, r.Scheme)
		} else {
			logger.Info("Snapshot should not already exist. Report this error", "snapshot", snap)
		}

		return nil
	})

	// Create the Pod/Job that runs the LDIF export

	logger.Info("Done")

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *DirectoryBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&directoryv1alpha1.DirectoryBackup{}).
		Complete(r)
}
