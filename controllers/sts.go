/*
	Copyright 2020 ForgeRock AS.
*/

package controllers

import (
	"context"
	"fmt"
	"strconv"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	k8slog "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	SnapshotApiGroup = "snapshot.storage.k8s.io"
)

func (r *DirectoryServiceReconciler) reconcileSTS(ctx context.Context, ds *directoryv1alpha1.DirectoryService, svcName string) error {
	log := k8slog.FromContext(ctx)
	var sts apps.StatefulSet
	sts.Name = ds.Name
	sts.Namespace = ds.Namespace

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &sts, func() error {
		log.V(8).Info("CreateorUpdate statefulset", "sts", sts)

		var err error
		// does the sts not exist yet?
		if sts.CreationTimestamp.IsZero() {
			// create the STS
			r.createDSStatefulSet(ctx, ds, &sts, svcName)
			_ = controllerutil.SetControllerReference(ds, &sts, r.Scheme)
			//
		} else {
			// If the sts exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			err = updateDSStatefulSet(ds, &sts)
		}

		log.V(8).Info("sts after update/create", "sts", sts)
		return err

	})
	if err != nil {
		return errors.Wrap(err, "unable to CreateOrUpdate StateFulSet")
	}
	return nil
}

// This function updates an existing statefulset to match settings in the custom resource
// StatefulSets allow only a limited number of changes
func updateDSStatefulSet(ds *directoryv1alpha1.DirectoryService, sts *apps.StatefulSet) error {

	// Copy our expected replicas to the statefulset
	sts.Spec.Replicas = ds.Spec.Replicas

	// copy the current sts replicas up the ds status
	ds.Status.CurrentReplicas = &sts.Status.CurrentReplicas

	// Update the image
	sts.Spec.Template.Spec.Containers[0].Image = ds.Spec.Image
	sts.Spec.Template.Spec.InitContainers[0].Image = ds.Spec.Image

	return nil
}

// https://godoc.org/k8s.io/api/apps/v1#StatefulSetSpec
func (r *DirectoryServiceReconciler) createDSStatefulSet(ctx context.Context, ds *directoryv1alpha1.DirectoryService, sts *apps.StatefulSet, svcName string) {
	// var initArgs []string // args provided to the init container
	var advertisedListenAddress = fmt.Sprintf("$(POD_NAME).%s", ds.Name)

	if ds.Spec.MultiCluster.ClusterTopology != "" {
		// Remove AdvertisedListenAddress default value so it can be configured by multi-cluster settings
		advertisedListenAddress = ""
	}

	var volumeMounts = []v1.VolumeMount{
		{
			Name:      "data",
			MountPath: DSDataPath,
		},

		{
			Name:      "admin-password",
			MountPath: "/var/run/secrets/admin",
		},
		{
			Name:      "monitor-password",
			MountPath: "/var/run/secrets/monitor",
		},
		{
			Name:      "ds-ssl-keypair",
			MountPath: SSLKeyPath,
		},
		{
			Name:      "ds-master-keypair",
			MountPath: MasterKeyPath,
		},
		{
			Name:      "truststore",
			MountPath: TruststoreKeyPath,
		},
		{
			Name:      "keys",
			MountPath: "/var/run/secrets/keys",
		},
	}

	var volumes = []v1.Volume{
		{
			Name: "ds-master-keypair", // master keypair for encryption
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: ds.Spec.Certificates.MasterSecretName,
				},
			},
		},
		{
			Name: "ds-ssl-keypair", // ssl between instances
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: ds.Spec.Certificates.SSLSecretName,
				},
			},
		},
		{
			Name: "truststore", // truststore
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: ds.Spec.Certificates.TruststoreSecretName,
				},
			},
		},
		{
			Name: "admin-password",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: ds.Spec.Passwords["uid=admin"].SecretName,
				},
			},
		},
		{
			Name: "monitor-password",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: ds.Spec.Passwords["uid=monitor"].SecretName,
				},
			},
		},
		{
			Name: "keys", // where DS expects to find the PEM keys
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}

	var mode int32 = 0755 // mode to mount scripts

	// If the user supplies a script configmap, mount it to /opt/opendj/scripts
	if ds.Spec.ScriptConfigMapName != "" {

		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      "scripts",
			MountPath: "/opt/opendj/scripts",
		})

		volumes = append(volumes, v1.Volume{
			Name: "scripts",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: ds.Spec.ScriptConfigMapName,
					},
					DefaultMode: &mode,
				},
			},
		})
	}

	var envVars = []v1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name:  "DS_ADVERTISED_LISTEN_ADDRESS",
			Value: advertisedListenAddress,
		},
		{
			Name:  "DS_GROUP_ID",
			Value: ds.Spec.GroupID,
		},
		{
			Name:  "DS_CLUSTER_TOPOLOGY",
			Value: ds.Spec.MultiCluster.ClusterTopology,
		},
		{
			Name:  "MCS_ENABLED",
			Value: strconv.FormatBool(ds.Spec.MultiCluster.McsEnabled),
		},
		{
			Name:  "DS_SET_UID_ADMIN_AND_MONITOR_PASSWORDS",
			Value: "true",
		},
		{
			Name:  "DS_UID_MONITOR_PASSWORD_FILE",
			Value: "/var/run/secrets/monitor/" + ds.Spec.Passwords["uid=monitor"].Key,
		},
		{
			Name:  "DS_UID_ADMIN_PASSWORD_FILE",
			Value: "/var/run/secrets/admin/" + ds.Spec.Passwords["uid=admin"].Key,
		},
	}

	// Create a template
	stemplate := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    createLabels(ds.Name, nil),
			Name:      ds.Name,
			Namespace: ds.Namespace,
		},
		Spec: apps.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":     LabelApplicationName,
					"app.kubernetes.io/instance": ds.Name,
				},
			},
			ServiceName: svcName,
			Replicas:    ds.Spec.Replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: createLabels(ds.Name, nil),
				},
				Spec: v1.PodSpec{
					// Spread the DS pods across zones if possible. If not possible, schedule anyways
					TopologySpreadConstraints: []v1.TopologySpreadConstraint{
						{
							MaxSkew:           1,
							TopologyKey:       "topology.kubernetes.io/zone",
							WhenUnsatisfiable: v1.ScheduleAnyway,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app.kubernetes.io/instance": ds.Name,
								},
							},
						},
					},
					// Tolerate any nodes tainted with kubectl taint nodes node1 key=directory:NoSchedule
					// This has no effect if the user does not wish to taint any nodes.
					Tolerations: []v1.Toleration{
						{
							Key:      "key",
							Operator: "Equal",
							Value:    "directory",
							Effect:   "NoSchedule",
						},
					},
					// Required for kubedns multi-cluster deployments
					Subdomain: svcName,
					InitContainers: []v1.Container{

						{
							Name:            "init",
							Image:           ds.Spec.Image,
							ImagePullPolicy: ds.Spec.ImagePullPolicy,
							Args:            []string{"init"},
							VolumeMounts:    volumeMounts,
							Resources:       ds.DeepCopy().Spec.Resources,
							Env:             envVars,
						},
					},
					Containers: []v1.Container{
						{
							Name:            "ds",
							Image:           ds.Spec.Image,
							ImagePullPolicy: ds.Spec.ImagePullPolicy,
							Args:            []string{"start-ds"},
							VolumeMounts:    volumeMounts,
							Resources:       ds.DeepCopy().Spec.Resources,
							Env:             envVars,
						},
					},
					SecurityContext: &v1.PodSecurityContext{
						FSGroup:   &RootGroup,
						RunAsUser: &ForgeRockUser,
					},
					Volumes: volumes,
				},
			},
			VolumeClaimTemplates: []v1.
				PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "data",
						Namespace: ds.Namespace,
						Labels:    createLabels(ds.Name, nil),
						Annotations: map[string]string{
							"pv.beta.kubernetes.io/gid": "0",
						},
					},
					Spec: r.setVolumeClaimTemplateFromSnapshot(ctx, ds),
				},
			},
		},
	}

	if DebugContainer {
		injectDebugContainers(stemplate, volumeMounts, ds.Spec.Image)
	}

	stemplate.DeepCopyInto(sts)
}

