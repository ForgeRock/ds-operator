/*
	Copyright 2021 ForgeRock AS.
*/

package controllers

import (
	"context"
	"fmt"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8slog "sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

/// NOTE: The proxy is not supported. This is stub code in preparation for future support.

func (r *DirectoryServiceReconciler) reconcileProxy(ctx context.Context, ds *directoryv1alpha1.DirectoryService) error {
	var deployment apps.Deployment
	proxyName := ds.Name + "-proxy"
	var log = k8slog.FromContext(ctx)

	if !ds.Spec.Proxy.Enabled {
		err := r.Client.Get(ctx, types.NamespacedName{Name: proxyName, Namespace: ds.Namespace}, &deployment)
		if k8sErrors.IsNotFound(err) {
			return nil
		}
		// if the object exists, check if we own it
		if err == nil {
			owner := metav1.GetControllerOf(&deployment)
			// if we own the object, delete it
			if owner.APIVersion == directoryv1alpha1.GroupVersion.String() && owner.UID == ds.GetUID() {
				return r.Client.Delete(ctx, &deployment, client.PropagationPolicy("Background"))
			}
			return nil
		}
		return err
	}

	deployment.Name = proxyName
	deployment.Namespace = ds.Namespace

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &deployment, func() error {
		log.V(8).Info("CreateorUpdate deployment", "deployment", deployment)

		var err error
		// does the deployment not exist yet?
		if deployment.CreationTimestamp.IsZero() {
			err = createDSProxyDeployment(ds, &deployment)
			_ = controllerutil.SetControllerReference(ds, &deployment, r.Scheme)
			log.V(8).Info("Created New deployment from template", "deployment", deployment)
		} else {
			// If the deployment exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			err = updateDSProxyDeployment(ds, &deployment)
		}

		log.V(8).Info("deployment after update/create", "deployment", deployment)
		return err

	})
	if err != nil {
		return errors.Wrap(err, "unable to CreateOrUpdate deployment")
	}
	return nil
}

// This function updates an existing deployment to match settings in the custom resource
func updateDSProxyDeployment(ds *directoryv1alpha1.DirectoryService, deployment *apps.Deployment) error {

	// Copy our expected replicas to the deployment
	deployment.Spec.Replicas = &ds.Spec.Proxy.Replicas

	// copy the current deployment replicas up the ds status
	ds.Status.ProxyStatus.ReadyReplicas = deployment.Status.ReadyReplicas
	ds.Status.ProxyStatus.Replicas = deployment.Status.Replicas

	// Update the image
	deployment.Spec.Template.Spec.Containers[0].Image = ds.Spec.Proxy.Image
	deployment.Spec.Template.Spec.InitContainers[0].Image = ds.Spec.Proxy.Image

	return nil
}

// https://godoc.org/k8s.io/api/apps/v1#DeploymentSpec
func createDSProxyDeployment(ds *directoryv1alpha1.DirectoryService, deployment *apps.Deployment) error {

	// TODO: What is the canonical go way of using these contants in a template. Go wants a pointer to these
	// not a constant
	var fsGroup int64 = 0
	var forgerockUser int64 = 11111

	proxyName := ds.Name + "-proxy"
	labels := createLabels(proxyName, map[string]string{
		"app.kubernetes.io/component": "ds-proxy",
	})

	// Create a template
	stemplate := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      proxyName,
			Namespace: ds.Namespace,
		},
		Spec: apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":     labels["app.kubernetes.io/name"],
					"app.kubernetes.io/instance": labels["app.kubernetes.io/instance"],
				},
			},
			Replicas: &ds.Spec.Proxy.Replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					// We use anti affinity to spread the pods out over host node
					Affinity: &v1.Affinity{
						NodeAffinity: &v1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
								{
									Weight: 100,
									Preference: v1.NodeSelectorTerm{
										MatchExpressions: []v1.NodeSelectorRequirement{
											{
												Key:      "ds",
												Operator: "Exists",
											},
										},
									},
								},
							},
						},
						// PodAffinity:     &v1.PodAffinity{},
						PodAntiAffinity: &v1.PodAntiAffinity{
							//RequiredDuringSchedulingIgnoredDuringExecution:  []v1.PodAffinityTerm{},
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: v1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{},
											MatchExpressions: []metav1.LabelSelectorRequirement{
												{
													Key:      "app.kubernetes.io/component",
													Operator: "In",
													Values:   []string{labels["app.kubernetes.io/component"]},
												},
											},
										},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						},
					},
					// Tolerate any nodes tainted with kubectl taint nodes node1 key=directory:NoSchedule
					// This has no effect if the user does not wish to taint any nodes.
					Tolerations: []v1.Toleration{
						{
							Key:      "WorkerDedicatedDS",
							Operator: "Exists",
						},
					},
					InitContainers: []v1.Container{
						{
							Name:            "initialize",
							Image:           ds.Spec.Proxy.Image,
							ImagePullPolicy: v1.PullIfNotPresent,
							Command:         []string{"/opt/opendj/scripts/init-and-restore.sh"},
							// Args: []string{"initialize-only"},
							VolumeMounts: []v1.VolumeMount{
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
							},
							Resources: ds.DeepCopy().Spec.Proxy.Resources,
							Env: []v1.EnvVar{
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
								{
									Name:  "DSPROXY_BOOTSTRAP_REPLICATION_SERVERS",
									Value: fmt.Sprintf("%s-0.%s:4444", ds.Name, ds.Name),
								},
								{
									Name:  "DSPROXY_PRIMARY_GROUP_ID",
									Value: ds.Spec.Proxy.PrimaryGroupID,
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:            "ds-proxy",
							Image:           ds.Spec.Proxy.Image,
							ImagePullPolicy: v1.PullIfNotPresent,
							Args:            []string{"start-ds"},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "data",
									MountPath: DSDataPath,
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
							},
							Resources: ds.DeepCopy().Spec.Proxy.Resources,
							Env: []v1.EnvVar{
								{
									Name:  "DSPROXY_BOOTSTRAP_REPLICATION_SERVERS",
									Value: fmt.Sprintf("%s-0.%s:4444", ds.Name, ds.Name),
								},
								{
									Name:  "DSPROXY_PRIMARY_GROUP_ID",
									Value: ds.Spec.Proxy.PrimaryGroupID,
								},
							},
						},
					},
					SecurityContext: &v1.PodSecurityContext{
						FSGroup:   &fsGroup,
						RunAsUser: &forgerockUser,
					},
					Volumes: []v1.Volume{
						{
							Name: "ds-master-keypair", // master keypair for encryption
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: ds.Spec.PodTemplate.Certificates.MasterSecretName,
								},
							},
						},
						{
							Name: "ds-ssl-keypair", // ssl between instances
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: ds.Spec.PodTemplate.Certificates.SSLSecretName,
								},
							},
						},
						{
							Name: "truststore", // truststore
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: ds.Spec.PodTemplate.Certificates.TruststoreSecretName,
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
							Name: "data",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
	stemplate.DeepCopyInto(deployment)
	return nil
}
