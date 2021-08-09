/*

Copyright 2021 ForgeRock AS.

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

	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	kbatch "k8s.io/api/batch/v1"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	snapshot "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
)

// DirectoryBackupReconciler reconciles a DirectoryBackup object
type DirectoryBackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Note the rbac rules are consolidated on the directoryservice_controller.go. These are for reference only
//
//+kubebuilder:rbac:groups=directory.forgerock.io,resources=directorybackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=directory.forgerock.io,resources=directorybackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=directory.forgerock.io,resources=directorybackups/finalizers,verbs=update

//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DirectoryBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var log = klog.FromContext(ctx)

	log.Info("Reconciling directorybackup")

	// fetch the DirectoryBackup object
	var db directoryv1alpha1.DirectoryBackup

	// Load the DirectoryBackup object
	if err := r.Get(ctx, req.NamespacedName, &db); err != nil {
		log.Info("Unable to fetch DirectoryBackup - it is in the process of being deleted. This is OK")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List in progress backup job to update status. The backup job has the same name as this object
	var backupJob kbatch.Job

	// Ignore if job is not found - it might not yet be created.
	if err := r.Get(ctx, req.NamespacedName, &backupJob); client.IgnoreNotFound(err) != nil {
		log.Error(err, "cant fetch job status")
		return ctrl.Result{}, err
	}

	// If the job has started (and possibly finished) update our status sub resource
	if !backupJob.ObjectMeta.CreationTimestamp.IsZero() {
		// update CRD status
		db.Status.StartTimestamp = &backupJob.ObjectMeta.CreationTimestamp
		if backupJob.Status.CompletionTime != nil {
			db.Status.CompletionTimestamp = backupJob.Status.CompletionTime
		}
	}

	// update status
	if err := r.Status().Update(ctx, &db); err != nil {
		log.Error(err, "unable to update DirectoryBackup status")
		return ctrl.Result{}, err
	}

	// Create the backup PVC
	var pvc v1.PersistentVolumeClaim

	pvc.Name = db.GetName() // Name is the same as this object
	pvc.Namespace = db.GetNamespace()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		if pvc.CreationTimestamp.IsZero() {
			log.Info("Creating backup pvc", "backupPvc", pvc.Name)
			pvc.ObjectMeta.Labels = createLabels(pvc.Name, nil)
			pvc.Annotations = map[string]string{
				"pv.beta.kubernetes.io/gid": "0",
			}
			pvc.Spec = *db.Spec.VolumeClaimSpec
			return controllerutil.SetControllerReference(&db, &pvc, r.Scheme)
		}
		return nil
	})

	// /// Create/update the backup target PVC that holds the backup
	// pvc, err := createPVC(ctx, r.Client, &db, db.Spec.BackupPVC.Size, db.Spec.BackupPVC.StorageClassName, "", r.Scheme)

	if err != nil {
		log.Error(err, "PVC claim creation failed", "pvcName", db.GetName())
		return ctrl.Result{}, err
	}

	//  Create a Snapshot of the target PVC to be backed up
	var snap snapshot.VolumeSnapshot
	snap.Name = "snap-" + db.GetName()
	snap.Namespace = db.GetNamespace()

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &snap, func() error {
		log.V(8).Info("CreateorUpdate snapshot", "name", snap.GetName())

		// does the snap not exist yet?
		if snap.CreationTimestamp.IsZero() {
			snap.ObjectMeta.Labels = createLabels(snap.GetName(), nil)
			snap.Spec = snapshot.VolumeSnapshotSpec{
				VolumeSnapshotClassName: &db.Spec.VolumeSnapshotClassName,
				Source:                  snapshot.VolumeSnapshotSource{PersistentVolumeClaimName: &db.Spec.ClaimToBackup}}

			return controllerutil.SetOwnerReference(&db, &snap, r.Scheme)
		} else {
			return controllerutil.SetOwnerReference(&db, &snap, r.Scheme)
		}
	})

	if err != nil {
		log.Error(err, "Snapshot creation failed", "claimToBackup", snap.Name)
		return ctrl.Result{}, err
	}

	// Now create a PVC with the contents of the snapshot. This PVC will be mounted by the backup job.
	// TODO: Backup Size should be calculated from the size target PVC
	// The datasource of the PVC is set to be the snapshot we just created above.

	var dataPVC v1.PersistentVolumeClaim

	dataPVC.Name = snap.GetName() // Name is the same as the VolumeSnapshot
	dataPVC.Namespace = snap.GetNamespace()

	// TODO: Fix these
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &dataPVC, func() error {
		log.V(8).Info("CreateorUpdate PVC from snapshot", "pvcName", snap.GetName())

		var apiGroup = SnapshotApiGroupString // so we can take address

		// does the snap not exist yet?
		if dataPVC.CreationTimestamp.IsZero() {
			dataPVC.ObjectMeta.Labels = createLabels(snap.GetName(), nil)
			dataPVC.Annotations = map[string]string{
				"pv.beta.kubernetes.io/gid": "0",
			}
			// TODO: Fix ME - this should be a copy of the original pvc
			// The size/class, etc. should come from the original pvc
			dataPVC.Spec = *db.Spec.VolumeClaimSpec
			dataPVC.Spec.DataSource = &v1.TypedLocalObjectReference{
				Kind:     "VolumeSnapshot",
				Name:     snap.GetName(),
				APIGroup: &apiGroup,
			}

			return controllerutil.SetOwnerReference(&db, &dataPVC, r.Scheme)
		} else {
			return controllerutil.SetOwnerReference(&db, &dataPVC, r.Scheme)
		}
	})

	if err != nil {
		log.Error(err, "PVC creation failed", "pvcName", snap.Name, "dataPVC", dataPVC)
		return ctrl.Result{}, err
	}
	// Create the Pod/Job that runs the LDIF export
	// args := []string{"/opt/opendj/scripts/ds-backup.sh"}
	args := []string{"backup"}

	job, err := createDSJob(ctx, r.Client, r.Scheme, &dataPVC, pvc.GetName(), &db.Spec.Keystore, args, db.Spec.Image, &db, db.Spec.ImagePullPolicy)

	if err != nil {
		log.Error(err, "Backup Job creation failed", "job", job)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *DirectoryBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// Every API object this controller can own needs to be setup here:
	return ctrl.NewControllerManagedBy(mgr).
		For(&directoryv1alpha1.DirectoryBackup{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&snapshot.VolumeSnapshot{}).
		Owns(&batch.Job{}).
		Complete(r)
}
