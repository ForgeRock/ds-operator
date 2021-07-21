package controllers

import (
	"context"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	k8slog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Create the backup Job
func (r *DirectoryBackupReconciler) createBackupJob(db *directoryv1alpha1.DirectoryBackup, ctx context.Context) (*batch.Job, error) {
	log := k8slog.FromContext(ctx)

	var job batch.Job
	job.Name = db.Name
	job.Namespace = db.Namespace

	backupPvcName := db.GetName() + "-pvc"
	dataPvc := "snap-" + db.GetName()

	// var frUser int64 = 11111
	// temporary hack to allow writing to /backup.
	// The FR user fails because the hostpath provisioner does not set the correct permissions.
	var frUser int64 = 0

	var frRootGroup int64 = 0

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &job, func() error {

		var err error

		log.V(8).Info("Creating job", "jobName", db.GetName())
		if job.CreationTimestamp.IsZero() {
			job.ObjectMeta.Labels = createLabels(db.GetName(), nil)
			job.Spec = batch.JobSpec{
				// Parallelism:             new(int32),
				// Completions:             new(int32),
				// ActiveDeadlineSeconds:   new(int64),
				// BackoffLimit:            new(int32),
				// Selector:                &v1.LabelSelector{},
				// ManualSelector:          new(bool),
				// TTLSecondsAfterFinished: new(int32),
				// CompletionMode:          &"",
				// Suspend:                 new(bool),
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Volumes: []v1.Volume{
							{
								VolumeSource: v1.VolumeSource{
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: backupPvcName},
								},
								Name: backupPvcName,
							},
							{
								VolumeSource: v1.VolumeSource{
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: dataPvc},
								},
								Name: dataPvc,
							},
							{
								VolumeSource: v1.VolumeSource{
									Secret: &v1.SecretVolumeSource{
										SecretName: db.Spec.Keystore.SecretName,
									},
								},
								Name: "secrets", // keystore and pin
							},
						},
						RestartPolicy: v1.RestartPolicyNever,
						SecurityContext: &v1.PodSecurityContext{
							SELinuxOptions: &v1.SELinuxOptions{},
							WindowsOptions: &v1.WindowsSecurityContextOptions{},
							RunAsUser:      &frUser,
							RunAsGroup:     &frRootGroup,
							FSGroup:        &frRootGroup,
						},
						Containers: []v1.Container{
							{
								Name:  "ds-backup",
								Image: db.Spec.Image,
								// Command: []string{"/bin/sh", "-c", "sleep 3000"},
								Args: []string{"/opt/opendj/scripts/ds-backup.sh"},
								// Sample command that worked:
								//  bin/export-ldif --ldifFile /var/tmp/test.ldif --backendId idmRepo --offline
								VolumeMounts: []v1.VolumeMount{
									{
										Name:      backupPvcName,
										MountPath: "/backup",
									},
									{
										Name:      dataPvc,
										MountPath: "/opt/opendj/data",
									},
									{
										Name:      "secrets",
										MountPath: "/opt/opendj/pem-keys-directory/master-key",
										SubPath:   "master-key-pair-combined.pem",
									},
								},
							},
						},
					},
				},
			}
			err = controllerutil.SetControllerReference(db, &job, r.Scheme)

		} else {
			// update the job.
			// nothing to do here....
			log.V(8).Info("update job, nothing to do..")
		}

		return err
	})

	if err != nil {
		log.Error(err, "failed to create backup job")
		return nil, err
	}
	// todo: make the snap pvc owned by the job.. since that is what uses it

	return &job, err
}
