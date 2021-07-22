package controllers

import (
	"context"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	k8slog "sigs.k8s.io/controller-runtime/pkg/log"
)

// User and Ground ID
// var frUser int64 = 11111
// temporary hack to allow writing to /backup.
// The FR user fails because the hostpath provisioner does not set the correct permissions.
var frUser int64 = 0
var frRootGroup int64 = 0

// Create a directory service job that can backup or restore data
func createDSJob(ctx context.Context, client client.Client, scheme *runtime.Scheme, dataPVC, backupPVC string,
	keystore *directoryv1alpha1.DirectoryKeystores, args []string, image string, owner metav1.Object) (*batch.Job, error) {

	var job batch.Job
	log := k8slog.FromContext(ctx)

	job.Name = owner.GetName()
	job.Namespace = owner.GetNamespace()

	_, err := ctrl.CreateOrUpdate(ctx, client, &job, func() error {
		var err error
		if job.CreationTimestamp.IsZero() {
			log.V(8).Info("Creating job", "jobName", job.GetName())

			job.ObjectMeta.Labels = createLabels(job.GetName(), nil)
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
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: backupPVC},
								},
								Name: "backup",
							},
							{
								VolumeSource: v1.VolumeSource{
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: dataPVC},
								},
								Name: "data",
							},
							{
								VolumeSource: v1.VolumeSource{
									Secret: &v1.SecretVolumeSource{
										SecretName: keystore.SecretName,
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
								Image: image,
								// to debug use this, and comment out Args
								// Command: []string{"/bin/sh", "-c", "sleep 3000"},
								Args: args,
								// Sample command that is executed in the container:
								//  bin/import-ldif --ldifFile /var/tmp/test.ldif --backendId idmRepo --offline

								Env: []v1.EnvVar{
									{Name: "NAMESPACE", Value: owner.GetNamespace()},
									{Name: "BACKUP_TYPE", Value: "ldif"}, // this all we support right now
								},
								VolumeMounts: []v1.VolumeMount{
									{
										Name:      "backup",
										MountPath: "/backup",
									},
									{
										Name:      "data",
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
			err = controllerutil.SetOwnerReference(owner, &job, scheme)
		} else {
			// update the job.
			// nothing to do here....
			log.V(8).Info("update job, nothing to do..")
		}
		return err
	})

	return &job, err
}

// Create the backup Job
func (r *DirectoryBackupReconciler) createBackupJob(db *directoryv1alpha1.DirectoryBackup, ctx context.Context) (*batch.Job, error) {
	log := k8slog.FromContext(ctx)

	var job batch.Job
	job.Name = db.Name
	job.Namespace = db.Namespace

	backupPvcName := db.GetName() + "-pvc"
	dataPvc := "snap-" + db.GetName()

	// // var frUser int64 = 11111
	// // temporary hack to allow writing to /backup.
	// // The FR user fails because the hostpath provisioner does not set the correct permissions.
	// var frUser int64 = 0

	// var frRootGroup int64 =

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
								// to debug use this, and comment out Args
								// Command: []string{"/bin/sh", "-c", "sleep 3000"},
								Args: []string{"/opt/opendj/scripts/ds-backup.sh"},
								// Sample command that is executed in the container:
								//  bin/export-ldif --ldifFile /var/tmp/test.ldif --backendId idmRepo --offline

								Env: []v1.EnvVar{
									{Name: "NAMESPACE", Value: db.Namespace},
									{Name: "BACKUP_TYPE", Value: "ldif"}, // this all we support right now
								},
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
