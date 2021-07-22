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
	"fmt"

	kbatch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
)

// DirectoryRestoreReconciler reconciles a DirectoryRestore object
type DirectoryRestoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=directory.forgerock.io,resources=directoryrestores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=directory.forgerock.io,resources=directoryrestores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=directory.forgerock.io,resources=directoryrestores/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DirectoryRestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var log = klog.FromContext(ctx)

	log.Info("Reconciling DirectoryRestore")

	// fetch the DirectoryRestore object
	var ds directoryv1alpha1.DirectoryRestore

	// Load the DirectoryRestore object
	if err := r.Get(ctx, req.NamespacedName, &ds); err != nil {
		log.Info("Unable to fetch DirectoryRestore - it is in the process of being deleted. This is OK")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List in progress backup job to update status. The backup job has the same name as this object
	var restoreJob kbatch.Job

	// Ignore if job is not found - it might not yet be created.
	if err := r.Get(ctx, req.NamespacedName, &restoreJob); client.IgnoreNotFound(err) != nil {
		log.Error(err, "cant fetch job status")
		return ctrl.Result{}, err
	}

	if !restoreJob.ObjectMeta.CreationTimestamp.IsZero() {
		// update CRD status
		ds.Status.StartTimestamp = &restoreJob.ObjectMeta.CreationTimestamp
		if restoreJob.Status.CompletionTime != nil {
			ds.Status.CompletionTimestamp = restoreJob.Status.CompletionTime
		}
	}

	// update status
	if err := r.Status().Update(ctx, &ds); err != nil {
		log.Error(err, "unable to update DirectoryRestore status")
		return ctrl.Result{}, err
	}

	/// Create/update the PVC to hold the restored data
	pvc, err := createPVC(ctx, r.Client, ds.Spec.RestorePVC.Name, ds.GetNamespace(), ds.Spec.RestorePVC.Size, ds.Spec.RestorePVC.StorageClassName, "")

	// keep linter happy
	fmt.Println("PVC: ", pvc)

	if err != nil {
		log.Error(err, "PVC claim creation failed", "pvcName", ds.Spec.RestorePVC.Name)
		return ctrl.Result{}, err
	}

	// var job kbatch.Job
	// job.Name = ds.GetName()
	// job.Namespace = ds.GetNamespace()
	args := []string{"foo"}

	// Create the restore Job
	job, err := createDSJob(ctx, r.Client, r.Scheme, ds.Spec.RestorePVC.Name, ds.Spec.SourcePVCName, &ds.Spec.Keystore, args, ds.Spec.Image, &ds)

	if err != nil {
		log.Error(err, "Job create failed", "jobName", job.Name)
		return ctrl.Result{}, err
	}
	// Make this job owned by this object so it is garbage collected
	// _ = controllerutil.SetOwnerReference(&ds, &job, r.Scheme)

	// Create the Restore Job to restore the data.

	// //  Create a Snapshot of the target PVC to be backed up
	// var snap snapshot.VolumeSnapshot
	// snap.Name = "snap-" + ds.GetName()
	// snap.Namespace = ds.GetNamespace()

	// _, err = ctrl.CreateOrUpdate(ctx, r.Client, &snap, func() error {
	// 	log.V(8).Info("CreateorUpdate snapshot", "name", snap.GetName())

	// 	// does the snap not exist yet?
	// 	if snap.CreationTimestamp.IsZero() {
	// 		snap.ObjectMeta.Labels = createLabels(snap.GetName(), nil)
	// 		snap.Spec = snapshot.VolumeSnapshotSpec{
	// 			VolumeSnapshotClassName: &ds.Spec.VolumeSnapshotClassName,
	// 			Source:                  snapshot.VolumeSnapshotSource{PersistentVolumeClaimName: &ds.Spec.ClaimToBackup}}

	// 		return controllerutil.SetOwnerReference(&ds, &snap, r.Scheme)
	// 	} else {
	// 		log.Info("Snapshot should not already exist. Report this error", "snapshot", snap)
	// 	}

	// 	return nil
	// })

	// if err != nil {
	// 	log.Error(err, "Snapshot creation failed", "claimToBackup", snap.Name)
	// 	return ctrl.Result{}, err
	// }

	// // Now create a PVC with the contents of the snapshot. This PVC will be mounted by the backup job.
	// // The Data pvc is named the same as the VolumeSnaphshot.
	// // TODO: Backup Size should be calulated from the size target PVC
	// // The datasource of the PVC is set to be the snapshot we just created above.
	// dataPVC, err := createPVC(ctx, r.Client, snap.Name, ds.GetNamespace(), ds.Spec.BackupPVC.Size, ds.Spec.BackupPVC.StorageClassName, snap.Name)

	// if err != nil {
	// 	log.Error(err, "PVC creation failed", "pvcName", snap.Name, "dataPVC", dataPVC)
	// 	return ctrl.Result{}, err
	// }
	// // Create the Pod/Job that runs the LDIF export
	// job, err := r.createBackupJob(&ds, ctx)

	// if err != nil {
	// 	log.Error(err, "Backup Job creation failed", "job", job)
	// 	return ctrl.Result{}, err
	// }

	// // set the data pvc to be owned by the Job
	// err = controllerutil.SetOwnerReference(&ds, &dataPVC, r.Scheme)

	// log.Info("Created job", "job", job.GetName())

	// log.Info("Done")

	// return ctrl.Result{}, err

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DirectoryRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&directoryv1alpha1.DirectoryRestore{}).
		Complete(r)
}
