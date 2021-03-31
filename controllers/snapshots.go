/*
	Copyright 2020 ForgeRock AS.
*/

package controllers

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	snapshot "github.com/kubernetes-csi/external-snapshotter/client/v3/apis/volumesnapshot/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Create snapshots if enabled..
// Compare status to last snapshot
func (r *DirectoryServiceReconciler) reconcileSnapshots(ctx context.Context, ds *directoryv1alpha1.DirectoryService) error {

	r.Log.Info("snapshot recon")

	if !ds.Spec.Snapshots.Enabled {
		return nil
	}

	now := time.Now().Unix()
	last := ds.Status.SnapshotStatus.LastSnapshotTimeStamp

	deadline := last + int64(ds.Spec.Snapshots.PeriodMinutes*60)

	r.Log.Info("snapshot deadine", "deadline", deadline, "lastSnapTimestamp", ds.Status.SnapshotStatus.LastSnapshotTimeStamp)
	// fmt.Printf("deadine %d  last %d", deadline, ds.Status.SnapshotStatus.LastSnapshotTimeStamp)

	if now < deadline {
		r.Log.V(8).Info("Snapshot deadline not passed yet.")
		return nil
	}

	pvcClaimToSnap := fmt.Sprintf("data-%s-0", ds.GetName())

	snapName := fmt.Sprintf("%s-%d", ds.GetName(), now)

	labels := createLabels(ds.GetName(), nil)

	// TODO: add labels, etc..
	var s = &snapshot.VolumeSnapshot{
		ObjectMeta: v1.ObjectMeta{Name: snapName, Namespace: ds.GetNamespace(),
			Labels:      labels,
			Annotations: map[string]string{"directory.forgerock.io/lastSnapshotTime": strconv.Itoa(int(now))},
		},
		Spec: snapshot.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &ds.Spec.Snapshots.VolumeSnapshotClassName,
			Source:                  snapshot.VolumeSnapshotSource{PersistentVolumeClaimName: &pvcClaimToSnap}}}

	fmt.Printf("snap is %+v\n", s)

	r.Log.Info("taking snapshot ", "snasphot", snapName, "pvc", pvcClaimToSnap)

	var snap snapshot.VolumeSnapshot
	snap.Name = s.GetName()
	snap.Namespace = s.GetNamespace()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &snap, func() error {
		r.Log.V(8).Info("CreateorUpdate snapshot", "name", snap.GetName())

		// does the sanp not exist yet?
		if snap.CreationTimestamp.IsZero() {
			s.DeepCopyInto(&snap)
			r.Log.V(8).Info("Setting ownerref for snapshot", "name", snap.Name)
			_ = controllerutil.SetControllerReference(ds, &snap, r.Scheme)
		} else {
			r.Log.V(8).Info("Snapshot should not already exist. Report this error", "snapshot", snap)
		}
		return nil
	})
	r.recorder.Event(ds, corev1.EventTypeNormal, "Created Snapshot", snap.Name)

	// update the status. Snapshots are expensive - and could pile up
	// Update the status sooner vs. later so we record the last snap time
	ds.Status.SnapshotStatus.LastSnapshotTimeStamp = now
	if err := r.Status().Update(ctx, ds); err != nil {
		r.Log.Error(err, "Could not update status")
		return err
	}

	snapList, err := r.getSnapshotList(ctx, ds)
	if err != nil {
		return err
	}

	numToDelete := len(snapList.Items) - int(ds.Spec.Snapshots.SnapshotsRetained)

	// If there are more snapshots than we are supposed to keep, delete the older ones
	if numToDelete > 0 {
		for i := 0; i < numToDelete; i++ {
			s := &snapList.Items[i]
			r.Log.Info("Pruning older snapshop", "snapshot", s.GetName())
			fmt.Printf("Snap to delete %+v\n", s)
			// Ignore any errors - attempt to complete all deletes
			if err := r.Client.Delete(ctx, s); err != nil {
				r.Log.Error(err, "Warning - could not delete snapshot", "snapshot", s.GetName())
			}
			r.recorder.Event(ds, corev1.EventTypeNormal, "Purged Snapshot", s.GetName())
		}
	}

	return nil
}

// Lookup the list of snapshots. The list will be returned in sorted order of time
func (r *DirectoryServiceReconciler) getSnapshotList(ctx context.Context, ds *directoryv1alpha1.DirectoryService) (*snapshot.VolumeSnapshotList, error) {
	// Now purge any older snapshots..
	// list snapshots
	var snapshotList snapshot.VolumeSnapshotList
	labels := createLabels(ds.GetName(), nil)

	// todo: Filter client.MatchingFields{jobOwnerKey: req.Name}
	err := r.Client.List(ctx, &snapshotList, client.InNamespace(ds.GetNamespace()), client.MatchingLabels(labels))
	if err != nil {
		r.Log.Error(err, "Could not list snapshots")
		return nil, err
	}

	// Sort the list
	items := snapshotList.Items[:]
	sort.Slice(items, func(i, j int) bool {
		x := items[i].CreationTimestamp
		y := items[j].CreationTimestamp
		return x.Unix() < y.Unix()
	})
	return &snapshotList, nil
}
