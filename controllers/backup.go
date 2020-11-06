package controllers

import (
	"context"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	ldap "github.com/ForgeRock/ds-operator/pkg/ldap"
)

/// Update the backup and purge schedules, and get backup status

func (r *DirectoryServiceReconciler) updateBackup(ctx context.Context, ds *directoryv1alpha1.DirectoryService, l *ldap.DSConnection) error {
	log := r.Log

	if !ds.Spec.Backup.Enabled {
		log.Info("Backup is disabled. Nothing to do")
		// Delete the backup tasks - even if they dont exist - this is no more expensive than querying then deleting..
		l.DeleteBackupTask(ds.Name)
		return nil
	}

	// else - we need to read the current task status, and see if we need to make changes
	var b *directoryv1alpha1.DirectoryBackup
	var err error

	b, err = l.GetBackupTask(ds.Name)

	if err != nil || backupTasksAreDifferent(&ds.Spec.Backup, b) {
		log.Info("Scheduling backup task", "task", ds.Spec.Backup)
		if err := l.ScheduleBackup(ds.Name, &ds.Spec.Backup); err != nil {
			log.Error(err, "Unable to create backup task")
			return err
		}
	}

	return nil
}

// Return true if the actual backup task in ds is not the same as the desired task state
func backupTasksAreDifferent(desired *directoryv1alpha1.DirectoryBackup, actual *directoryv1alpha1.DirectoryBackup) bool {
	// note the actual values comes from the directory itself - and it does not have the secretName
	// so we copy that, and then do a struct compare
	actual.SecretName = desired.SecretName
	return *actual != *desired
}

func (r *DirectoryServiceReconciler) updateBackupStatus(ctx context.Context, ds *directoryv1alpha1.DirectoryService, l *ldap.DSConnection) error {
	r.Log.V(5).Info("Current backup status", "status", ds.Status.BackupStatus)
	stat, err := l.GetBackupTaskStatus(ds.Name)

	if err != nil {
		r.Log.V(5).Info("Can't get backup status. This is OK if no backups have been scheduled", "err", err)
		return err
	}
	r.Log.V(5).Info("Backup status", "stat", stat)
	ds.Status.BackupStatus = stat

	return nil
}
