/*
	Copyright 2021 ForgeRock AS.
*/

/// Manage Volume Snapshot Creation
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
	"github.com/prometheus/common/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Create snapshots if enabled..
// The spec.status records the last snapshot time.
func (r *DirectoryServiceReconciler) reconcileSnapshots(ctx context.Context, ds *directoryv1alpha1.DirectoryService) error {

	r.Log.V(5).Info("snapshot reconcile")

	if !ds.Spec.Snapshots.Enabled {
		return nil
	}
	now := time.Now().Unix()

	// If the timestamp is 0, we have never taken a snapshot before
	// We want to skip the first snapshot time - because DS is in the process of initializing the disk
	// We dont want to snapshot an empty disk
	if ds.Status.SnapshotStatus.LastSnapshotTimeStamp == 0 {
		ds.Status.SnapshotStatus.LastSnapshotTimeStamp = now
		if err := r.Status().Update(ctx, ds); err != nil {
			r.Log.Error(err, "Could not update status")
			return err
		}
		r.Log.Info("Skipping first snapshot to allow the directory to come up")
		return nil
	}

	last := ds.Status.SnapshotStatus.LastSnapshotTimeStamp
	deadline := last + int64(ds.Spec.Snapshots.PeriodMinutes*60)

	r.Log.V(5).Info("snapshot deadine", "deadline", deadline, "lastSnapTimestamp", ds.Status.SnapshotStatus.LastSnapshotTimeStamp)
	// fmt.Printf("deadine %d  last %d", deadline, ds.Status.SnapshotStatus.LastSnapshotTimeStamp)

	if now < deadline {
		r.Log.V(5).Info("Snapshot deadline not passed yet.")
		return nil
	}

	// We always snapshot the first disk
	pvcClaimToSnap := fmt.Sprintf("data-%s-0", ds.GetName())
	// snaphsot name suffix is the current timestamp
	snapName := fmt.Sprintf("%s-%d", ds.GetName(), now)

	// TODO: add labels, etc..
	var s = &snapshot.VolumeSnapshot{
		ObjectMeta: v1.ObjectMeta{Name: snapName, Namespace: ds.GetNamespace(),
			Labels:      createLabels(ds.GetName(), nil),
			Annotations: map[string]string{"directory.forgerock.io/lastSnapshotTime": strconv.Itoa(int(now))},
		},
		Spec: snapshot.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &ds.Spec.Snapshots.VolumeSnapshotClassName,
			Source:                  snapshot.VolumeSnapshotSource{PersistentVolumeClaimName: &pvcClaimToSnap}}}

	r.Log.Info("taking snapshot ", "snasphot", snapName, "pvc", pvcClaimToSnap)

	var snap snapshot.VolumeSnapshot
	snap.Name = s.GetName()
	snap.Namespace = s.GetNamespace()

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &snap, func() error {
		r.Log.V(8).Info("CreateorUpdate snapshot", "name", snap.GetName())

		// does the snap not exist yet?
		if snap.CreationTimestamp.IsZero() {
			s.DeepCopyInto(&snap)
			// Note: We dont set the ownerref - we want to snapshot to persist even if the
			// directory instance is deleted
			// r.Log.V(8).Info("Setting ownerref for snapshot", "name", snap.Name)
			// _ = controllerutil.SetControllerReference(ds, &snap, r.Scheme)
		} else {
			r.Log.V(8).Info("Snapshot should not already exist. Report this error", "snapshot", snap)
		}
		return nil
	})

	if err != nil {
		log.Error(err, "Warning, Create/Update of VolumeSnapshot failed. Will continue processing")
	}

	r.recorder.Event(ds, corev1.EventTypeNormal, "Created Snapshot", snap.Name)

	// Update the status ASAP. Snapshots are expensive, so
	// we record the last snap time so we dont accidently try to create a bunch of snapshots in rapid succession.
	ds.Status.SnapshotStatus.LastSnapshotTimeStamp = now
	if err := r.Status().Update(ctx, ds); err != nil {
		r.Log.Error(err, "Could not update status")
		return err
	}

	snapList, err := r.getSnapshotList(ctx, ds)
	if err != nil {
		return err
	}

	// Delete older snapshots
	numToDelete := len(snapList.Items) - int(ds.Spec.Snapshots.SnapshotsRetained)

	if numToDelete > 0 {
		for i := 0; i < numToDelete; i++ {
			s := &snapList.Items[i]
			r.Log.Info("Pruning older snapshot", "snapshot", s.GetName())
			// Ignore errors - attempt to complete all deletes
			if err := r.Client.Delete(ctx, s); err != nil {
				r.Log.Error(err, "Warning - could not delete snapshot", "snapshot", s.GetName())
			}
			r.recorder.Event(ds, corev1.EventTypeNormal, "Purged Snapshot", s.GetName())
		}
	}

	return nil
}

// Lookup the list of snapshots. The list is sorted by snapshot time
func (r *DirectoryServiceReconciler) getSnapshotList(ctx context.Context, ds *directoryv1alpha1.DirectoryService) (*snapshot.VolumeSnapshotList, error) {
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
