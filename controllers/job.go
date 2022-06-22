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
	podTemplate *directoryv1alpha1.DirectoryPodTemplate, args []string, owner metav1.Object, kind string) (*batch.Job, error) {

	var job batch.Job
	log := k8slog.FromContext(ctx)

	job.Name = owner.GetName()
	job.Namespace = owner.GetNamespace()

	user := ForgeRockUser
	if DebugContainer {
		// The hostpath creates volumes owned by root - which the forgerock user can not access.
		// This for development of the operator minikube only.
		log.V(8).Info("Debug container being configured, running as root.")
		user = 0
	}

	var envVars = []v1.EnvVar{
		{Name: "NAMESPACE", Value: owner.GetNamespace()},
	}

	if podTemplate.Env != nil {
		envVars = append(envVars, podTemplate.Env...)
	}

	var volumes = []v1.Volume{
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
					SecretName: podTemplate.Certificates.MasterSecretName,
				},
			},
			Name: "master-keypair", // pem based master key pair for crypting data
		},
		{
			Name: "keys", // where DS expects to find the PEM keys
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}

	var volumeMounts = []v1.VolumeMount{
		{
			Name:      "backup",
			MountPath: "/backup",
		},
		{
			Name:      "data",
			MountPath: DSDataPath,
		},
		{
			Name:      "master-keypair",
			MountPath: MasterKeyPath,
		},
		{
			Name:      "keys",
			MountPath: "/var/run/secrets/keys",
		},
	}

	var mode int32 = 0755 // mode to mount scripts

	// If the user supplies a script configmap, mount it to /opt/opendj/scripts
	if podTemplate.ScriptConfigMapName != "" {

		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      "scripts",
			MountPath: "/opt/opendj/scripts",
		})

		volumes = append(volumes, v1.Volume{
			Name: "scripts",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: podTemplate.ScriptConfigMapName,
					},
					DefaultMode: &mode,
				},
			},
		})
	}

	_, err := ctrl.CreateOrUpdate(ctx, client, &job, func() error {
		var err error
		//var controllerName string
		if job.CreationTimestamp.IsZero() {
			log.V(8).Info("Creating job", "jobName", job.GetName())

			job.ObjectMeta.Labels = createLabels(job.GetName(), kind, nil)
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
						Volumes:       volumes,
						RestartPolicy: v1.RestartPolicyNever,
						SecurityContext: &v1.PodSecurityContext{
							SELinuxOptions: &v1.SELinuxOptions{},
							WindowsOptions: &v1.WindowsSecurityContextOptions{},
							RunAsUser:      &user,
							RunAsGroup:     &RootGroup,
							FSGroup:        &RootGroup,
						},
						ServiceAccountName: podTemplate.ServiceAccountName,
						Containers: []v1.Container{
							{
								Name:            "ds-job",
								Image:           podTemplate.Image,
								Args:            args,
								ImagePullPolicy: podTemplate.ImagePullPolicy,
								Env:             envVars,
								Resources:       podTemplate.Resources,
								VolumeMounts:    volumeMounts,
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
