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
	"time"

	kbatch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	klog "sigs.k8s.io/controller-runtime/pkg/log"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	snapshot "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
)

// DirectoryRestoreReconciler reconciles a DirectoryRestore object
type DirectoryRestoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Note the rbac rules are consolidated on the directoryservice_controller.go. These are for reference only

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

	// List in progress backup job to update status.
	var restoreJob kbatch.Job

	// Ignore if job is not found - it might not yet be created. Note the job  name is the same as the request object.
	if err := r.Get(ctx, req.NamespacedName, &restoreJob); client.IgnoreNotFound(err) != nil {
		log.Error(err, "cant fetch job status")
		return ctrl.Result{}, err
	}

	requeueWhileJobRunning := true

	// if the job has started, or possibly finished, update our status
	if !restoreJob.ObjectMeta.CreationTimestamp.IsZero() {
		// update CRD status
		ds.Status.StartTimestamp = &restoreJob.ObjectMeta.CreationTimestamp

		// Has the Restore Job completed?
		if restoreJob.Status.CompletionTime != nil {

			requeueWhileJobRunning = false

			log.Info("Job completed at", "completionTime", restoreJob.Status.CompletionTime)
			ds.Status.CompletionTimestamp = restoreJob.Status.CompletionTime

			// Did it fail?
			if restoreJob.Status.Failed > 0 {
				err := fmt.Errorf("Restore Job had failures. Cant take snapshot")
				log.Error(err, "restore job failed")
				return ctrl.Result{}, err
			}

			// If the restore job succeeded OK, we can now take the volume snapshot of the restore data volume
			if restoreJob.Status.Succeeded > 0 {
				var snap snapshot.VolumeSnapshot
				snap.Name = ds.GetName()
				snap.Namespace = ds.GetNamespace()
				ctrl.CreateOrUpdate(ctx, r.Client, &snap, func() error {
					log.Info("CreateorUpdate snapshot", "name", snap.GetName())
					// does the snap not exist yet?
					if snap.CreationTimestamp.IsZero() {
						snap.ObjectMeta.Labels = createLabels(ds.GetName(), nil)
						snap.Spec = snapshot.VolumeSnapshotSpec{
							VolumeSnapshotClassName: &ds.Spec.VolumeSnapshotClassName,
							Source:                  snapshot.VolumeSnapshotSource{PersistentVolumeClaimName: &ds.Name},
						}
						return controllerutil.SetOwnerReference(&ds, &snap, r.Scheme)

					} else {
						log.V(8).Info("Snapshot exits")
					}
					return nil
				})
			}
		} else {
			// Job is not finished yet - come back later
			log.V(8).Info("Restore Job status", "jobStatus", restoreJob.Status.Conditions)
		}
	}

	// update status
	if err := r.Status().Update(ctx, &ds); err != nil {
		log.Error(err, "unable to update DirectoryRestore status")
		return ctrl.Result{}, err
	}

	// Create/update the PVC to hold the restored data

	var pvc v1.PersistentVolumeClaim
	pvc.Name = ds.GetName()
	pvc.Namespace = ds.GetNamespace()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		if pvc.CreationTimestamp.IsZero() {
			log.Info("Creating restore pvc", "restorePvc", pvc.Name)
			pvc.ObjectMeta.Labels = createLabels(pvc.Name, nil)
			pvc.Annotations = map[string]string{
				"pv.beta.kubernetes.io/gid": "0",
			}
			pvc.Spec = *ds.Spec.VolumeClaimSpec
			return controllerutil.SetControllerReference(&ds, &pvc, r.Scheme)
		}
		return nil
	})

	if err != nil {
		log.Error(err, "PVC claim creation failed", "pvcName", ds.Name)
		return ctrl.Result{}, err
	}

	// Note we override the docker-entrypoint.sh here for restore - invoking ds-restore.sh directly
	command := []string{"/opt/opendj/scripts/ds-restore.sh"}
	args := []string{}
	// Create the restore Job
	job, err := createDSJob(ctx, r.Client, r.Scheme, &pvc, ds.Spec.SourcePVCName, &ds.Spec.Keystore, command, args, ds.Spec.Image, &ds, ds.Spec.ImagePullPolicy)

	if err != nil {
		log.Error(err, "Job create failed", "jobName", job.Name)
		return ctrl.Result{}, err
	}

	// Job is still not done - come back later
	if requeueWhileJobRunning {
		return ctrl.Result{RequeueAfter: time.Second * 60}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DirectoryRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&directoryv1alpha1.DirectoryRestore{}).
		Complete(r)
}
