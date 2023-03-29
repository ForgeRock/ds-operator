/*
Copyright 2022, ForgeRock AS.
*/
package controllers

import (
	"context"
	"fmt"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	k8slog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconcile peristent volume claims. We pre-create the PVCs here before the template in the statefulset takes effect.
// This is in case we want to scale up a new DS node
// and we want the pvc to be created from the most recent snapshot, or if the user wants to edit the DS CR and change the volume source.
//
//	The StatefulSet volume claim template is immutable. By pre-creating the pvc, If Kubernetes sees the PVC exists already, it will not create a new one.
func (r *DirectoryServiceReconciler) reconcilePVC(ctx context.Context, ds *directoryv1alpha1.DirectoryService, svcName string) error {
	log := k8slog.FromContext(ctx)

	// For each number replicas, make sure there is a corresponding PVC
	for i := 0; i < int(*ds.Spec.Replicas); i++ {
		var pvc v1.PersistentVolumeClaim
		// Name here must match the sts template name
		pvc.Name = fmt.Sprintf("data-%s-%d", ds.GetObjectMeta().GetName(), i)
		pvc.Namespace = ds.GetObjectMeta().GetNamespace()

		_, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
			log.V(8).Info("CreateorUpdate PVC", "pvc", pvc)

			var err error
			// does the pvc not exist yet?
			if pvc.CreationTimestamp.IsZero() {
				pvc.ObjectMeta.Labels = createLabels(ds.GetObjectMeta().GetName(), ds.Kind, nil)
				// enables the volume to be writen by the forgerock user in the root group.
				pvc.ObjectMeta.Annotations = map[string]string{
					"pv.beta.kubernetes.io/gid": "0",
				}

				pvc.Spec = r.setVolumeClaimTemplateFromSnapshot(ctx, ds)

				// Note we choose not to set an ownerReference so the PVC is not deleted when DS is deleted

			} else {
				// If the PVC exists already - we want to update any fields to bring its state into
				// alignment with the Custom Resource. However - PVC once created are immutable.
				// Not clear we can do much here..
				log.V(8).Info("TODO: Handle update of PVC service")
			}

			log.V(8).Info("pvc after update/create", "pvc", pvc)
			return err
		})

		if err != nil {
			log.Error(err, "Failed to create or update PVC", "pvc", pvc.Name)
			return err
		}

	}
	return nil

}

// If the user supplies a snapshot update the PVC volume claim to initialize from it
func (r *DirectoryServiceReconciler) setVolumeClaimTemplateFromSnapshot(ctx context.Context, ds *directoryv1alpha1.DirectoryService) v1.PersistentVolumeClaimSpec {
	spec := ds.Spec.PodTemplate.VolumeClaimSpec.DeepCopy()

	// If the user wants to init from a snapshot, and they use the sentinel value "$(latest)" - Then try to calculate the latest snapshot name
	if spec.DataSource != nil && spec.DataSource.Name == "$(latest)" {
		snapList, err := r.getSnapshotList(ctx, ds)
		if err != nil || len(snapList.Items) == 0 {
			// nill the datasource
			spec.DataSource = nil
		} else {
			// The snapList is sorted - the last entry is the most recent
			// Set the datasource to the latest snapshot name
			spec.DataSource.Name = snapList.Items[len(snapList.Items)-1].GetName()
		}
	}
	return *spec
}