// If the user supplies a snapshot update the PVC volume claim to initialize from it
func (r *DirectoryServiceReconciler) setVolumeClaimTemplateFromSnapshot(ctx context.Context, ds *directoryv1alpha1.DirectoryService) v1.PersistentVolumeClaimSpec {
	spec := ds.Spec.VolumeClaimSpec.DeepCopy()

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

var rootUser int64 = 0 // todo: remove me

// Adds a debug init and sidecar containers.
func injectDebugContainers(sts *apps.StatefulSet, volumeMounts []v1.VolumeMount, image string) {

	var rootUser int64 = 0

	// add the debug init container. You can add an optional sleep here to inspect the pvc before the pods starts.
	// This is needed when the hostpath provisioner is used as it does not chown volumes to the pod user.
	var debugInit = []v1.Container{
		{
			Name:            "debug-init",
			Image:           image,
			ImagePullPolicy: v1.PullIfNotPresent,
			Command:         []string{"sh", "-c", "echo debug pod running && chown -R 11111:0 /opt/opendj/dsdata"},
			// Args: []string{"sleep 1000"},
			VolumeMounts: volumeMounts,
			// Currently the debug init runs as root so we can chmod the hostpath provisioner. This is only needed in testing.
			SecurityContext: &v1.SecurityContext{RunAsUser: &rootUser},
		},
	}

	// The debug sidecar has all the ds tools. It just sleeps waiting for the user to exec into the pod
	var debugSidecar = []v1.Container{
		{
			Name:            "debug",
			Image:           image,
			ImagePullPolicy: v1.PullIfNotPresent,
			Command:         []string{"bash", "-c", "echo debug pod running && while true; do sleep 300; done"},
			VolumeMounts:    volumeMounts,
		},
	}

	sts.Spec.Template.Spec.InitContainers = append(debugInit, sts.Spec.Template.Spec.InitContainers...)
	sts.Spec.Template.Spec.Containers = append(debugSidecar, sts.Spec.Template.Spec.Containers...)

}
