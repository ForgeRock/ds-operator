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
)

const (
	SnapshotApiGroupString = "snapshot.storage.k8s.io"
)

func createPVC(ctx context.Context, client client.Client, owner metav1.Object, size string, storageClassName string, snapshot string, scheme *runtime.Scheme) (v1.PersistentVolumeClaim, error) {
	var pvc v1.PersistentVolumeClaim

	setOwner := false
	// If a snapshot is provided, we use a snap prefix to name the pvc
	if snapshot != "" {
		pvc.Name = "snap-" + owner.GetName()
		setOwner = true
	} else {
		pvc.Name = owner.GetName()
	}

	pvc.Namespace = owner.GetNamespace()

	_, err := ctrl.CreateOrUpdate(ctx, client, &pvc, func() error {
		// does the sts not exist yet?
		if pvc.CreationTimestamp.IsZero() {
			var x *v1.PersistentVolumeClaim = &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvc.Name,
					Namespace: pvc.Namespace,
					Labels:    createLabels(pvc.Name, nil),
					Annotations: map[string]string{
						"pv.beta.kubernetes.io/gid": "0",
					},
				},
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceName(v1.ResourceStorage): resource.MustParse(size),
						},
					},
					StorageClassName: &storageClassName,
				},
			}

			apiGroup := SnapshotApiGroupString // assign so we can take address

			if snapshot != "" {
				x.Spec.DataSource = &v1.TypedLocalObjectReference{
					Kind:     "VolumeSnapshot",
					Name:     snapshot,
					APIGroup: &apiGroup,
				}
			}
			x.DeepCopyInto(&pvc)

			// Set the reference outside of this function
			if setOwner {
				_ = controllerutil.SetOwnerReference(owner, &pvc, scheme)
			}
			//
		} else {
			// If the sts exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			// nothing to do here - we cant' update pvcs
			if setOwner {
				_ = controllerutil.SetOwnerReference(owner, &pvc, scheme)
			}
		}
		return nil
	})

	return pvc, err
}
