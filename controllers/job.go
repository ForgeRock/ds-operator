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

// Create a directory service job that can backup or restore data
func createDSJob(ctx context.Context, client client.Client, scheme *runtime.Scheme, dataPVC *v1.PersistentVolumeClaim, backupPVC string,
	keystore *directoryv1alpha1.DirectoryKeystores, command, args []string, image string, owner metav1.Object, pullPolicy v1.PullPolicy) (*batch.Job, error) {

	var job batch.Job
	log := k8slog.FromContext(ctx)

	//  Use in the security context only when testing using the hostpath provisioner
	// The hostpath creates volumes owned by root - which the forgerock user can not access.
	var rootUserOnlyForTesting int64 = 0

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
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: dataPVC.GetName()},
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
							// RunAsUser:      &ForgeRockUser,
							// ****** FOR TESTING ONLY ****** TODO: Remove.
							// TODO: Add a runtime flag to the controller to enable this "feature"
							RunAsUser:  &rootUserOnlyForTesting,
							RunAsGroup: &RootGroup,
							FSGroup:    &RootGroup,
						},
						Containers: []v1.Container{
							{
								Name:  "ds-job",
								Image: image,
								// to debug use this, and comment out Args
								// Command: []string{"/bin/sh", "-c", "sleep 3000"},
								Command:         command,
								Args:            args,
								ImagePullPolicy: pullPolicy,
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
			// Set the data pvc to be owned by the Job
			_ = controllerutil.SetOwnerReference(&job, dataPVC, scheme)
			// Set the Job to be owned by the CR
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
