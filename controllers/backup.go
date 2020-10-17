package controllers

import (
	"context"
	"fmt"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	ldap "github.com/ForgeRock/ds-operator/pkg/ldap"
)

/// Update the backup schedule, and get backup status

func (r *DirectoryServiceReconciler) updateBackup(ctx context.Context, ds *directoryv1alpha1.DirectoryService, l *ldap.DSConnection) error {
	log := r.Log

	if !ds.Spec.Backup.Enabled {
		log.Info("Backup is disabled. Nothing to do")
		// todo: We could save a call to ds by checking our status to see if we need to disable ds..
		// but this is very robust...
		l.DeleteBackupSchedule(ds.Name)
		return nil
	}
	// todo: We also need to check to see if the backup has already been scheduled, in which case we dont need to do it again

	bp := ldap.BackupParams{ID: ds.Name, Cron: ds.Spec.Backup.Cron, Path: ds.Spec.Backup.Path}
	err := l.ScheduleBackup(&bp)
	if err != nil {
		log.Error(err, "backup schedule failed")
		return err
	}

	return nil
}

func (r *DirectoryServiceReconciler) updateBackupStatus(ctx context.Context, ds *directoryv1alpha1.DirectoryService, l *ldap.DSConnection) error {
	fmt.Printf("**** current stat %v", ds.Status.BackupStatus)
	ds.Status.BackupStatus = nil
	stat, err := l.GetBackupTaskStatus(ds.Name)
	// l.GetBackupTaskStatus(ds.Name)

	// take the first 10 ???
	ds.Status.BackupStatus = stat[:10]
	r.Log.Info("status", "stat", stat)
	if err != nil {
		return err
	}

	return nil
}
